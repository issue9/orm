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

// 基本的数据类型
var types = map[reflect.Type]core.PrimitiveType{
	reflect.TypeOf(true):              core.Bool,
	reflect.TypeOf(int(1)):            core.Int,
	reflect.TypeOf(int8(1)):           core.Int8,
	reflect.TypeOf(int16(1)):          core.Int16,
	reflect.TypeOf(int32(1)):          core.Int32,
	reflect.TypeOf(int64(1)):          core.Int64,
	reflect.TypeOf(uint(1)):           core.Uint,
	reflect.TypeOf(uint8(1)):          core.Uint8,
	reflect.TypeOf(uint16(1)):         core.Uint16,
	reflect.TypeOf(uint32(1)):         core.Uint32,
	reflect.TypeOf(uint64(1)):         core.Uint64,
	reflect.TypeOf(float32(1)):        core.Float32,
	reflect.TypeOf(float64(1)):        core.Float64,
	reflect.TypeOf(""):                core.String,
	reflect.TypeOf(sql.NullString{}):  core.NullString,
	reflect.TypeOf(sql.NullInt64{}):   core.NullInt64,
	reflect.TypeOf(sql.NullInt32{}):   core.NullInt32,
	reflect.TypeOf(sql.NullBool{}):    core.NullBool,
	reflect.TypeOf(sql.NullFloat64{}): core.NullFloat64,
	reflect.TypeOf([]byte{}):          core.Bytes,
	reflect.TypeOf(sql.RawBytes{}):    core.RawBytes,
	reflect.TypeOf(time.Time{}):       core.Time,
	reflect.TypeOf(sql.NullTime{}):    core.NullTime,
}

var primitiveTyperType = reflect.TypeOf((*core.PrimitiveTyper)(nil)).Elem()

type column struct {
	*core.Column
	GoType reflect.Type
}

func newColumn(field reflect.StructField) (*column, error) {
	t := field.Type
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	primitiveType, found := types[t]
	if !found {
		v := reflect.New(t).Elem()
		if t.Implements(primitiveTyperType) {
			primitiveType = v.Interface().(core.PrimitiveTyper).PrimitiveType()
		} else if v.Addr().Type().Implements(primitiveTyperType) {
			primitiveType = v.Addr().Interface().(core.PrimitiveTyper).PrimitiveType()
		}
	}

	col, err := core.NewColumn(primitiveType)
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
//  len(len1,len2)
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
	if p, ok := v.(core.DefaultParser); ok {
		if err := p.ParseDefault(vals[0]); err != nil {
			return err
		}
		col.Default = v
		return nil
	}
	v = rval.Elem().Interface()
	if p, ok := v.(core.DefaultParser); ok {
		if err := p.ParseDefault(vals[0]); err != nil {
			return err
		}
		col.Default = v
		return nil
	}

	v = rval.Interface()
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
