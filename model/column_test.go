// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"reflect"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/sqlbuilder"
)

func TestColumn_SetLen(t *testing.T) {
	a := assert.New(t)
	col := &Column{
		Column: &sqlbuilder.Column{},
	}

	a.NotError(col.setLen([]string{})).Empty(col.Length)
	a.NotError(col.setLen([]string{"1", "2"})).
		Equal(col.Length[0], 1).
		Equal(col.Length[1], 2)
	a.Error(col.setLen([]string{"1", "2", "3"}))
	a.Error(col.setLen([]string{"1", "one"}))
	a.Error(col.setLen([]string{"one", "one"}))
}

func TestColumn_IsZero(t *testing.T) {
	a := assert.New(t)
	col := &Column{
		Column: &sqlbuilder.Column{},
	}

	col.GoType = sqlbuilder.IntType
	col.zero = reflect.Zero(col.GoType).Interface()
	a.True(col.IsZero(reflect.ValueOf(int(0))))
	a.False(col.IsZero(reflect.ValueOf(1)))

	col.GoType = reflect.TypeOf([]byte{})
	col.zero = reflect.Zero(col.GoType).Interface()
	a.True(col.IsZero(reflect.ValueOf([]byte{})))
	a.True(col.IsZero(reflect.ValueOf([]byte(""))))
	a.False(col.IsZero(reflect.ValueOf([]byte{'0'})))

	col.GoType = sqlbuilder.RawBytesType
	col.zero = reflect.Zero(col.GoType).Interface()
	a.True(col.IsZero(reflect.ValueOf([]byte{})))
	a.True(col.IsZero(reflect.ValueOf([]byte(""))))
	a.False(col.IsZero(reflect.ValueOf([]byte{'0'})))
}

func TestColumn_checkLen(t *testing.T) {
	a := assert.New(t)

	col := &Column{
		Column: &sqlbuilder.Column{
			GoType: sqlbuilder.StringType,
			Length: []int{-1},
		},
	}
	a.NotError(col.checkLen())

	col.Length[0] = 0
	a.Error(col.checkLen())

	col.Length[0] = -2
	a.Error(col.checkLen())

	col.GoType = sqlbuilder.IntType
	col.Length[0] = -2
	a.Error(col.checkLen())

	col.Length[0] = -1
	a.Error(col.checkLen())

	col.Length[0] = 0
	a.NotError(col.checkLen())
}

func TestColumn_SetNullable(t *testing.T) {
	a := assert.New(t)

	col := &Column{
		Column: &sqlbuilder.Column{},
	}

	a.False(col.Nullable)
	a.NotError(col.setNullable([]string{})).True(col.Nullable)
	a.NotError(col.setNullable([]string{"false"})).False(col.Nullable)
	a.NotError(col.setNullable([]string{"T"})).True(col.Nullable)
	a.NotError(col.setNullable([]string{"0"})).False(col.Nullable)

	a.Error(col.setNullable([]string{"1", "2"}))
	a.Error(col.setNullable([]string{"T1"}))

	ms := NewModels()
	a.NotNil(ms)

	// 将 AI 设置为 nullable
	m, err := ms.New(&User{})
	a.NotError(err).NotNil(m)
	col.AI = true
	a.Error(col.setNullable([]string{"true"}))
}
