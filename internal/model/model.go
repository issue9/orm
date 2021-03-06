// SPDX-License-Identifier: MIT

// Package model 管理数据模型
package model

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/issue9/orm/v3/core"
	"github.com/issue9/orm/v3/fetch"
	"github.com/issue9/orm/v3/internal/tags"
)

func propertyError(field, name, message string) error {
	return fmt.Errorf("%s 的 %s 属性发生以下错误: %s", field, name, message)
}

// New 从一个 obj 声明一个 Model 实例
//
// obj 可以是一个 struct 实例或是指针。
func (ms *Models) New(obj interface{}) (*core.Model, error) {
	rtype := reflect.TypeOf(obj)
	for rtype.Kind() == reflect.Ptr {
		rtype = rtype.Elem()
	}

	if rtype.Kind() != reflect.Struct {
		return nil, fetch.ErrInvalidKind
	}

	ms.locker.Lock()
	defer ms.locker.Unlock()

	if m, found := ms.items[rtype]; found {
		return m, nil
	}

	m := core.NewModel(core.Table, "#"+rtype.Name(), rtype.NumField())
	m.GoType = rtype

	if err := parseColumns(m, rtype); err != nil {
		return nil, err
	}

	if meta, ok := obj.(core.Metaer); ok {
		if err := parseMeta(m, meta.Meta()); err != nil {
			return nil, err
		}
	}

	if view, ok := obj.(core.Viewer); ok {
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

func (ms *Models) addModel(goType reflect.Type, m *core.Model) (err error) {
	ms.items[goType] = m

	w := func(name string) {
		if err == nil {
			err = ms.addNames(name)
		}
	}

	for name := range m.Indexes {
		w(name)
	}

	for name := range m.Uniques {
		w(name)
	}

	for name := range m.Checks {
		w(name)
	}

	for name := range m.ForeignKeys {
		w(name)
	}

	if m.AutoIncrement != nil {
		w(m.AIName())
	}

	if len(m.PrimaryKey) > 0 {
		w(m.PKName())
	}

	return
}

// 将 rval 中的结构解析到 m 中，支持匿名字段。
func parseColumns(m *core.Model, rtype reflect.Type) error {
	num := rtype.NumField()
	for i := 0; i < num; i++ {
		field := rtype.Field(i)

		if field.Anonymous {
			if err := parseColumns(m, field.Type); err != nil {
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

		if err := col.parseTags(m, tag); err != nil {
			return err
		}
	}

	return nil
}

// 分析 meta 接口数据
func parseMeta(m *core.Model, tag string) error {
	ts := tags.Parse(tag)
	for _, v := range ts {
		switch v.Name {
		case "name":
			if len(v.Args) != 1 {
				return propertyError("Metaer", "name", "太多的值")
			}

			m.Name = "#" + v.Args[0] // 所有 model 生成的表都带 #
		case "check":
			if len(v.Args) != 2 {
				return propertyError("Metaer", "check", "参数个数不正确")
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
func setOCC(m *core.Model, c *column, vals []string) error {
	if len(vals) > 0 {
		return propertyError(c.Name, "occ", "指定了太多的值")
	}
	return m.SetOCC(c.Column)
}

// index(idx_name)
func setIndex(m *core.Model, col *column, vals []string) error {
	if len(vals) != 1 {
		return propertyError(col.Name, "index", "太多的值")
	}
	return m.AddIndex(core.IndexDefault, strings.ToLower(vals[0]), col.Column)
}

// pk
func setPK(m *core.Model, col *column, vals []string) error {
	if len(vals) != 0 {
		return propertyError(col.Name, "pk", "太多的值")
	}
	return m.AddPrimaryKey(col.Column)
}

// unique(unique_name)
func setUnique(m *core.Model, col *column, vals []string) error {
	if len(vals) != 1 {
		return propertyError(col.Name, "unique", "只能带一个参数")
	}
	return m.AddUnique(strings.ToLower(vals[0]), col.Column)
}

// fk(fk_name,refTable,refColName,updateRule,deleteRule)
func setFK(m *core.Model, col *column, vals []string) error {
	if len(vals) < 3 || len(vals) > 5 {
		return propertyError(col.Name, "fk", "参数数量不正确")
	}

	fk := &core.ForeignKey{
		Column:       col.Column,
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
