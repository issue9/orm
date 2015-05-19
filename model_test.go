// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"fmt"
	"testing"

	"github.com/issue9/assert"
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

type modelGroup struct {
	Group int `orm:"name(group);fk(fk_name,table.group,id,NO ACTION,)"`
}

type modelUser struct {
	modelGroup

	Id       int    `orm:"name(id);ai;"`
	Email    string `orm:"unique(unique_name);index(index_name);nullable;"`
	Username string `orm:"index(index_name);len(50)"`

	Regdate int `orm:"-"`
}

func (m *modelUser) Meta() string {
	return "check(chk_name,id>5);engine(innodb);charset(utf-8);name(user)"
}

func TestModels(t *testing.T) {
	a := assert.New(t)

	ClearModels()
	a.Equal(0, len(models.items))

	m, err := NewModel(&modelUser{})
	a.NotError(err).
		NotNil(m).
		Equal(1, len(models.items))

	// 相同的model实例，不会增加数量
	m, err = NewModel(&modelUser{})
	a.NotError(err).
		NotNil(m).
		Equal(1, len(models.items))

	// 添加新的model
	m, err = NewModel(&modelGroup{})
	a.NotError(err).
		NotNil(m).
		Equal(2, len(models.items))

	ClearModels()
	a.Equal(0, len(models.items))
}

// 传递给NewModel是一个指针时的各种情况
func TestModel(t *testing.T) {
	ClearModels()
	a := assert.New(t)

	// todo 正确声明第二个参数！！
	m, err := NewModel(&modelUser{})
	a.NotError(err).NotNil(m)

	// cols
	idCol, found := m.Cols["id"] // 指定名称为小写
	a.True(found)

	emailCol, found := m.Cols["Email"] // 未指定别名，与字段名相同
	a.True(found).True(emailCol.Nullable, "emailCol.Nullable==false")

	usernameCol, found := m.Cols["Username"]
	a.True(found)

	groupCol, found := m.Cols["group"]
	a.True(found)

	// 通过struct tag过滤掉的列
	regdate, found := m.Cols["Regdate"]
	a.False(found).Nil(regdate)

	// index
	index, found := m.KeyIndexes["index_name"]
	a.True(found).
		Equal(emailCol, index[0]).
		Equal(usernameCol, index[1])

	// ai
	a.Equal(m.AI, idCol)

	// 主键应该和自增列相同
	a.NotNil(m.PK).Equal(m.PK[0], idCol)

	// unique_name
	unique, found := m.UniqueIndexes["unique_name"]
	a.True(found).Equal(unique[0], emailCol)

	fk, found := m.FK["fk_name"]
	a.True(found).
		Equal(fk.Col, groupCol).
		Equal(fk.RefTableName, "table.group").
		Equal(fk.RefColName, "id").
		Equal(fk.UpdateRule, "NO ACTION").
		Equal(fk.DeleteRule, "")

	// check
	chk, found := m.Check["chk_name"]
	a.True(found).Equal(chk, "id>5")

	// meta
	a.Equal(m.Meta, map[string][]string{
		"engine":  []string{"innodb"},
		"charset": []string{"utf-8"},
	})

	// Meta返回的name属性
	a.Equal(m.Name, "user")
}
