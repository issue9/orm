// SPDX-License-Identifier: MIT

package model

import (
	"database/sql"
	"reflect"
	"strconv"
	"time"

	"github.com/issue9/conv"

	"github.com/issue9/orm/v3/core"
	"github.com/issue9/orm/v3/internal/tags"
)

func newColumn(field reflect.StructField) (*core.Column, error) {
	col, err := core.NewColumnFromGoType(field.Type)
	if err != nil {
		return nil, err
	}

	col.Name = field.Name
	col.GoName = field.Name
	return col, nil
}

// 分析一个字段
func parseColumn(m *core.Model, col *core.Column, tag string) (err error) {
	if err = m.AddColumn(col); err != nil {
		return err
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

// 从参数中获取 Column 的 len1 和 len2 变量
//  len(len1,len2)
func setColumnLen(c *core.Column, vals []string) (err error) {
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

// nullable
func setColumnNullable(c *core.Column, vals []string) (err error) {
	if len(vals) > 0 {
		return propertyError(c.Name, "nullable", "指定了太多的值")
	}

	if c.AI {
		return propertyError(c.Name, "nullable", "自增列不能设置此值")
	}

	c.Nullable = true
	return nil
}

// default(5)
func setDefault(col *core.Column, vals []string) error {
	if len(vals) != 1 {
		return propertyError(col.Name, "default", "太多的值")
	}
	col.HasDefault = true

	rval := reflect.New(col.GoType)
	v := rval.Interface()
	if p, ok := v.(core.DefaultParser); ok {
		if err := p.ParseDefault(vals[0]); err != nil {
			return err
		}

		col.Default = v
		return nil
	}
	if p, ok := v.(sql.Scanner); ok {
		if err := p.Scan(vals[0]); err != nil {
			return err
		}
		col.Default = v
		return nil
	}

	switch col.GoType {
	case core.TimeType:
		v, err := time.Parse(time.RFC3339, vals[0])
		if err != nil {
			return err
		}
		col.Default = v
	default:
		for rval.Kind() == reflect.Ptr {
			rval = rval.Elem()
		}

		if err := conv.Value(vals[0], rval); err != nil {
			return err
		}
		col.Default = rval.Interface()
	}

	return nil
}

// ai
func setAI(m *core.Model, col *core.Column, vals []string) error {
	if len(vals) != 0 {
		return propertyError(col.Name, "ai", "太多的值")
	}

	return m.SetAutoIncrement(col)
}
