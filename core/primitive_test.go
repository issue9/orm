// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package core

import (
	"reflect"
	"testing"

	"github.com/issue9/assert/v4"
)

func TestPrimitiveType(t *testing.T) {
	a := assert.New(t, false)

	// 保证 PrimitiveType.String() 拥有所有的值。
	for i := Auto; i < maxPrimitiveType; i++ {
		a.NotEmpty(i.String())
	}
	a.Length(typeStrings, int(maxPrimitiveType))
}

func TestGetPrimitiveType(t *testing.T) {
	a := assert.New(t, false)

	a.Equal(GetPrimitiveType(reflect.TypeFor[int]()), Int)
	a.Equal(GetPrimitiveType(reflect.TypeFor[[]byte]()), Bytes)
	a.Equal(GetPrimitiveType(reflect.TypeFor[string]()), String)
	a.Equal(GetPrimitiveType(reflect.TypeFor[any]()), Int)

	// 指针的 PrimitiveType
	a.Equal(GetPrimitiveType(reflect.TypeFor[*int]()), Auto)

	// 自定义类型，但是未实现 PrimitiveTyper 接口
	type T int16
	a.Equal(GetPrimitiveType(reflect.TypeFor[T]()), Int16)

	type obj struct{}
	a.Equal(GetPrimitiveType(reflect.TypeFor[obj]()), Auto)

	type obj2 struct {
		Any any
	}
	o2 := obj2{}
	field, _ := reflect.ValueOf(o2).Type().FieldByName("Any")
	a.Equal(GetPrimitiveType(field.Type), String)
}
