// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package core

import (
	"fmt"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/internal/modeltest"
)

func TestConType_String(t *testing.T) {
	a := assert.New(t)

	a.Equal("<none>", none.String()).
		Equal("KEY INDEX", fmt.Sprint(index)).
		Equal("UNIQUE INDEX", unique.String()).
		Equal("FOREIGN KEY", fk.String()).
		Equal("CHECK", check.String())

	var c1 conType
	a.Equal("<none>", c1.String())

	c1 = 100
	a.Equal("<unknown>", c1.String())
}

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

func TestModels(t *testing.T) {
	a := assert.New(t)

	ClearModels()
	a.Equal(0, len(models.items))

	m, err := NewModel(&modeltest.User{})
	a.NotError(err).
		NotNil(m).
		Equal(1, len(models.items))

	// 相同的 model 实例，不会增加数量
	m, err = NewModel(&modeltest.User{})
	a.NotError(err).
		NotNil(m).
		Equal(1, len(models.items))

	// 添加新的 model
	m, err = NewModel(&modeltest.Admin{})
	a.NotError(err).
		NotNil(m).
		Equal(2, len(models.items))

	ClearModels()
	a.Equal(0, len(models.items))
}

// 传递给 NewModel 是一个指针时的各种情况
func TestModel(t *testing.T) {
	ClearModels()
	a := assert.New(t)

	m, err := NewModel(&modeltest.Admin{})
	a.NotError(err).NotNil(m)

	// cols
	idCol, found := m.Cols["id"] // 指定名称为小写
	a.True(found)

	usernameCol, found := m.Cols["Username"] // 未指定别名，与字段名相同
	a.True(found).False(usernameCol.Nullable)

	// 通过 struct tag 过滤掉的列
	regdate, found := m.Cols["Regdate"]
	a.False(found).Nil(regdate)

	groupCol, found := m.Cols["group"]
	a.True(found)

	// index
	index, found := m.KeyIndexes["index_name"]
	a.True(found).Equal(usernameCol, index[0])

	// ai
	a.Equal(m.AI, idCol)

	// 主键应该和自增列相同
	a.NotNil(m.PK).Equal(m.PK[0], idCol)

	// unique_name
	unique, found := m.UniqueIndexes["unique_username"]
	a.True(found).Equal(unique[0], usernameCol)

	fk, found := m.FK["fk_name"]
	a.True(found).
		Equal(fk.Col, groupCol).
		Equal(fk.RefTableName, "#groups").
		Equal(fk.RefColName, "id").
		Equal(fk.UpdateRule, "NO ACTION").
		Equal(fk.DeleteRule, "")

	// check
	chk, found := m.Check["chk_name"]
	a.True(found).Equal(chk, "id>0")

	// meta
	a.Equal(m.Meta, map[string][]string{
		"engine":  []string{"innodb"},
		"charset": []string{"utf-8"},
	})

	// Meta返回的name属性
	a.Equal(m.Name, "administrators")
}

// BenchmarkNewModelNoCached	  100000	     23724 ns/op
func BenchmarkNewModelNoCached(b *testing.B) {
	ClearModels()
	a := assert.New(b)

	for i := 0; i < b.N; i++ {
		m, err := NewModel(&modeltest.User{})
		ClearModels()
		a.NotError(err).NotNil(m)
	}
}

// BenchmarkNewModelCached	 3000000	       480 ns/op
func BenchmarkNewModelCached(b *testing.B) {
	ClearModels()
	a := assert.New(b)

	for i := 0; i < b.N; i++ {
		m, err := NewModel(&modeltest.User{})
		a.NotError(err).NotNil(m)
	}
}
