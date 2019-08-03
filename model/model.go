// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package model 管理数据模型
package model

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/issue9/orm/v2/core"
	"github.com/issue9/orm/v2/fetch"
	"github.com/issue9/orm/v2/internal/tags"
)

func propertyError(field, name, message string) error {
	return fmt.Errorf("%s 的 %s 属性发生以下错误: %s", field, name, message)
}

// New 从一个 obj 声明一个 Model 实例。
// obj 可以是一个 struct 实例或是指针。
func (ms *Models) New(obj interface{}) (*core.Model, error) {
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

	m := core.NewModel(core.Table, "#"+rtype.Name(), rtype.NumField())

	if err := parseColumns(m, rval); err != nil {
		return nil, err
	}

	if meta, ok := obj.(Metaer); ok {
		if err := parseMeta(m, meta.Meta()); err != nil {
			return nil, err
		}
	}

	if view, ok := obj.(Viewer); ok {
		m.Type = core.View
		sql, err := view.ViewAs(ms.engine)
		if err != nil {
			return nil, err
		}
		m.ViewAs = sql
	}

	if err := m.Sanitize(); err != nil {
		return nil, err
	}

	if err := ms.addModel(rtype, m); err != nil {
		return nil, err
	}

	return m, nil
}

func (ms *Models) addModel(goType reflect.Type, m *core.Model) error {
	ms.items[goType] = m

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

	for name := range m.ForeignKeys {
		if err := ms.addNames(name); err != nil {
			return err
		}
	}

	if m.AutoIncrement != nil {
		if err := ms.addNames(m.AIName()); err != nil {
			return err
		}
	}

	if len(m.PrimaryKey) > 0 {
		if err := ms.addNames(m.PKName()); err != nil {
			return err
		}
	}

	return nil
}

// 将 rval 中的结构解析到 m 中。支持匿名字段
func parseColumns(m *core.Model, rval reflect.Value) error {
	rtype := rval.Type()
	num := rtype.NumField()
	for i := 0; i < num; i++ {
		field := rtype.Field(i)

		if field.Anonymous {
			if err := parseColumns(m, rval.Field(i)); err != nil {
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

		col, err := newColumn(field)
		if err != nil {
			return err
		}

		if err := parseColumn(m, col, tag); err != nil {
			return err
		}
	}

	return nil
}

// 分析一个字段。
func parseColumn(m *core.Model, col *core.Column, tag string) (err error) {
	if err = m.AddColumn(col); err != nil {
		return err
	}

	if len(tag) == 0 { // 没有附加的 struct tag，直接取得几个关键信息返回。
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
			err = setIndex(m, col, tag.Args)
		case "pk":
			err = setPK(m, col, tag.Args)
		case "unique":
			err = setUnique(m, col, tag.Args)
		case "nullable":
			err = setColumnNullable(col, tag.Args)
		case "ai":
			err = setAI(m, col, tag.Args)
		case "len":
			err = setColumnLen(col, tag.Args)
		case "fk":
			err = setFK(m, col, tag.Args)
		case "default":
			err = setDefault(col, tag.Args)
		case "occ":
			err = setOCC(m, col, tag.Args)
		default:
			err = propertyError(col.Name, tag.Name, "未知的属性")
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// 分析 meta 接口数据。
func parseMeta(m *core.Model, tag string) error {
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

			m.SetName("#" + v.Args[0]) // 所有 model 生成的表都带 #
		case "check":
			if len(v.Args) != 2 {
				return propertyError("Metaer", "check", "参数个数不正确")
			}

			if _, found := m.Checks[v.Args[0]]; found {
				return propertyError("Metaer", "check", "已经存在相同名称的 check 约束")
			}

			if err := m.NewCheck(strings.ToLower(v.Args[0]), v.Args[1]); err != nil {
				return err
			}
		default:
			m.Meta[v.Name] = v.Args
		}
	}

	return nil
}

// occ
func setOCC(m *core.Model, c *core.Column, vals []string) error {
	if len(vals) > 0 {
		return propertyError(c.Name, "occ", "指定了太多的值")
	}

	return m.SetOCC(c)
}

// index(idx_name)
func setIndex(m *core.Model, col *core.Column, vals []string) error {
	if len(vals) != 1 {
		return propertyError(col.Name, "index", "太多的值")
	}

	name := strings.ToLower(vals[0])
	return m.AddIndex(core.IndexDefault, name, col)
}

// pk
func setPK(m *core.Model, col *core.Column, vals []string) error {
	if len(vals) != 0 {
		return propertyError(col.Name, "pk", "太多的值")
	}

	return m.AddPrimaryKey(col)
}

// unique(unique_name)
func setUnique(m *core.Model, col *core.Column, vals []string) error {
	if len(vals) != 1 {
		return propertyError(col.Name, "unique", "只能带一个参数")
	}

	name := strings.ToLower(vals[0])
	return m.AddUnique(name, col)
}

// fk(fk_name,refTable,refColName,updateRule,deleteRule)
func setFK(m *core.Model, col *core.Column, vals []string) error {
	if len(vals) < 3 {
		return propertyError(col.Name, "fk", "参数不够")
	}

	fk := &core.ForeignKey{
		Column:       col,
		RefTableName: vals[1],
		RefColName:   vals[2],
	}

	if len(vals) > 3 { // 存在 updateRule
		fk.UpdateRule = vals[3]
	}
	if len(vals) > 4 { // 存在 deleteRule
		fk.DeleteRule = vals[4]
	}

	return m.NewForeignKey(strings.ToLower(vals[0]), fk)
}
