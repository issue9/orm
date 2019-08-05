// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v3/fetch"
	"github.com/issue9/orm/v3/internal/test"
	"github.com/issue9/orm/v3/sqlbuilder"
)

var (
	_ sqlbuilder.SQLer       = &sqlbuilder.SelectStmt{}
	_ sqlbuilder.WhereStmter = &sqlbuilder.SelectStmt{}
)

func TestSelect(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		stmt := sqlbuilder.Select(t.DB).Column("*").
			From("users").
			Where("id<?", 5).
			Desc("id")

		id, err := stmt.QueryInt("id")
		a.NotError(err).
			Equal(id, 4)

		f, err := stmt.QueryFloat("id")
		a.NotError(err).
			Equal(f, 4.0)

		// 不存在的列
		f, err = stmt.QueryFloat("id_not_exists")
		a.Error(err).Empty(f)

		name, err := stmt.QueryString("name")
		a.NotError(err).
			Equal(name, "4")

		obj := &user{}
		size, err := stmt.QueryObject(true, obj)
		a.NotError(err).Equal(1, size)
		a.Equal(obj.ID, 4)

		cnt, err := stmt.Count("count(*) as cnt").QueryInt("cnt")
		a.NotError(err).
			Equal(cnt, 4)

		// 没有符合条件的数据
		stmt.Reset()
		stmt.Column("*").
			From("users").
			Where("id<?", -100).
			Desc("id")
		id, err = stmt.QueryInt("id")
		a.ErrorType(err, sqlbuilder.ErrNoData)
	})
}

func TestSelectStmt_Join(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		insert := sqlbuilder.Insert(t.DB)
		r, err := insert.Table("info").
			Columns("uid", "nickname", "tel", "address").
			Values(1, "n1", "tel-1", "address-1").
			Values(1, "n2", "tel-2", "address-2").
			Exec()
		t.NotError(err).NotNil(r)

		sel := sqlbuilder.Select(t.DB)
		rows, err := sel.Columns("i.nickname", "i.uid").
			From("users", "u").
			Where("uid=?", 1).
			Join("LEFT", "info", "i", "i.uid=u.id").
			Query()
		a.NotError(err).NotNil(rows)
		defer func() {
			t.NotError(rows.Close())
		}()
		maps, err := fetch.Map(false, rows)
		a.NotError(err).
			NotNil(maps).
			Equal(2, len(maps)).
			Equal(maps[0]["nickname"], "n1").
			Equal(maps[1]["nickname"], "n2")
	})
}

func TestSelectStmt_Group(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		r, err := sqlbuilder.Update(t.DB).
			Table("users").
			Set("name", "2").
			Where("id>?", 1).
			Exec()
		a.NotError(err).NotNil(r)

		var list []*user
		cnt, err := sqlbuilder.Select(t.DB).
			Columns("sum(age) as {age}", "name").
			From("users").
			Group("name").
			QueryObject(true, &list)
		a.NotError(err).NotEmpty(cnt).Equal(2, len(list))
	})
}

func TestSelectStmt_Union(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		r, err := sqlbuilder.Insert(t.DB).Columns("uid", "tel", "nickname", "address").
			Values(1, "1", "1", "1").
			Values(2, "2", "2", "2").
			Table("info").
			Exec()
		t.NotError(err).NotNil(r)

		sel1 := sqlbuilder.Select(t.DB).
			Column("id").
			From("users").
			Where("id=?", 1)
		sel2 := sqlbuilder.Select(t.DB).
			Column("uid").
			From("info").
			Where("uid=?", 1)
		rows, err := sel1.Union(false, sel2).Query()
		t.NotError(err).NotNil(rows)
		defer func() {
			t.NotError(rows.Close())
		}()

		maps, err := fetch.Map(false, rows)
		t.NotError(err).NotNil(maps)
		t.Equal(1, len(maps)).
			Equal(maps[0]["id"], 1)
		_, found := maps[0]["uid"] // 名称跟随第一个 select
		t.False(found)

		// 添加了一个新的列名。导致长度不相同
		sel2.Column("name")
		rs, err := sel1.Query() // 不能命名为 rows，否则会影响上面 rows.Close 的执行
		a.ErrorType(err, sqlbuilder.ErrUnionColumnNotMatch).Nil(rs)
	})
}

func TestSelectStmt_UnionAll(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		r, err := sqlbuilder.Insert(t.DB).Columns("uid", "tel", "nickname", "address").
			Values(1, "1", "1", "1").
			Values(2, "2", "2", "2").
			Table("info").
			Exec()
		t.NotError(err).NotNil(r)

		sel1 := sqlbuilder.Select(t.DB).
			Column("id").
			From("users").
			Where("id=?", 1)
		sel2 := sqlbuilder.Select(t.DB).
			Column("uid").
			From("info").
			Where("uid=?", 1)
		rows, err := sel1.Union(true, sel2).Query()
		t.NotError(err).NotNil(rows)
		defer func() {
			t.NotError(rows.Close())
		}()

		maps, err := fetch.Map(false, rows)
		t.NotError(err).NotNil(maps)
		t.Equal(2, len(maps)).
			Equal(maps[0]["id"], 1).
			Equal(maps[1]["id"], 1)
		_, found := maps[0]["uid"] // 名称跟随第一个 select
		t.False(found)
	})
}
