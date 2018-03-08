// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package core

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/issue9/orm/internal/tags"
)

// ErrInvalidKind 表示该类型不是结构体或是结构体的指针
var ErrInvalidKind = errors.New("不支持的 reflect.Kind()，只能是结构体或是结构体指针")

type conType int

// 预定的约束类型，方便 Model 中使用。
const (
	none conType = iota
	index
	unique
	fk
	check
)

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

// model 缓存
var models = &modelsMap{items: map[reflect.Type]*Model{}}

type modelsMap struct {
	sync.Mutex
	items map[reflect.Type]*Model
}

// Metaer 用于指定一个表级别的元数据。如表名，存储引擎等：
//  "name(tbl_name);engine(myISAM);charset(utf-8)"
type Metaer interface {
	Meta() string
}

// Model 表示一个数据库的表模型。数据结构从字段和字段的 struct tag 中分析得出。
type Model struct {
	Name string // 表的名称

	Cols          map[string]*Column     // 所有的列
	KeyIndexes    map[string][]*Column   // 索引列
	UniqueIndexes map[string][]*Column   // 唯一索引列
	FK            map[string]*ForeignKey // 外键
	PK            []*Column              // 主键
	AI            *Column                // 自增列
	OCC           *Column                // 锁观锁
	Check         map[string]string      // Check 键名为约束名，键值为约束表达式
	Meta          map[string][]string    // 表级别的数据，如存储引擎，表名和字符集等。

	constraints map[string]conType // 约束名缓存
}

// ForeignKey 外键
type ForeignKey struct {
	Col                      *Column
	RefTableName, RefColName string
	UpdateRule, DeleteRule   string
}

// Column 列结构
type Column struct {
	model *Model

	Name     string       // 数据库的字段名
	Len1     int          // 长度1，仅对部分类型启作用
	Len2     int          // 长度2，仅对部分类型启作用
	Nullable bool         // 是否可以为 NULL
	GoType   reflect.Type // Go 语言中的数据类型
	Zero     interface{}  // GoType 的零值
	GoName   string       // 结构字段名

	HasDefault bool
	Default    string // 默认值
}

// IsAI 当前列是否为自增列
func (c *Column) IsAI() bool {
	return (c.model != nil) && (c.model.AI == c)
}

// 从参数中获取 Column 的 len1 和 len2 变量。
// len(len1,len2)
func (c *Column) setLen(vals []string) (err error) {
	switch len(vals) {
	case 0:
	case 1:
		c.Len1, err = strconv.Atoi(vals[0])
	case 2:
		c.Len1, err = strconv.Atoi(vals[0])
		c.Len2, err = strconv.Atoi(vals[1])
	default:
		err = fmt.Errorf("setLen:[%v]字段的len属性指定了过多的参数:[%v]", c.Name, vals)
	}

	return
}

// 从 vals 中分析，得出 Column.Nullable 的值。
// nullable; or nullable(true);
func (c *Column) setNullable(vals []string) (err error) {
	if c.IsAI() {
		return fmt.Errorf("setNullable:自增列[%v]不能为nullable", c.Name)
	}

	switch len(vals) {
	case 0:
		c.Nullable = true
	case 1:
		if c.Nullable, err = strconv.ParseBool(vals[0]); err != nil {
			return err
		}
	default:
		return fmt.Errorf("setNullable:[%v]字段的nullable属性指定了太多的值:[%v]", c.Name, vals)
	}

	return nil
}

// NewModel 从一个 obj 声明一个 Model 实例。
// obj 可以是一个 struct 实例或是指针。
func NewModel(obj interface{}) (*Model, error) {
	models.Lock()
	defer models.Unlock()

	rval := reflect.ValueOf(obj)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}
	rtype := rval.Type()

	if rtype.Kind() != reflect.Struct {
		return nil, ErrInvalidKind
	}

	// 是否已经缓存的数组
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

	if err := m.parseMeta(obj); err != nil {
		return nil, err
	}

	models.items[rtype] = m
	return m, nil
}

// 将rval中的结构解析到m中。支持匿名字段
func (m *Model) parseColumns(rval reflect.Value) error {
	rtype := rval.Type()
	num := rtype.NumField()
	for i := 0; i < num; i++ {
		field := rtype.Field(i)

		if field.Anonymous {
			m.parseColumns(rval.Field(i))
			continue
		}

		if err := m.parseColumn(field); err != nil {
			return err
		}
	}

	return nil
}

func (m *Model) newColumn(field reflect.StructField) *Column {
	return &Column{
		GoType: field.Type,
		Zero:   reflect.Zero(field.Type).Interface(),
		Name:   field.Name,
		model:  m,
		GoName: field.Name,
	}
}

// 分析一个字段。
func (m *Model) parseColumn(field reflect.StructField) (err error) {
	if unicode.IsLower(rune(field.Name[0])) { // 忽略以小写字母开头的字段
		return nil
	}

	tagTxt := field.Tag.Get("orm")
	if tagTxt == "-" {
		return nil
	}

	col := m.newColumn(field)

	if len(tagTxt) == 0 { // 没有附加的 struct tag，直接取得几个关键信息返回。
		m.Cols[col.Name] = col
		return nil
	}

	tags := tags.Parse(tagTxt)
	for k, v := range tags {
		switch k {
		case "name": // name(colname)
			if len(v) != 1 {
				return fmt.Errorf("parseColumn:name属性指定了太多的参数：[%v]", v)
			}
			col.Name = v[0]
		case "index":
			err = m.setIndex(col, v)
		case "pk":
			err = m.setPK(col, v)
		case "unique":
			err = m.setUnique(col, v)
		case "nullable":
			err = col.setNullable(v)
		case "ai":
			err = m.setAI(col, v)
		case "len":
			err = col.setLen(v)
		case "fk":
			err = m.setFK(col, v)
		case "default":
			err = m.setDefault(col, v)
		case "occ":
			err = m.setOCC(col, v)
		default:
			err = fmt.Errorf("parseColumn:未知的struct tag属性:[%v]", k)
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
func (m *Model) parseMeta(obj interface{}) error {
	meta, ok := obj.(Metaer)
	if !ok {
		return nil
	}

	tags := tags.Parse(meta.Meta())
	if len(tags) == 0 {
		return nil
	}

	for k, v := range tags {
		switch k {
		case "name":
			if len(v) != 1 {
				return fmt.Errorf("parseMeta:Meta接口的name属性指定了太多参数：[%v]", v)
			}

			m.Name = v[0]
		case "check":
			if len(v) != 2 {
				return fmt.Errorf("parseMeta:Meta接口的check属性的参数只能为2个，当前值为:[%v]", v)
			}

			if _, found := m.Check[v[0]]; found {
				return fmt.Errorf("parseMeta:已经存在相同名称[%v]的Check约束", v[0])
			}

			if typ := m.hasConstraint(v[0], check); typ != none {
				return fmt.Errorf("parseMeta:已经存在相同的约束名[%v]，位于[%v]中", v[0], typ)
			}

			m.constraints[v[0]] = check
			m.Check[v[0]] = v[1]
		default:
			m.Meta[k] = v
		}
	}

	return nil
}

// occ(true) or occ
func (m *Model) setOCC(c *Column, vals []string) error {
	if c.IsAI() || c.Nullable {
		return errors.New("自增列和允许为空的列不能作为乐观锁列")
	}

	if m.OCC != nil {
		return errors.New("已经存在一个乐观锁")
	}

	switch c.GoType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	default:
		return errors.New("乐观锁只能是数值类型")
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
		return fmt.Errorf("[%v]字段的 occ 属性指定了太多的值:[%v]", c.Name, vals)
	}

	return nil
}

// default(5)
func (m *Model) setDefault(col *Column, vals []string) error {
	if m.AI == col {
		return errors.New("setDefault:自增列不能设置默认值")
	}

	for _, c := range m.PK {
		if c == col {
			return errors.New("setDefault:不能为主键设置默认值")
		}
	}

	if len(vals) != 1 {
		return fmt.Errorf("setDefault:[%v]字段的default属性指定了太多的参数：[%v]", col.Name, vals)
	}

	col.HasDefault = true
	col.Default = vals[0]

	return nil
}

// index(idx_name)
func (m *Model) setIndex(col *Column, vals []string) error {
	if len(vals) != 1 {
		return fmt.Errorf("setIndex:[%v]字段的index属性指定了太多的参数:[%v]", col.Name, vals)
	}

	if typ := m.hasConstraint(vals[0], index); typ != none {
		return fmt.Errorf("setIndex:已经存在相同的约束名[%v]，位于[%v]中", vals[0], typ)
	}

	m.constraints[vals[0]] = index
	m.KeyIndexes[vals[0]] = append(m.KeyIndexes[vals[0]], col)
	return nil
}

// pk
func (m *Model) setPK(col *Column, vals []string) error {
	if col.HasDefault {
		return fmt.Errorf("setPK:不能将一个含有默认值的列[%v]设置为主键", col.Name)
	}

	if len(vals) != 0 {
		return fmt.Errorf("setPK:[%v]字段的pk属性指定了太多的参数:[%v]", col.Name, vals)
	}

	if m.AI != nil {
		return fmt.Errorf("setPK:已经存在自增列，不需要再次指定主键")
	}

	m.PK = append(m.PK, col)
	return nil
}

// unique(unique_name)
func (m *Model) setUnique(col *Column, vals []string) error {
	if len(vals) != 1 {
		return fmt.Errorf("setUnique:[%v]字段的unique属性只能带一个参数:[%v]", col.Name, vals)
	}

	if typ := m.hasConstraint(vals[0], unique); typ != none {
		return fmt.Errorf("setUnique:已经存在相同的约束名[%v]，位于[%v]中", vals[0], typ)
	}

	m.constraints[vals[0]] = unique
	m.UniqueIndexes[vals[0]] = append(m.UniqueIndexes[vals[0]], col)

	return nil
}

// fk(fk_name,refTable,refColName,updateRule,deleteRule)
func (m *Model) setFK(col *Column, vals []string) error {
	if len(vals) < 3 {
		return errors.New("setFK:fk参数必须大于3个")
	}

	if typ := m.hasConstraint(vals[0], fk); typ != none {
		return fmt.Errorf("setFK:已经存在相同的约束名[%v]，位于[%v]中", vals[0], typ)
	}

	if _, found := m.FK[vals[0]]; found {
		return fmt.Errorf("setFK:重复的外键约束名:[%v]", vals[0])
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

	m.constraints[vals[0]] = fk
	m.FK[vals[0]] = fkInst
	return nil
}

// ai(colName,start,step)
func (m *Model) setAI(col *Column, vals []string) (err error) {
	if col.HasDefault {
		return fmt.Errorf("setAI:不能将一个含有默认值的列[%v]设置为自增", col.Name)
	}

	if len(vals) != 0 {
		return fmt.Errorf("setAI:[%v]字段的ai属性指定了太多的参数:[%v]", col.Name, vals)
	}

	if col.Nullable {
		return fmt.Errorf("setAI:nullable列不能为自增列[%v]", col.Name)
	}

	switch col.GoType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	default:
		return errors.New("setAI:自增列只能是整数类型")
	}

	m.AI = col

	// 去掉其它主键，将自增列设置为主键
	m.PK = append(m.PK[:0], col)
	return nil
}

// 是否存在指定名称的约束名，name 不区分大小写。
// 若已经存在返回表示该约束类型的常量，否则返回 none。
func (m *Model) hasConstraint(name string, except conType) conType {
	// 约束名不区分大小写
	if typ, found := m.constraints[strings.ToLower(name)]; found && typ != except {
		return typ
	}

	return none
}

// ClearModels 清除所有的Model缓存。
func ClearModels() {
	models.Lock()
	defer models.Unlock()

	models.items = map[reflect.Type]*Model{}
}
