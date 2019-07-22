// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package core

import (
	"database/sql"
	"reflect"
	"time"
)

// 基本的数据类型
var (
	BoolType    = reflect.TypeOf(true)
	IntType     = reflect.TypeOf(int(1))
	Int8Type    = reflect.TypeOf(int8(1))
	Int16Type   = reflect.TypeOf(int16(1))
	Int32Type   = reflect.TypeOf(int32(1))
	Int64Type   = reflect.TypeOf(int64(1))
	UintType    = reflect.TypeOf(uint(1))
	Uint8Type   = reflect.TypeOf(uint8(1))
	Uint16Type  = reflect.TypeOf(uint16(1))
	Uint32Type  = reflect.TypeOf(uint32(1))
	Uint64Type  = reflect.TypeOf(uint64(1))
	Float32Type = reflect.TypeOf(float32(1))
	Float64Type = reflect.TypeOf(float64(1))
	StringType  = reflect.TypeOf("")

	NullStringType  = reflect.TypeOf(sql.NullString{})
	NullInt64Type   = reflect.TypeOf(sql.NullInt64{})
	NullBoolType    = reflect.TypeOf(sql.NullBool{})
	NullFloat64Type = reflect.TypeOf(sql.NullFloat64{})
	RawBytesType    = reflect.TypeOf(sql.RawBytes{})
	TimeType        = reflect.TypeOf(time.Time{})
)

// Column 列结构
type Column struct {
	Name       string // 数据库的字段名
	AI         bool
	Nullable   bool
	HasDefault bool
	Default    interface{}
	Length     []int

	GoType reflect.Type // Go 语言中的数据类型
	GoName string       // Go 中的字段名
	goZero interface{}  // Go 中的零值
}

// NewColumnFromGoType 从 Go 类型中生成 Column，会初始化 goZero
func NewColumnFromGoType(goType reflect.Type) *Column {
	return &Column{
		GoType: goType,
		goZero: reflect.Zero(goType).Interface(),
	}
}

// IsZero 是否为零值
func (c *Column) IsZero(v reflect.Value) bool {
	if !v.IsValid() {
		return false
	}

	if c.GoType.Comparable() {
		return c.goZero == v.Interface()
	}

	if v.Kind() == reflect.Slice {
		return v.Len() == 0
	}

	return false
}

// Clone 复制 Column
func (c *Column) Clone() *Column {
	cc := &Column{}
	*cc = *c

	return cc
}
