// SPDX-FileCopyrightText: 2014-2025 caixw
//
// SPDX-License-Identifier: MIT

package core

import (
	"database/sql"
	"reflect"
	"time"
)

// TimeFormatLayout 时间如果需要转换成字符串采用此格式
const TimeFormatLayout = time.RFC3339

// 当前支持的 [PrimitiveType] 值
//
// 其中的 [String] 被设计成可以保存部分类型为 [reflect.Interface] 的数据结构，
// 但是一个有限的集合，比如将一个 any 字段赋予 slice 类型，在保存时可能不被支持。
// 且在读取时，各个数据库略有不同，比如 mysql 返回 []byte，而其它数据一般返回 string。
const (
	Auto PrimitiveType = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Float32
	Float64
	String
	Bytes
	Time
	Decimal
	maxPrimitiveType
)

var (
	typeStrings = map[PrimitiveType]string{
		Auto:    "auto",
		Bool:    "bool",
		Int:     "int",
		Int8:    "int8",
		Int16:   "int16",
		Int32:   "int32",
		Int64:   "int64",
		Uint:    "uint",
		Uint8:   "uint8",
		Uint16:  "uint16",
		Uint32:  "uint32",
		Uint64:  "uint64",
		Float32: "float32",
		Float64: "float64",
		String:  "string",
		Bytes:   "bytes",
		Time:    "time",
		Decimal: "decimal",
	}

	types = map[reflect.Type]PrimitiveType{
		reflect.TypeFor[bool]():         Bool,
		reflect.TypeFor[int]():          Int,
		reflect.TypeFor[int8]():         Int8,
		reflect.TypeFor[int16]():        Int16,
		reflect.TypeFor[int32]():        Int32,
		reflect.TypeFor[int64]():        Int64,
		reflect.TypeFor[uint]():         Uint,
		reflect.TypeFor[uint8]():        Uint8,
		reflect.TypeFor[uint16]():       Uint16,
		reflect.TypeFor[uint32]():       Uint32,
		reflect.TypeFor[uint64]():       Uint64,
		reflect.TypeFor[float32]():      Float32,
		reflect.TypeFor[float64]():      Float64,
		reflect.TypeFor[string]():       String,
		reflect.TypeFor[[]byte]():       Bytes,
		reflect.TypeFor[sql.RawBytes](): Bytes,
		reflect.TypeFor[time.Time]():    Time,

		reflect.TypeFor[sql.NullString]():  String,
		reflect.TypeFor[sql.NullByte]():    Bytes,
		reflect.TypeFor[sql.NullInt64]():   Int64,
		reflect.TypeFor[sql.NullInt32]():   Int32,
		reflect.TypeFor[sql.NullInt16]():   Int16,
		reflect.TypeFor[sql.NullBool]():    Bool,
		reflect.TypeFor[sql.NullFloat64](): Float64,
		reflect.TypeFor[sql.NullTime]():    Time,
	}

	kinds = map[reflect.Kind]PrimitiveType{
		reflect.Bool:      Bool,
		reflect.Int:       Int,
		reflect.Int8:      Int8,
		reflect.Int16:     Int16,
		reflect.Int32:     Int32,
		reflect.Int64:     Int64,
		reflect.Uint:      Uint,
		reflect.Uint8:     Uint8,
		reflect.Uint16:    Uint16,
		reflect.Uint32:    Uint32,
		reflect.Uint64:    Uint64,
		reflect.Float32:   Float32,
		reflect.Float64:   Float64,
		reflect.String:    String,
		reflect.Interface: String,
	}

	primitiveTyperType = reflect.TypeFor[PrimitiveTyper]()
)

type PrimitiveTyper interface {
	// NOTE: 最简单的方法是复用 [driver.Valuer] 接口，从其返回值中获取类型信息，
	// 但是该接口有可能返回 nil 值，无法确定类型。

	// PrimitiveType 返回当前对象所表示的 [PrimitiveType] 值
	//
	// NOTE: 每个对象在任何时间返回的值应该都是固定的。
	PrimitiveType() PrimitiveType
}

// PrimitiveType 表示 Go 对象在数据库中实际的存储方式
//
// PrimitiveType 由 [Dialect.SQLType] 转换成相应数据的实际类型。
type PrimitiveType int

// GetPrimitiveType 获取 t 所关联的 [PrimitiveType] 值
//
// t.Kind 不能为 [reflect.Ptr] 否则将返回 [Auto]。
func GetPrimitiveType(t reflect.Type) PrimitiveType {
	primitiveType, found := kinds[t.Kind()]
	if found {
		return primitiveType
	}

	primitiveType, found = types[t]
	if !found {
		v := reflect.New(t).Elem()
		if t.Implements(primitiveTyperType) {
			primitiveType = v.Interface().(PrimitiveTyper).PrimitiveType()
		} else if v.Addr().Type().Implements(primitiveTyperType) {
			primitiveType = v.Addr().Interface().(PrimitiveTyper).PrimitiveType()
		}
	}

	return primitiveType
}

func (t PrimitiveType) String() string { return typeStrings[t] }
