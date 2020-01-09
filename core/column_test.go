// SPDX-License-Identifier: MIT

package core

import (
	"testing"

	"github.com/issue9/assert"
)

func TestColumn_Clone(t *testing.T) {
	a := assert.New(t)

	col, err := NewColumnFromGoType(IntType)
	a.NotError(err).NotNil(col)
	col.Nullable = true

	cc := col.Clone()
	a.Equal(cc, col)  // 值相同
	a.True(cc != col) // 但不是同一个实例
}

func TestColumn_Check(t *testing.T) {
	a := assert.New(t)

	col, err := NewColumnFromGoType(StringType)
	a.NotError(err).NotNil(col)
	col.Length = []int{-1}

	a.NotError(col.Check())

	col.Length[0] = 0
	a.Error(col.Check())

	col.Length[0] = -2
	a.Error(col.Check())

	col, err = NewColumnFromGoType(IntType)
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
