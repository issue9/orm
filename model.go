// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/issue9/orm/fetch"
	"github.com/issue9/orm/internal/tags"
)

// 预定的约束类型，方便 Model 中使用。
const (
	none conType = iota
	index
	unique
	fk
	check
)

// model 缓存
var models = &struct {
	sync.Mutex
	items map[reflect.Type]*Model
}{
	items: map[reflect.Type]*Model{},
}

type conType int8

// Model 表示一个数据库的表模型。数据结构从字段和字段的 struct tag 中分析得出。
type Model struct {
	Name          string                 // 表的名称
	Cols          map[string]*Column     // 所有的列
	KeyIndexes    map[string][]*Column   // 索引列
	UniqueIndexes map[string][]*Column   // 唯一索引列
	FK            map[string]*ForeignKey // 外键
	PK            []*Column              // 主键
	AI            *Column                // 自增列
	OCC           *Column                // 乐观锁
	Check         map[string]string      // Check 键名为约束名，键值为约束表达式
	Meta          map[string][]string    // 表级别的数据，如存储引擎，表名和字符集等。

	constraints map[string]conType // 约束名缓存
}

func propertyError(field, name, message string) error {
	return fmt.Errorf("%s 的 %s 属性发生以下错误: %s", field, name, message)
}

func (t conType) String() string {
	switch t {
	case none:
		return "<none>"
	case index:
		return "KEY INDEX"
	case unique:
		return "UNIQUE INDEX"
	case fk:
		return "FOREIGN KEY"
	case check:
		return "CHECK"
	default:
		return "<unknown>"
	}
}

// NewModel 从一个 obj 声明一个 Model 实例。
// obj 可以是一个 struct 实例或是指针。
func NewModel(obj interface{}) (*Model, error) {
	rval := reflect.ValueOf(obj)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}
	rtype := rval.Type()

	if rtype.Kind() != reflect.Struct {
		return nil, fetch.ErrInvalidKind
	}

	models.Lock()
	defer models.Unlock()

	if m, found := models.items[rtype]; found {
		return m, nil
	}

	m := &Model{
		Cols:          map[string]*Column{},
		KeyIndexes:    map[string][]*Column{},
		UniqueIndexes: map[string][]*Column{},
		Name:          rtype.Name(),
		FK:            map[string]*ForeignKey{},
		Check:         map[string]string{},
		Meta:          map[string][]string{},
		constraints:   map[string]conType{},
	}

	if err := m.parseColumns(rval); err != nil {
		return nil, err
	}

	if meta, ok := obj.(Metaer); ok {
		if err := m.parseMeta(meta.Meta()); err != nil {
			return nil, err
		}
	}

	if err := m.check(); err != nil {
		return nil, err
	}

	models.items[rtype] = m
	return m, nil
}

// 对整个对象做一次检测，查看是否合法
// 必须要在 Model 初始化完成之后调用。
func (m *Model) check() error {
	if m.AI != nil {
		if len(m.PK) > 1 { // 肯定不是 AI 列
			return propertyError(m.AI.Name, "pk", "不能与自增列并存")
		}

		if len(m.PK) == 1 && m.PK[0] != m.AI {
			return propertyError(m.AI.Name, "pk", "不能与自增列并存")
		}

		if m.AI.Nullable {
			return propertyError(m.AI.Name, "nullable", "不能与自增列并存")
		}

		if m.AI.HasDefault {
			return propertyError(m.AI.Name, "default", "不能与自增列并存")
		}
	}

	if len(m.PK) == 1 && m.PK[0].HasDefault {
		return propertyError(m.PK[0].Name, "default", "不能为单一主键")
	}

	for _, c := range m.Cols {
		if err := c.checkLen(); err != nil {
			return err
		}
	}

	return nil
}

// 将 rval 中的结构解析到 m 中。支持匿名字段
func (m *Model) parseColumns(rval reflect.Value) error {
	rtype := rval.Type()
	num := rtype.NumField()
	for i := 0; i < num; i++ {
		field := rtype.Field(i)

		if field.Anonymous {
			m.parseColumns(rval.Field(i))
			continue
		}

		if unicode.IsLower(rune(field.Name[0])) { // 忽略以小写字母开头的字段
			continue
		}

		tag := field.Tag.Get("orm")
		if tag == "-" {
			continue
		}

		col := m.newColumn(field)
		if err := m.parseColumn(col, tag); err != nil {
			return err
		}
	}

	return nil
}

// 分析一个字段。
func (m *Model) parseColumn(col *Column, tag string) (err error) {
	if len(tag) == 0 { // 没有附加的 struct tag，直接取得几个关键信息返回。
		m.Cols[col.Name] = col
		return nil
	}

	tags := tags.Parse(tag)
	for _, tag := range tags {
		switch tag.Name {
		case "name": // name(colname)
			if len(tag.Args) != 1 {
				return propertyError(col.Name, "name", "过多的参数值")
			}
			col.Name = tag.Args[0]
		case "index":
			err = m.setIndex(col, tag.Args)
		case "pk":
			err = m.setPK(col, tag.Args)
		case "unique":
			err = m.setUnique(col, tag.Args)
		case "nullable":
			err = col.setNullable(tag.Args)
		case "ai":
			err = m.setAI(col, tag.Args)
		case "len":
			err = col.setLen(tag.Args)
		case "fk":
			err = m.setFK(col, tag.Args)
		case "default":
			err = m.setDefault(col, tag.Args)
		case "occ":
			err = m.setOCC(col, tag.Args)
		default:
			err = propertyError(col.Name, tag.Name, "未知的属性")
		}

		if err != nil {
			return err
		}
	}

	// col.Name 可能在上面的 for 循环中被更改，所以要在最后再添加到 m.Cols 中
	m.Cols[col.Name] = col

	return nil
}

// 分析 meta 接口数据。
func (m *Model) parseMeta(tag string) error {
	tags := tags.Parse(tag)
	if len(tags) == 0 {
		return nil
	}

	for _, v := range tags {
		switch v.Name {
		case "name":
			if len(v.Args) != 1 {
				return propertyError("Metaer", "name", "太多的值")
			}

			m.Name = v.Args[0]
		case "check":
			if len(v.Args) != 2 {
				return propertyError("Metaer", "check", "参数个数不正确")
			}

			if _, found := m.Check[v.Args[0]]; found {
				return propertyError("Metaer", "check", "已经存在相同名称的 check 约束")
			}

			if typ := m.hasConstraint(v.Args[0], check); typ != none {
				return propertyError("Metaer", "check", "与其它约束名称相同")
			}

			name := strings.ToLower(v.Args[0])
			m.constraints[name] = check
			m.Check[name] = v.Args[1]
		default:
			m.Meta[v.Name] = v.Args
		}
	}

	return nil
}

// occ(true) or occ
func (m *Model) setOCC(c *Column, vals []string) error {
	if c.IsAI() || c.Nullable {
		return propertyError(c.Name, "occ", "自增列和允许为空的列不能作为乐观锁列")
	}

	if m.OCC != nil {
		return propertyError(c.Name, "occ", "已经指定了一个乐观锁")
	}

	switch c.GoType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	default:
		return propertyError(c.Name, "occ", "值只能是数值")
	}

	switch len(vals) {
	case 0:
		m.OCC = c
	case 1:
		val, err := strconv.ParseBool(vals[0])
		if err != nil {
			return err
		}
		if val {
			m.OCC = c
		}
	default:
		return propertyError(c.Name, "occ", "指定了太多的值")
	}

	return nil
}

// default(5)
func (m *Model) setDefault(col *Column, vals []string) error {
	if len(vals) != 1 {
		return propertyError(col.Name, "default", "太多的值")
	}

	col.HasDefault = true
	col.Default = vals[0]

	return nil
}

// index(idx_name)
func (m *Model) setIndex(col *Column, vals []string) error {
	if len(vals) != 1 {
		return propertyError(col.Name, "index", "太多的值")
	}

	if typ := m.hasConstraint(vals[0], index); typ != none {
		return propertyError(col.Name, "index", "已经存在相同的约束名")
	}

	name := strings.ToLower(vals[0])
	m.constraints[name] = index
	m.KeyIndexes[name] = append(m.KeyIndexes[name], col)
	return nil
}

// pk
func (m *Model) setPK(col *Column, vals []string) error {
	if len(vals) != 0 {
		return propertyError(col.Name, "pk", "太多的值")
	}

	m.PK = append(m.PK, col)
	return nil
}

// unique(unique_name)
func (m *Model) setUnique(col *Column, vals []string) error {
	if len(vals) != 1 {
		return propertyError(col.Name, "unique", "只能带一个参数")
	}

	if typ := m.hasConstraint(vals[0], unique); typ != none {
		return propertyError(col.Name, "unique", "已经存在相同的约束名")
	}

	name := strings.ToLower(vals[0])
	m.constraints[name] = unique
	m.UniqueIndexes[name] = append(m.UniqueIndexes[name], col)

	return nil
}

// fk(fk_name,refTable,refColName,updateRule,deleteRule)
func (m *Model) setFK(col *Column, vals []string) error {
	if len(vals) < 3 {
		return propertyError(col.Name, "fk", "参数不够")
	}

	if typ := m.hasConstraint(vals[0], fk); typ != none {
		return propertyError(col.Name, "fk", "已经存在相同的约束名")
	}

	if _, found := m.FK[vals[0]]; found {
		return propertyError(col.Name, "fk", "重复的外键约束名")
	}

	fkInst := &ForeignKey{
		Col:          col,
		RefTableName: vals[1],
		RefColName:   vals[2],
	}

	if len(vals) > 3 { // 存在updateRule
		fkInst.UpdateRule = vals[3]
	}
	if len(vals) > 4 { // 存在deleteRule
		fkInst.DeleteRule = vals[4]
	}

	name := strings.ToLower(vals[0])
	m.constraints[name] = fk
	m.FK[name] = fkInst
	return nil
}

// ai
func (m *Model) setAI(col *Column, vals []string) (err error) {
	if len(vals) != 0 {
		return propertyError(col.Name, "ai", "太多的值")
	}

	switch col.GoType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	default:
		return propertyError(col.Name, "ai", "类型只能是数值")
	}

	m.AI = col

	// 去掉其它主键，将自增列设置为主键
	m.PK = append(m.PK[:0], col)
	return nil
}

// 是否存在指定名称的约束名，name 不区分大小写。
// 若已经存在返回表示该约束类型的常量，否则返回 none。
//
// NOTE:约束名不区分大小写
func (m *Model) hasConstraint(name string, except conType) conType {
	if typ, found := m.constraints[strings.ToLower(name)]; found && typ != except {
		return typ
	}

	return none
}

// ClearModels 清除所有的 Model 缓存。
func ClearModels() {
	models.Lock()
	defer models.Unlock()

	models.items = map[reflect.Type]*Model{}
}
