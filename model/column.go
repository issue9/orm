// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"reflect"
	"strconv"

	"github.com/issue9/orm/v2/core"
)

// Column 列结构
type Column struct {
	*core.Column

	model  *Model
	zero   interface{} // GoType 的零值
	GoName string      // 结构字段名
}

func (m *Model) newColumn(field reflect.StructField) *Column {
	return &Column{
		Column: &core.Column{
			Name:   field.Name,
			GoType: field.Type,
		},
		zero:   reflect.Zero(field.Type).Interface(),
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

// 检测长试是否合法，必须要  Column 初始化已经完成。
func (c *Column) checkLen() error {
	if c.GoType.Kind() == reflect.String {
		if len(c.Length) > 0 && (c.Length[0] < -1 || c.Length[0] == 0) {
			return propertyError(c.Name, "len", "必须大于 0 或是等于 -1")
		}
	} else {
		for _, v := range c.Length {
			if v < 0 {
				return propertyError(c.Name, "len", "不能小于 0")
			}
		}
	}

	return nil
}

// 从参数中获取 Column 的 len1 和 len2 变量。
// len(len1,len2)
func (c *Column) setLen(vals []string) (err error) {
	l := len(vals)
	switch l {
	case 1:
	case 2:
	case 0:
		return nil
	default:
		return propertyError(c.Name, "len", "过多的参数")
	}

	c.Length = make([]int, 0, l)
	for _, val := range vals {
		v, err := strconv.Atoi(val)
		if err != nil {
			return err
		}
		c.Length = append(c.Length, v)
	}

	return nil
}

// 从 vals 中分析，得出 Column.Nullable 的值。
// nullable; or nullable(true);
func (c *Column) setNullable(vals []string) (err error) {
	if c.AI {
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
