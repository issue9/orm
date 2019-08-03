// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"reflect"
	"testing"
	"time"

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

func TestModel_setDefault(t *testing.T) {
	a := assert.New(t)
	m := core.NewModel(core.Table, "m1", 10)

	col, err := core.NewColumnFromGoType(core.IntType)
	a.NotError(err).NotNil(col)
	col.Name = "def"
	a.NotError(m.AddColumn(col))

	// 未指定参数
	a.Error(setDefault(col, nil))

	// 过多的参数
	a.Error(setDefault(col, []string{"1", "2"}))

	// 正常
	a.NotError(setDefault(col, []string{"1"}))
	a.True(col.HasDefault).
		Equal(col.Default, 1)

	// 可以是主键的一部分
	m.PrimaryKey = []*core.Column{col, col}
	a.NotError(setDefault(col, []string{"1"}))
	a.True(col.HasDefault).
		Equal(col.Default, 1)

	col, err = core.NewColumnFromGoType(reflect.TypeOf(&last{}))
	a.NotError(err).NotNil(col)

	// 格式不正确
	a.Error(setDefault(col, []string{"1"}))

	// 格式正确
	now := "2019-07-29T00:38:59+08:00"
	tt, err := time.Parse(time.RFC3339, now)
	a.NotError(err).NotEmpty(tt)
	a.NotError(setDefault(col, []string{"192.168.1.1," + now}))
	a.Equal(col.Default, &last{
		IP:      "192.168.1.1",
		Created: tt.Unix(),
	})

	col, err = core.NewColumnFromGoType(reflect.TypeOf(time.Time{}))
	a.NotError(err).NotNil(col)

	// 格式不正确
	a.Error(setDefault(col, []string{"1"}))

	// 格式正确
	a.NotError(setDefault(col, []string{now}))
	a.Equal(col.Default, tt)
}
