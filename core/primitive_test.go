// SPDX-License-Identifier: MIT

package core

import (
	"reflect"
	"testing"

	"github.com/issue9/assert"
)

func TestGetPrimitiveType(t *testing.T) {
	a := assert.New(t)

	a.Equal(GetPrimitiveType(reflect.TypeOf(1)), Int)
	a.Equal(GetPrimitiveType(reflect.TypeOf([]byte{1, 2})), Bytes)
	a.Equal(GetPrimitiveType(reflect.TypeOf("string")), String)

	// 指针的 PrimitiveType
	x := 5
	a.Equal(GetPrimitiveType(reflect.TypeOf(&x)), Auto)

	// 自定义类型，但是未实现 PrimitiveTyper 接口
	type T int16
	a.Equal(GetPrimitiveType(reflect.TypeOf(T(1))), Int16)

	type obj struct{}
	a.Equal(GetPrimitiveType(reflect.TypeOf(obj{})), Auto)
}
