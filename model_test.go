// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/internal/modeltest"
)

func TestContType(t *testing.T) {
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

func TestNewModel(t *testing.T) {
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

func TestModel_setOCC(t *testing.T) {
	a := assert.New(t)
	m := &Model{}
	col := &Column{
		model:  m,
		GoType: reflect.TypeOf(123),
	}

	a.NotError(m.setOCC(col, nil))
	a.Equal(col, m.OCC)

	// m.OCC 已经存在
	a.Error(m.setOCC(col, nil))

	// occ(true)
	m.OCC = nil
	a.NotError(m.setOCC(col, []string{"true"}))

	// 太多的值，occ(true,123)
	m.OCC = nil
	a.Error(m.setOCC(col, []string{"true", "123"}))

	// 无法转换的值，occ("xx123")
	m.OCC = nil
	a.Error(m.setOCC(col, []string{"xx123"}))

	// 已经是 AI
	m.OCC = nil
	m.AI = col
	a.Error(m.setOCC(col, []string{"true"}))

	// 列有 nullable 属性
	m.OCC = nil
	m.AI = nil
	col.Nullable = true
	a.Error(m.setOCC(col, []string{"true"}))

	// 列属性不为数值型
	m.OCC = nil
	m.AI = nil
	col.Nullable = false
	col.GoType = reflect.TypeOf("string")
	a.Error(m.setOCC(col, []string{"true"}))
}

func TestModel_setDefault(t *testing.T) {
	a := assert.New(t)
	m := &Model{}
	col := &Column{
		model: m,
	}

	// 未指定参数
	a.Error(m.setDefault(col, nil))

	// 过多的参数
	a.Error(m.setDefault(col, []string{"1", "2"}))

	// 正常
	a.NotError(m.setDefault(col, []string{"1"}))
	a.True(col.HasDefault).Equal(col.Default, "1")

	// 不能同时是 AI
	m.AI = col
	a.Error(m.setDefault(col, []string{"1"}))

	// 不能与主键相同
	m.AI = nil
	m.PK = []*Column{col}
	a.Error(m.setDefault(col, []string{"1"}))

	// 可以是主键的一部分
	m.PK = []*Column{col, col}
	a.NotError(m.setDefault(col, []string{"1"}))
	a.True(col.HasDefault).Equal(col.Default, "1")
}

func TestModel_hasConstraint(t *testing.T) {
	a := assert.New(t)
	m := &Model{}

	a.Equal(m.hasConstraint("index", index), none)

	m.constraints = map[string]conType{"index": index}
	a.Equal(m.hasConstraint("index", index), none)
	a.Equal(m.hasConstraint("INDEX", fk), index)
}
