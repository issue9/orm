// SPDX-License-Identifier: MIT

package core

import (
	"testing"

	"github.com/issue9/assert/v3"
)

func TestNewColumn(t *testing.T) {
	a := assert.New(t, false)

	col, err := NewColumn(Int)
	a.NotError(err).NotNil(col).Equal(col.PrimitiveType, Int)

	col, err = NewColumn(Bool)
	a.NotError(err).NotNil(col).Equal(col.PrimitiveType, Bool)

	col, err = NewColumn(Auto)
	a.ErrorIs(err, ErrInvalidColumnType).Nil(col)

	col, err = NewColumn(maxPrimitiveType)
	a.ErrorIs(err, ErrInvalidColumnType).Nil(col)
}

func TestModel_AddColumns(t *testing.T) {
	a := assert.New(t, false)
	m := NewModel(Table, "m1", 10)
	a.NotNil(m)

	ai, err := NewColumn(Int)
	a.NotError(err).NotNil(ai)
	ai.AI = true
	a.Error(m.AddColumns(ai)) // 没有名称

	col, err := NewColumn(Int)
	a.NotError(err).NotNil(col)

	// 同名
	ai.Name = "ai"
	col.Name = "ai"
	a.Error(m.AddColumns(ai, col))

	// 正常
	m.Reset()
	col.Name = "col"
	a.NotError(m.AddColumns(ai, col))
}

func TestColumn_Clone(t *testing.T) {
	a := assert.New(t, false)

	col, err := NewColumn(Int)
	a.NotError(err).NotNil(col)
	col.Nullable = true

	cc := col.Clone()
	a.Equal(cc, col)  // 值相同
	a.True(cc != col) // 但不是同一个实例
}

func TestColumn_Check(t *testing.T) {
	a := assert.New(t, false)

	col, err := NewColumn(String)
	a.NotError(err).NotNil(col)
	col.Length = []int{-1}

	a.NotError(col.Check())

	col.Length[0] = 0
	a.Error(col.Check())

	col.Length[0] = -2
	a.Error(col.Check())

	col, err = NewColumn(Int)
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
