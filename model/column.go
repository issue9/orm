// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"fmt"
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
	Zero     interface{}  // GoType 的零值
	GoName   string       // 结构字段名

	HasDefault bool
	Default    string // 默认值
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
