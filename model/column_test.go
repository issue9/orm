// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"testing"

	"github.com/issue9/assert"
)

func TestColumn_SetLen(t *testing.T) {
	a := assert.New(t)
	col := &Column{}

	a.NotError(col.setLen([]string{})).Equal(col.Len1, 0).Equal(col.Len2, 0)
	a.NotError(col.setLen([]string{"1", "2"})).Equal(col.Len1, 1).Equal(col.Len2, 2)
	a.Error(col.setLen([]string{"1", "2", "3"}))
	a.Error(col.setLen([]string{"1", "one"}))
}

func TestColumn_SetNullable(t *testing.T) {
	a := assert.New(t)

	col := &Column{}

	a.False(col.Nullable)
	a.NotError(col.setNullable([]string{})).True(col.Nullable)
	a.NotError(col.setNullable([]string{"false"})).False(col.Nullable)
	a.NotError(col.setNullable([]string{"T"})).True(col.Nullable)
	a.NotError(col.setNullable([]string{"0"})).False(col.Nullable)

	a.Error(col.setNullable([]string{"1", "2"}))
	a.Error(col.setNullable([]string{"T1"}))
}
