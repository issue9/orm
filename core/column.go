// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package core

import (
	"database/sql"
	"reflect"
	"time"
)

// CreateTableStmt.Column 用到的数据类型。
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

	//UintptrType=reflect.TypeOf(uintptr(1))
	//Complex64Type=reflect.TypeOf(complex64(1,1))
	//Complex128Type=reflect.TypeOf(complex128(1,1))
)

// Column 列结构
type Column struct {
	Name       string       // 数据库的字段名
	GoType     reflect.Type // Go 语言中的数据类型
	AI         bool
	Nullable   bool
	HasDefault bool
	Default    interface{}
	Length     []int
}
