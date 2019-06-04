// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"reflect"
	"strconv"
)

// Column 列结构
type Column struct {
	model *Model

	Name     string       // 数据库的字段名
	Len1     int          // 长度1，仅对部分类型启作用
	Len2     int          // 长度2，仅对部分类型启作用
	Nullable bool         // 是否可以为 NULL
	GoType   reflect.Type // Go 语言中的数据类型
	zero     interface{}  // GoType 的零值
	GoName   string       // 结构字段名

	HasDefault bool
	Default    string // 默认值
}

func (m *Model) newColumn(field reflect.StructField) *Column {
	return &Column{
		GoType: field.Type,
		zero:   reflect.Zero(field.Type).Interface(),
		Name:   field.Name,
		model:  m,
		GoName: field.Name,
	}
}

// IsZero 是否为零值
func (c *Column) IsZero(v reflect.Value) bool {
	if !v.IsValid() {
		return false
	}

	if c.GoType.Comparable() {
		return c.zero == v.Interface()
	}

	if v.Kind() == reflect.Slice {
		return v.Len() == 0
	}

	return false
}

// IsAI 当前列是否为自增列
func (c *Column) IsAI() bool {
	return (c.model != nil) && (c.model.AI == c)
}

// 检测长试是否合法，必须要  Column 初始化已经完成。
func (c *Column) checkLen() error {
	if c.GoType.Kind() == reflect.String {
		if c.Len1 == 0 {
			return propertyError(c.Name, "len", "字符类型，必须指定 len 属性")
		}

		if c.Len1 < -1 {
			return propertyError(c.Name, "len", "必须大于 0 或是 -1")
		}
	} else {
		if c.Len1 < 0 { // 不能为负数
			return propertyError(c.Name, "len", "不能小于 0")
		}

		if c.Len2 < 0 { // 不能为负数
			return propertyError(c.Name, "len", "不能小于 0")
		}
	}

	return nil
}

// 从参数中获取 Column 的 len1 和 len2 变量。
// len(len1,len2)
func (c *Column) setLen(vals []string) (err error) {
	switch len(vals) {
	case 0:
	case 1:
		c.Len1, err = strconv.Atoi(vals[0])
	case 2:
		if c.Len1, err = strconv.Atoi(vals[0]); err != nil {
			return err
		}

		c.Len2, err = strconv.Atoi(vals[1])
	default:
		return propertyError(c.Name, "len", "过多的参数")
	}

	return
}

// 从 vals 中分析，得出 Column.Nullable 的值。
// nullable; or nullable(true);
func (c *Column) setNullable(vals []string) (err error) {
	if c.IsAI() {
		return propertyError(c.Name, "nullable", "自增列不能设置此值")
	}

	switch len(vals) {
	case 0:
		c.Nullable = true
	case 1:
		if c.Nullable, err = strconv.ParseBool(vals[0]); err != nil {
			return err
		}
	default:
		return propertyError(c.Name, "nullable", "过多的参数值")
	}

	return nil
}
