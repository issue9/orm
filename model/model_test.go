// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/internal/modeltest"
)

func TestModels(t *testing.T) {
	a := assert.New(t)

	ClearModels()
	a.Equal(0, len(models.items))

	m, err := New(&modeltest.User{})
	a.NotError(err).
		NotNil(m).
		Equal(1, len(models.items))

	// 相同的 model 实例，不会增加数量
	m, err = New(&modeltest.User{})
	a.NotError(err).
		NotNil(m).
		Equal(1, len(models.items))

	// 添加新的 model
	m, err = New(&modeltest.Admin{})
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

	m, err := New(&modeltest.Admin{})
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
