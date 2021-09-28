// SPDX-License-Identifier: MIT

package core

import (
	"database/sql"
	"reflect"
	"time"
)

// TimeFormatLayout 时间如果需要转换成字符串采用此格式
const TimeFormatLayout = time.RFC3339

// 所有的 PrimitiveType
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
	types = map[reflect.Type]PrimitiveType{
		reflect.TypeOf(true):           Bool,
		reflect.TypeOf(int(1)):         Int,
		reflect.TypeOf(int8(1)):        Int8,
		reflect.TypeOf(int16(1)):       Int16,
		reflect.TypeOf(int32(1)):       Int32,
		reflect.TypeOf(int64(1)):       Int64,
		reflect.TypeOf(uint(1)):        Uint,
		reflect.TypeOf(uint8(1)):       Uint8,
		reflect.TypeOf(uint16(1)):      Uint16,
		reflect.TypeOf(uint32(1)):      Uint32,
		reflect.TypeOf(uint64(1)):      Uint64,
		reflect.TypeOf(float32(1)):     Float32,
		reflect.TypeOf(float64(1)):     Float64,
		reflect.TypeOf(""):             String,
		reflect.TypeOf([]byte{}):       Bytes,
		reflect.TypeOf(sql.RawBytes{}): Bytes,
		reflect.TypeOf(time.Time{}):    Time,

		reflect.TypeOf(sql.NullString{}):  String,
		reflect.TypeOf(sql.NullInt64{}):   Int64,
		reflect.TypeOf(sql.NullInt32{}):   Int32,
		reflect.TypeOf(sql.NullBool{}):    Bool,
		reflect.TypeOf(sql.NullFloat64{}): Float64,
		reflect.TypeOf(sql.NullTime{}):    Time,
	}

	kinds = map[reflect.Kind]PrimitiveType{
		reflect.Bool:    Bool,
		reflect.Int:     Int,
		reflect.Int8:    Int8,
		reflect.Int16:   Int16,
		reflect.Int32:   Int32,
		reflect.Int64:   Int64,
		reflect.Uint:    Uint,
		reflect.Uint8:   Uint8,
		reflect.Uint16:  Uint16,
		reflect.Uint32:  Uint32,
		reflect.Uint64:  Uint64,
		reflect.Float32: Float32,
		reflect.Float64: Float64,
		reflect.String:  String,
	}

	primitiveTyperType = reflect.TypeOf((*PrimitiveTyper)(nil)).Elem()
)

// PrimitiveTyper 提供了 PrimitiveType 方法
//
// 如果用户需要将自定义类型写入数据，需要提供该类型所表示的 PrimitiveType 值，
// 最终会以该类型的值写入数据库。
//
// NOTE: 最简单的方法是复用 driver.Valuer 接口，从其返回值中获取类型信息，
// 但是该接口有可能返回 nil 值，无法确定类型。
type PrimitiveTyper interface {
	// PrimitiveType 返回当前对象所表示的 PrimitiveType 值
	//
	// NOTE: 每个对象在任何时间返回的值应该都是固定的。
	PrimitiveType() PrimitiveType
}

// PrimitiveType 表示 Dialect 支持的数据类型
//
// 所有 Go 对象需要指定一种类型作为其在数据库中的存储方式，
// 内置的 Go 类型除去 complex 之外将作为相应在的定义，可通过 GetPrimitiveType 获取。
//
// Dialect 的实现者也需要在 SQLType 中实现对所有 PrimitiveType 在数据库中的真实类型。
type PrimitiveType int

// GetPrimitiveType 获取 t 所关联的 PrimitiveType 值
//
// 如果 t.Kind == Ptr，则需要用户自行处理获取其对象的类型，否则返回 Auto。
func GetPrimitiveType(t reflect.Type) PrimitiveType {
	primitiveType, found := types[t]
	if !found {
		v := reflect.New(t).Elem()
		if t.Implements(primitiveTyperType) {
			primitiveType = v.Interface().(PrimitiveTyper).PrimitiveType()
		} else if v.Addr().Type().Implements(primitiveTyperType) {
			primitiveType = v.Addr().Interface().(PrimitiveTyper).PrimitiveType()
		}
	}

	if primitiveType == Auto {
		primitiveType = kinds[t.Kind()]
	}

	return primitiveType
}
