// SPDX-License-Identifier: MIT

package core

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"time"
)

// ErrInvalidColumnType 无效的列类型
//
// 作为列类型，该数据类型必须是可序列化的。
// 像 reflect.Func 和 reflect.Chan 等都将返回该错误。
var ErrInvalidColumnType = errors.New("无效的列类型")

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
	NullString
	NullInt64
	NullInt32
	NullBool
	NullFloat64
	RawBytes
	Time
	NullTime
)

// 基本的数据类型

var types = map[reflect.Type]PrimitiveType{
	reflect.TypeOf(true):              Bool,
	reflect.TypeOf(int(1)):            Int,
	reflect.TypeOf(int8(1)):           Int8,
	reflect.TypeOf(int16(1)):          Int16,
	reflect.TypeOf(int32(1)):          Int32,
	reflect.TypeOf(int64(1)):          Int64,
	reflect.TypeOf(uint(1)):           Uint,
	reflect.TypeOf(uint8(1)):          Uint8,
	reflect.TypeOf(uint16(1)):         Uint16,
	reflect.TypeOf(uint32(1)):         Uint32,
	reflect.TypeOf(uint64(1)):         Uint64,
	reflect.TypeOf(float32(1)):        Float32,
	reflect.TypeOf(float64(1)):        Float64,
	reflect.TypeOf(""):                String,
	reflect.TypeOf(sql.NullString{}):  NullString,
	reflect.TypeOf(sql.NullInt64{}):   NullInt64,
	reflect.TypeOf(sql.NullInt32{}):   NullInt32,
	reflect.TypeOf(sql.NullBool{}):    NullBool,
	reflect.TypeOf(sql.NullFloat64{}): NullFloat64,
	reflect.TypeOf(sql.RawBytes{}):    RawBytes,
	reflect.TypeOf(time.Time{}):       Time,
	reflect.TypeOf(sql.NullTime{}):    NullTime,
}

var primitiveTyperType = reflect.TypeOf((*PrimitiveTyper)(nil)).Elem()

// DefaultParser 提供了 ParseDefault 函数
//
// 在 struct tag 中可以通过 default 指定默认值，
// 该值的表示可能与数据库中的表示不尽相同，
// 所以自定义的数据类型，需要实现该接口，以便能正确转换成该类型的值。
//
// 如果用户不提供该接口实现，那么默认情况下，
// 系统会采用 github.com/issue9/conv.Value() 函数作默认转换。
type DefaultParser interface {
	// 将默认值从字符串解析成 t 类型的值
	ParseDefault(v string) error
}

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

// PrimitiveType 表示支持转换成数据库类型的 Go 类型信息
type PrimitiveType int

// Column 列结构
type Column struct {
	Name       string // 数据库的字段名
	AI         bool
	Nullable   bool
	HasDefault bool
	Default    interface{}
	Length     []int

	PrimitiveType PrimitiveType
	GoType        reflect.Type // Go 语言中的数据类型
	GoName        string       // Go 中的字段名
}

// NewColumnFromGoType 从 Go 类型中生成 Column
func NewColumnFromGoType(goType reflect.Type) (*Column, error) {
	for goType.Kind() == reflect.Ptr {
		goType = goType.Elem()
	}

	primitiveType, found := types[goType]
	if !found {
		v := reflect.New(goType).Elem()
		if goType.Implements(primitiveTyperType) {
			primitiveType = v.Interface().(PrimitiveTyper).PrimitiveType()
		} else if v.Addr().Type().Implements(primitiveTyperType) {
			primitiveType = v.Addr().Interface().(PrimitiveTyper).PrimitiveType()
		}
	}

	if primitiveType == Auto || goType.Kind() == reflect.Chan || goType.Kind() == reflect.Func {
		return nil, ErrInvalidColumnType
	}

	return &Column{
		PrimitiveType: primitiveType,
		GoType:        goType,
	}, nil
}

// Clone 复制 Column
func (c *Column) Clone() *Column {
	cc := &Column{}
	*cc = *c

	return cc
}

// SetDefault 为列设置默认值
func (c *Column) SetDefault(v interface{}) {
	c.HasDefault = true
	c.Default = v
}

// Check 检测 Column 内容是否合法
func (c *Column) Check() error {
	if c.AI && c.HasDefault {
		return fmt.Errorf("AutoIncrement 列 %s 不能同时包含默认值", c.Name)
	}

	if c.AI && c.Nullable {
		return fmt.Errorf("AutoIncrement 列 %s 不能同时带 NULL 约束", c.Name)
	}

	if c.PrimitiveType == String || c.PrimitiveType == NullString {
		if len(c.Length) > 0 && (c.Length[0] < -1 || c.Length[0] == 0) {
			return fmt.Errorf("列 %s 的长度只能是 -1 或是 >0", c.Name)
		}
	} else {
		for _, v := range c.Length {
			if v < 0 {
				return fmt.Errorf("列 %s 的长度只能是不能小于 0", c.Name)
			}
		}
	}

	return nil
}

// FindColumn 查找指定名称的列
//
// 不存在该列则返回 nil
func (m *Model) FindColumn(name string) *Column {
	for _, col := range m.Columns {
		if col.Name == name {
			return col
		}
	}
	return nil
}

func errColumnNotFound(col string) error {
	return fmt.Errorf("列 %s 未找到", col)
}

func errColumnExists(col string) error {
	return fmt.Errorf("列 %s 已经存在", col)
}

func (m *Model) columnExists(col *Column) bool {
	for _, c := range m.Columns {
		if c == col {
			return true
		}
	}

	return false
}
