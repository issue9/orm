// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/core"
)

func TestSetColumnLen(t *testing.T) {
	a := assert.New(t)
	col := &core.Column{}

	a.NotError(setColumnLen(col, []string{})).Empty(col.Length)
	a.NotError(setColumnLen(col, []string{"1", "2"})).
		Equal(col.Length[0], 1).
		Equal(col.Length[1], 2)
	a.Error(setColumnLen(col, []string{"1", "2", "3"}))
	a.Error(setColumnLen(col, []string{"1", "one"}))
	a.Error(setColumnLen(col, []string{"one", "one"}))
}

func TestCheckColumnLen(t *testing.T) {
	a := assert.New(t)

	col := &core.Column{
		GoType: core.StringType,
		Length: []int{-1},
	}
	a.NotError(checkColumnLen(col))

	col.Length[0] = 0
	a.Error(checkColumnLen(col))

	col.Length[0] = -2
	a.Error(checkColumnLen(col))

	col.GoType = core.IntType
	col.Length[0] = -2
	a.Error(checkColumnLen(col))

	col.Length[0] = -1
	a.Error(checkColumnLen(col))

	col.Length[0] = 0
	a.NotError(checkColumnLen(col))
}

func TestSetColumnNullable(t *testing.T) {
	a := assert.New(t)

	col := &core.Column{}

	a.False(col.Nullable)
	a.NotError(setColumnNullable(col, []string{})).True(col.Nullable)
	a.NotError(setColumnNullable(col, []string{"false"})).False(col.Nullable)
	a.NotError(setColumnNullable(col, []string{"T"})).True(col.Nullable)
	a.NotError(setColumnNullable(col, []string{"0"})).False(col.Nullable)

	a.Error(setColumnNullable(col, []string{"1", "2"}))
	a.Error(setColumnNullable(col, []string{"T1"}))

	ms := NewModels(nil)
	a.NotNil(ms)

	// 将 AI 设置为 nullable
	m, err := ms.New(&User{})
	a.NotError(err).NotNil(m)
	col.AI = true
	a.Error(setColumnNullable(col, []string{"true"}))
}
