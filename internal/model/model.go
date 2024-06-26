// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package model 管理数据模型
package model

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/issue9/orm/v6/core"
	"github.com/issue9/orm/v6/fetch"
)

func propertyError(field, name, message string) error {
	return fmt.Errorf("%s 的 %s 属性发生以下错误: %s", field, name, message)
}

// New 从一个 obj 声明 [core.Model] 实例
//
// obj 可以是一个结构体或是指针。
func (ms *Models) New(obj core.TableNamer) (*core.Model, error) {
	rtype := reflect.TypeOf(obj)
	for rtype.Kind() == reflect.Ptr {
		rtype = rtype.Elem()
	}

	if rtype.Kind() != reflect.Struct {
		return nil, fetch.ErrUnsupportedKind()
	}

	if m, found := ms.models.Load(rtype); found {
		return m.(*core.Model), nil
	}

	m := core.NewModel(core.Table, "#"+obj.TableName(), rtype.NumField())
	m.GoType = rtype

	if err := parseColumns(m, rtype); err != nil {
		return nil, err
	}

	if am, ok := obj.(core.ApplyModeler); ok {
		if err := am.ApplyModel(m); err != nil {
			return nil, err
		}
	}

	if view, ok := obj.(core.Viewer); ok {
		m.Type = core.View
		sql, err := view.ViewAs()
		if err != nil {
			return nil, err
		}
		m.ViewAs = sql
	}

	if err := m.Sanitize(); err != nil {
		return nil, err
	}

	// 在构建完 core.Model 时在其它地方写入了相同名称的 core.Model，
	// 相当于在函数的开始阶段判断是否存在同名的对象，返回已经存在的对象。
	if m, found := ms.models.Load(rtype); found {
		return m.(*core.Model), nil
	}
	ms.models.Store(rtype, m)
	return m, nil
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

		col, err := NewColumn(field)
		if err != nil {
			return err
		}

		// 这属于代码级别的错误，直接 panic 了。
		if err := col.ParseTags(m, tag); err != nil {
			panic(err)
		}
	}

	return nil
}

// occ
func SetOCC(m *core.Model, c *Column, vals []string) error {
	if len(vals) > 0 {
		return propertyError(c.Name, "occ", "指定了太多的值")
	}
	return m.SetOCC(c.Column)
}

// index(idx_name)
func setIndex(m *core.Model, col *Column, vals []string) error {
	if len(vals) != 1 {
		return propertyError(col.Name, "index", "太多的值")
	}
	return m.AddIndex(core.IndexDefault, strings.ToLower(vals[0]), col.Column)
}

// pk
func SetPK(m *core.Model, col *Column, vals []string) error {
	if len(vals) != 0 {
		return propertyError(col.Name, "pk", "太多的值")
	}
	return m.AddPrimaryKey(col.Column)
}

// unique(unique_name)
func setUnique(m *core.Model, col *Column, vals []string) error {
	if len(vals) != 1 {
		return propertyError(col.Name, "unique", "只能带一个参数")
	}
	return m.AddUnique(strings.ToLower(vals[0]), col.Column)
}

// fk(fk_name,refTable,refColName,updateRule,deleteRule)
func setFK(m *core.Model, col *Column, vals []string) error {
	if len(vals) < 3 || len(vals) > 5 {
		return propertyError(col.Name, "fk", "参数数量不正确")
	}

	fk := &core.ForeignKey{
		Name:         strings.ToLower(vals[0]),
		Column:       col.Column,
		RefTableName: "#" + vals[1],
		RefColName:   vals[2],
	}

	if len(vals) > 3 { // 存在 updateRule
		fk.UpdateRule = vals[3]
	}
	if len(vals) > 4 { // 存在 deleteRule
		fk.DeleteRule = vals[4]
	}

	return m.NewForeignKey(fk)
}
