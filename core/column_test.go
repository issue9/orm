// SPDX-License-Identifier: MIT

package core

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/issue9/assert"
)

func TestNewColumnFromGoType(t *testing.T) {
	a := assert.New(t)

	col, err := NewColumnFromGoType(reflect.TypeOf(1))
	a.NotError(err).NotNil(col).Equal(col.PrimitiveType, Int)

	col, err = NewColumnFromGoType(reflect.TypeOf(int8(1)))
	a.NotError(err).NotNil(col).Equal(col.PrimitiveType, Int8)

	col, err = NewColumnFromGoType(reflect.TypeOf(true))
	a.NotError(err).NotNil(col).Equal(col.PrimitiveType, Bool)

	col, err = NewColumnFromGoType(reflect.TypeOf(sql.NullFloat64{}))
	a.NotError(err).NotNil(col).Equal(col.PrimitiveType, NullFloat64)

	col, err = NewColumnFromGoType(reflect.TypeOf(struct{}{}))
	a.Error(err).Nil(col)

	col, err = NewColumnFromGoType(reflect.TypeOf(Unix{}))
	a.NotError(err).NotNil(col).Equal(col.PrimitiveType, Int64)

	col, err = NewColumnFromGoType(reflect.TypeOf(&Unix{}))
	a.NotError(err).NotNil(col).Equal(col.PrimitiveType, Int64)

	col, err = NewColumnFromGoType(reflect.TypeOf(Unix{}))
	a.NotError(err).NotNil(col).Equal(col.PrimitiveType, Int64)
}

func TestColumn_Clone(t *testing.T) {
	a := assert.New(t)

	col, err := NewColumnFromGoType(reflect.TypeOf(1))
	a.NotError(err).NotNil(col)
	col.Nullable = true

	cc := col.Clone()
	a.Equal(cc, col)  // 值相同
	a.True(cc != col) // 但不是同一个实例
}

func TestColumn_Check(t *testing.T) {
	a := assert.New(t)

	col, err := NewColumnFromGoType(reflect.TypeOf(""))
	a.NotError(err).NotNil(col)
	col.Length = []int{-1}

	a.NotError(col.Check())

	col.Length[0] = 0
	a.Error(col.Check())

	col.Length[0] = -2
	a.Error(col.Check())

	col, err = NewColumnFromGoType(reflect.TypeOf(5))
	a.NotError(err).NotNil(col)
	col.Length = []int{-2}
	a.Error(col.Check())

	col.Length[0] = -1
	a.Error(col.Check())

	col.Length[0] = 0
	a.NotError(col.Check())

	col.AI = true
	col.HasDefault = true
	a.Error(col.Check())

	col.AI = true
	col.HasDefault = false
	col.Nullable = true
	a.Error(col.Check())
}
