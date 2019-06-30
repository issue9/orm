// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package model 管理数据模型
package model

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/issue9/orm/v2/fetch"
	"github.com/issue9/orm/v2/internal/tags"
)

const (
	defaultAINameSuffix = "_ai"
	defaultPKNameSuffix = "_pk"
)

// Model 表示一个数据库的表模型。数据结构从字段和字段的 struct tag 中分析得出。
type Model struct {
	Name    string              // 表的名称
	Columns []*Column           // 所有的列
	OCC     *Column             // 乐观锁
	Meta    map[string][]string // 表级别的数据，如存储引擎，表名和字符集等。

	// 索引内容
	//
	// 目前不支持唯一索引，如果需要唯一索引，可以设置成唯一约束。
	Indexes map[string][]*Column

	// 唯一约束
	//
	// 键名为约束名，键值为该约束关联的列
	Uniques map[string][]*Column

	// Check 约束
	//
	// 键名为约束名，键值为约束表达式
	Checks map[string]string

	FK []*ForeignKey

	// 自增约束
	//
	// AI 为自增约束的列。
	// AIName 为自增约束的名称，部分数据库需要，如果不指定，会采用 表名 + _ai 的格式
	AI     *Column
	AIName string

	// 主键约束
	//
	// PK 为主键约束的列列表。
	// PKName 为主键约束的名称，如果未指定，会采用 表名 + _ai 的格式
	PK     []*Column
	PKName string
}

func propertyError(field, name, message string) error {
	return fmt.Errorf("%s 的 %s 属性发生以下错误: %s", field, name, message)
}

// New 从一个 obj 声明一个 Model 实例。
// obj 可以是一个 struct 实例或是指针。
func (ms *Models) New(obj interface{}) (*Model, error) {
	rval := reflect.ValueOf(obj)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}
	rtype := rval.Type()

	if rtype.Kind() != reflect.Struct {
		return nil, fetch.ErrInvalidKind
	}

	ms.locker.Lock()
	defer ms.locker.Unlock()

	if m, found := ms.items[rtype]; found {
		return m, nil
	}

	m := &Model{
		Columns: make([]*Column, 0, rtype.NumField()),
		Indexes: map[string][]*Column{},
		Uniques: map[string][]*Column{},
		Name:    rtype.Name(),
		FK:      []*ForeignKey{},
		Checks:  map[string]string{},
		Meta:    map[string][]string{},
	}

	// NOTE: 需要保证表名的获取在 parseColumns 之前
	// 诸如 AIName、PKName 等依赖表名字段。
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

	if err := ms.addModel(rtype, m); err != nil {
		return nil, err
	}

	return m, nil
}

func (ms *Models) addModel(gotype reflect.Type, m *Model) error {
	ms.items[gotype] = m

	for name := range m.Indexes {
		if err := ms.addNames(name); err != nil {
			return err
		}
	}

	for name := range m.Uniques {
		if err := ms.addNames(name); err != nil {
			return err
		}
	}

	for name := range m.Checks {
		if err := ms.addNames(name); err != nil {
			return err
		}
	}

	for _, fk := range m.FK {
		if err := ms.addNames(fk.Name); err != nil {
			return err
		}
	}

	if m.AIName != "" {
		if err := ms.addNames(m.AIName); err != nil {
			return err
		}
	}

	if m.PKName != "" {
		if err := ms.addNames(m.PKName); err != nil {
			return err
		}
	}

	return nil
}

// 对整个对象做一次检测，查看是否合法
// 必须要在 Model 初始化完成之后调用。
func (m *Model) check() error {
	if m.AI != nil {
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

	for _, c := range m.Columns {
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
			if err := m.parseColumns(rval.Field(i)); err != nil {
				return err
			}
			continue
		}

		if unicode.IsLower(rune(field.Name[0])) { // 忽略以小写字母开头的字段
			continue
		}

		tag := field.Tag.Get(fetch.Tag)
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
		m.Columns = append(m.Columns, col)
		return nil
	}

	ts := tags.Parse(tag)
	for _, tag := range ts {
		switch tag.Name {
		case "name": // name(col)
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
	m.Columns = append(m.Columns, col)

	return nil
}

// 分析 meta 接口数据。
func (m *Model) parseMeta(tag string) error {
	ts := tags.Parse(tag)
	if len(ts) == 0 {
		return nil
	}

	for _, v := range ts {
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

			if _, found := m.Checks[v.Args[0]]; found {
				return propertyError("Metaer", "check", "已经存在相同名称的 check 约束")
			}

			name := strings.ToLower(v.Args[0])
			m.Checks[name] = v.Args[1]
		default:
			m.Meta[v.Name] = v.Args
		}
	}

	return nil
}

// occ(true) or occ
func (m *Model) setOCC(c *Column, vals []string) error {
	if c.AI || c.Nullable {
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

	name := strings.ToLower(vals[0])
	m.Indexes[name] = append(m.Indexes[name], col)
	return nil
}

// pk
func (m *Model) setPK(col *Column, vals []string) error {
	if len(m.PK) > 0 {
		return propertyError(col.Name, "pk", "已经存在自增约束，不能再指定主键约束")
	}

	if len(vals) != 0 {
		return propertyError(col.Name, "pk", "太多的值")
	}

	m.PK = append(m.PK, col)
	m.PKName = m.Name + defaultPKNameSuffix
	return nil
}

// unique(unique_name)
func (m *Model) setUnique(col *Column, vals []string) error {
	if len(vals) != 1 {
		return propertyError(col.Name, "unique", "只能带一个参数")
	}

	name := strings.ToLower(vals[0])
	m.Uniques[name] = append(m.Uniques[name], col)

	return nil
}

// fk(fk_name,refTable,refColName,updateRule,deleteRule)
func (m *Model) setFK(col *Column, vals []string) error {
	if len(vals) < 3 {
		return propertyError(col.Name, "fk", "参数不够")
	}

	fkInst := &ForeignKey{
		Name:         strings.ToLower(vals[0]),
		Column:       col,
		RefTableName: vals[1],
		RefColName:   vals[2],
	}

	if len(vals) > 3 { // 存在 updateRule
		fkInst.UpdateRule = vals[3]
	}
	if len(vals) > 4 { // 存在 deleteRule
		fkInst.DeleteRule = vals[4]
	}

	m.FK = append(m.FK, fkInst)
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
	m.AIName = m.Name + defaultAINameSuffix
	col.AI = true

	return nil
}

// FindColumn 查找指定名称的列
//
// 不存在该列则返回 nil
func (m *Model) FindColumn(name string) *Column {
	for _, col := range m.Columns {
		if col.Name == name {
			return col
		}
	}
	return nil
}
