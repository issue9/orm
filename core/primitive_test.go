// SPDX-License-Identifier: MIT

package core

import (
	"reflect"
	"testing"

	"github.com/issue9/assert/v3"
)

func TestGetPrimitiveType(t *testing.T) {
	a := assert.New(t, false)

	a.Equal(GetPrimitiveType(reflect.TypeOf(1)), Int)
	a.Equal(GetPrimitiveType(reflect.TypeOf([]byte{1, 2})), Bytes)
	a.Equal(GetPrimitiveType(reflect.TypeOf("string")), String)
	a.Equal(GetPrimitiveType(reflect.TypeOf(any(5))), Int)

	// 指针的 PrimitiveType
	x := 5
	a.Equal(GetPrimitiveType(reflect.TypeOf(&x)), Auto)

	// 自定义类型，但是未实现 PrimitiveTyper 接口
	type T int16
	a.Equal(GetPrimitiveType(reflect.TypeOf(T(1))), Int16)

	type obj struct{}
	a.Equal(GetPrimitiveType(reflect.TypeOf(obj{})), Auto)

	type obj2 struct {
		Any any
	}
	o2 := obj2{}
	field, _ := reflect.ValueOf(o2).Type().FieldByName("Any")
	a.Equal(GetPrimitiveType(field.Type), String)
}
