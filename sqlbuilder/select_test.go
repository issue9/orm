// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/fetch"
	"github.com/issue9/orm/v2/internal/test"
	"github.com/issue9/orm/v2/sqlbuilder"
)

var (
	_ sqlbuilder.SQLer       = &sqlbuilder.SelectStmt{}
	_ sqlbuilder.WhereStmter = &sqlbuilder.SelectStmt{}
)

func TestSelect(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		initDB(t)
		defer clearDB(t)

		db := t.DB.DB
		d := t.DB.Dialect()

		stmt := sqlbuilder.Select(db, d).Select("*").
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

		cnt, err := stmt.Count("count(*) AS cnt").QueryInt("cnt")
		a.NotError(err).
			Equal(cnt, 4)

		// 没有符合条件的数据
		stmt.Reset()
		stmt.Select("*").
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

	suite.ForEach(func(t *test.Test) {
		initDB(t)
		defer clearDB(t)
		db := t.DB.DB
		dialect := t.DB.Dialect()

		insert := sqlbuilder.Insert(db, dialect)
		r, err := insert.Table("info").
			Columns("uid", "nickname", "tel", "address").
			Values(1, "n1", "tel-1", "address-1").
			Values(1, "n2", "tel-2", "address-2").
			Exec()
		t.NotError(err).NotNil(r)

		sel := sqlbuilder.Select(db, dialect)
		rows, err := sel.Select("i.nickname,i.uid").
			From("users", "u").
			Where("uid=?", 1).
			Join("LEFT", "info AS i", "i.uid=u.id").
			Query()
		a.NotError(err).NotNil(rows)
		defer func() {
			t.NotError(rows.Close())
		}()
		maps, err := fetch.Map(false, rows)
		a.NotError(err).
			NotNil(maps).
			Equal(1, len(maps)).
			Equal(maps[0]["nickname"], "n1")
	})
}

func TestSelectStmt_Group(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		// TODO
	})
}
