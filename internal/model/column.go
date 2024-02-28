// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package model

import (
	"database/sql"
	"reflect"
	"strconv"
	"time"

	"github.com/issue9/conv"

	"github.com/issue9/orm/v5/core"
	"github.com/issue9/orm/v5/internal/tags"
)

type column struct {
	*core.Column
	GoType reflect.Type
}

func newColumn(field reflect.StructField) (*column, error) {
	t := field.Type
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	col, err := core.NewColumn(core.GetPrimitiveType(t))
	if err != nil {
		return nil, err
	}

	col.Name = field.Name
	col.GoName = field.Name
	return &column{
		Column: col,
		GoType: t,
	}, nil
}

func (col *column) parseTags(m *core.Model, tag string) (err error) {
	if err = m.AddColumn(col.Column); err != nil {
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
			err = col.setNullable(tag.Args)
		case "ai":
			err = col.setAI(m, tag.Args)
		case "len":
			err = col.setLen(tag.Args)
		case "fk":
			err = setFK(m, col, tag.Args)
		case "default":
			err = col.setDefault(tag.Args)
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
//
//	len(len1,len2)
func (col *column) setLen(vals []string) (err error) {
	l := len(vals)
	switch l {
	case 1:
	case 2:
	case 0:
		return nil
	default:
		return propertyError(col.Name, "len", "过多的参数")
	}

	col.Length = make([]int, 0, l)
	for _, val := range vals {
		v, err := strconv.Atoi(val)
		if err != nil {
			return err
		}
		col.Length = append(col.Length, v)
	}

	return nil
}

// nullable
func (col *column) setNullable(vals []string) (err error) {
	if len(vals) > 0 {
		return propertyError(col.Name, "nullable", "指定了太多的值")
	}

	if col.AI {
		return propertyError(col.Name, "nullable", "自增列不能设置此值")
	}

	col.Nullable = true
	return nil
}

// default(5)
func (col *column) setDefault(vals []string) error {
	if len(vals) != 1 {
		return propertyError(col.Name, "default", "太多的值")
	}
	col.HasDefault = true

	rval := reflect.New(col.GoType)

	v := rval.Interface()
	if p, ok := v.(sql.Scanner); ok {
		if err := p.Scan(vals[0]); err != nil {
			return err
		}
		col.Default = v
		return nil
	}
	v = rval.Elem().Interface()
	if p, ok := v.(sql.Scanner); ok {
		if err := p.Scan(vals[0]); err != nil {
			return err
		}
		col.Default = v
		return nil
	}

	switch col.PrimitiveType {
	case core.Time:
		v, err := time.Parse(core.TimeFormatLayout, vals[0])
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
func (col *column) setAI(m *core.Model, vals []string) error {
	if len(vals) != 0 {
		return propertyError(col.Name, "ai", "太多的值")
	}

	return m.SetAutoIncrement(col.Column)
}
