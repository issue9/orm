// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/core"
	"github.com/issue9/orm/v2/fetch"
)

// User 带自增和一个唯一约束
type User struct {
	ID       int    `orm:"name(id);ai;"`
	Username string `orm:"unique(unique_user_username);index(index_user_name);len(50)"`
	Password string `orm:"name(password);len(20)"`
	Regdate  int    `orm:"-"`
}

type Admin1 struct {
	ID       int    `orm:"name(id);ai;"`
	Username string `orm:"unique(unique_admin_username);index(index_admin_name);len(50)"`
	Password string `orm:"name(password);len(20)"`
	Regdate  int    `orm:"-"`
}

// Meta 指定表属性
func (m *User) Meta() string {
	return "mysql_engine(innodb);mysql_charset(utf8);name(users)"
}

// Admin 带自增和两个唯一约束
type Admin struct {
	Admin1
	Email string `orm:"name(email);len(20);unique(unique_admin_email)"`
	Group int64  `orm:"name(group);fk(fk_admin_name,#groups,id,NO ACTION)"`
}

// Meta 指定表属性
func (m *Admin) Meta() string {
	return "check(admin_chk_name,{group}>0);mysql_engine(innodb);mysql_charset(utf8);name(administrators)"
}

type obj struct {
	ID int `orm:"name(id);ai"`
}

// fun
func (m obj) Meta() string {
	return `name(objs)`
}

func TestModels_New(t *testing.T) {
	a := assert.New(t)
	ms := NewModels()
	a.NotNil(ms)

	m, err := ms.New(&Admin{})
	a.NotError(err).NotNil(m)

	// cols
	idCol := m.FindColumn("id") // 指定名称为小写
	a.NotNil(idCol)

	usernameCol := m.FindColumn("Username") // 未指定别名，与字段名相同
	a.NotNil(usernameCol).False(usernameCol.Nullable)

	// 通过 struct tag 过滤掉的列
	regdate := m.FindColumn("Regdate")
	a.Nil(regdate)

	groupCol := m.FindColumn("group")
	a.NotNil(groupCol)

	// index
	index, found := m.Indexes["index_admin_name"]
	a.True(found).Equal(usernameCol, index[0])

	// ai
	a.Equal(m.AI, idCol).
		Empty(m.PK) // 有自增，则主键为空

	// unique_name
	unique, found := m.Uniques["unique_admin_username"]
	a.True(found).Equal(unique[0], usernameCol)

	fk := m.FK[0]
	a.True(found).
		Equal(fk.Name, "fk_admin_name").
		Equal(fk.Column, groupCol).
		Equal(fk.RefTableName, "#groups").
		Equal(fk.RefColName, "id").
		Equal(fk.UpdateRule, "NO ACTION").
		Equal(fk.DeleteRule, "")

	// check
	chk, found := m.Checks["admin_chk_name"]
	a.True(found).Equal(chk, "{group}>0")

	// meta
	a.Equal(m.Meta, map[string][]string{
		"mysql_engine":  {"innodb"},
		"mysql_charset": {"utf8"},
	})

	// Meta返回的name属性
	a.Equal(m.Name, "administrators")

	// 多层指针下的 Receive
	o := obj{}
	m, err = ms.New(o)
	a.NotError(err).NotNil(m)
	a.Equal(m.Name, "objs")
	m, err = ms.New(&o)
	a.NotError(err).NotNil(m)
	a.Equal(m.Name, "objs")
	oo := &o
	m, err = ms.New(&oo)
	a.NotError(err).NotNil(m)
	a.Equal(m.Name, "objs")

	// 无效的 New
	m, err = ms.New(123)
	a.ErrorType(err, fetch.ErrInvalidKind).Nil(m)
}

func TestModel_sanitize(t *testing.T) {
	a := assert.New(t)

	ai := &Column{
		Column: &core.Column{
			GoType: core.IntType,
		},
	}

	pk1 := &Column{
		Column: &core.Column{
			GoType: core.IntType,
		},
	}

	pk2 := &Column{
		Column: &core.Column{
			GoType: core.IntType,
		},
	}

	nullable := &Column{
		Column: &core.Column{
			GoType:   core.IntType,
			Nullable: true,
		},
	}

	def := &Column{
		Column: &core.Column{
			GoType:     core.IntType,
			HasDefault: true,
			Default:    "1",
		},
	}

	m := &Model{
		Name:    "m1",
		Columns: []*Column{ai, pk1, pk2, nullable, def},
		AI:      ai,
	}

	a.NotError(m.sanitize())

	// AI 不能是 nullable
	m.PK = nil
	m.AI = nullable
	a.Error(m.sanitize())

	// AI 不能是 HasDefault=true
	m.AI = def
	a.Error(m.sanitize())

	// 多列主键约束
	m.AI = nil
	m.PK = []*Column{pk1, pk2}
	a.NotError(m.sanitize())

	// 多列主键约束，可以有 nullable 和 default
	m.AI = nil
	m.PK = []*Column{pk1, pk2, nullable, def}
	a.NotError(m.sanitize())

	// 单列主键，可以是 nullable
	m.AI = nil
	m.PK = []*Column{nullable}
	a.NotError(m.sanitize())

	// 单列主键，不能是 default
	m.AI = nil
	m.PK = []*Column{def}
	a.Error(m.sanitize())
}

func TestModel_parseColumn(t *testing.T) {
	a := assert.New(t)
	m := &Model{
		Columns: []*Column{},
	}

	// 不存在 struct tag，则以 col.Name 作为键名
	col := &Column{
		Column: &core.Column{
			Name: "xx",
		},
	}
	a.NotError(m.parseColumn(col, ""))
	a.Equal(col.Name, "xx")

	// name 值过多
	col = &Column{
		Column: &core.Column{},
	}
	a.Error(m.parseColumn(col, "name(m1,m2)"))

	// 不存在的属性名称
	col = &Column{
		Column: &core.Column{},
	}
	a.Error(m.parseColumn(col, "not-exists-property(p1)"))
}

func TestModel_parseMeta(t *testing.T) {
	a := assert.New(t)
	m := &Model{
		Checks: map[string]string{},
	}

	// 空值不算错误
	a.NotError(m.parseMeta(""))

	// name 属性过多
	a.Error(m.parseMeta("name(m1,m2)"))

	// check 属性过多或是过少
	a.Error(m.parseMeta("check(ck,id>0 AND id<10,error)"))

	// check 添加成功
	a.NotError(m.parseMeta("check(ck,id>0 AND id<10)"))

	// check 与已有 check 名称相同
	//a.Error(m.parseMeta("check(ck,id>0)"))

	// check 与其它约束名相同
	//m.Constraints = map[string]ConType{"fk": Fk}
	//a.Error(m.parseMeta("check(fk,id>0)"))
}

func TestModel_setOCC(t *testing.T) {
	a := assert.New(t)
	m := &Model{}
	col := &Column{
		Column: &core.Column{
			GoType: core.IntType,
		},
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
	a.NotError(m.setAI(col, nil))
	a.Error(m.setOCC(col, []string{"true"}))

	// 列有 nullable 属性
	m.OCC = nil
	m.AI = nil
	col.AI = false
	col.Nullable = true
	a.Error(m.setOCC(col, []string{"true"}))

	// 列属性不为数值型
	m.OCC = nil
	m.AI = nil
	col.Nullable = false
	col.GoType = core.StringType
	a.Error(m.setOCC(col, []string{"true"}))
}

func TestModel_setDefault(t *testing.T) {
	a := assert.New(t)
	m := &Model{}
	col := &Column{
		Column: &core.Column{},
	}

	// 未指定参数
	a.Error(m.setDefault(col, nil))

	// 过多的参数
	a.Error(m.setDefault(col, []string{"1", "2"}))

	// 正常
	a.NotError(m.setDefault(col, []string{"1"}))
	a.True(col.HasDefault).Equal(col.Default, "1")

	// 可以是主键的一部分
	m.PK = []*Column{col, col}
	a.NotError(m.setDefault(col, []string{"1"}))
	a.True(col.HasDefault).Equal(col.Default, "1")
}

func TestModel_setPK(t *testing.T) {
	a := assert.New(t)
	m := &Model{}
	col := &Column{
		Column: &core.Column{},
	}

	// 过多的参数
	a.Error(m.setPK(col, []string{"123"}))
}

func TestModel_setAI(t *testing.T) {
	a := assert.New(t)
	m := &Model{}

	col := &Column{
		Column: &core.Column{
			GoType:     core.StringType,
			HasDefault: true,
		},
	}

	// 太多的参数
	a.Error(m.setAI(col, []string{"true", "false"}))

	// 列类型只能是整数型
	col.GoType = core.Float32Type
	a.Error(m.setAI(col, nil))

	col.GoType = core.IntType
	a.NotError(m.setAI(col, nil))
}
