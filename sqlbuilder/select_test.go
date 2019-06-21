// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"database/sql"
	"github.com/issue9/orm/internal/sqltest"

	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/v2"
	"github.com/issue9/orm/v2/dialect"
	"github.com/issue9/orm/v2/sqlbuilder"
)

var (
	_ sqlbuilder.SQLer       = &sqlbuilder.SelectStmt{}
	_ sqlbuilder.WhereStmter = &sqlbuilder.SelectStmt{}
)

func TestSelect_Query(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer clearDB(a, db)

	sql := sqlbuilder.Select(db, db.Dialect()).Select("*").
		From("#user").
		Where("id<?", 5).
		Desc("id")

	id, err := sql.QueryInt("id")
	a.NotError(err).
		Equal(id, 4)

	name, err := sql.QueryString("name")
	a.NotError(err).
		Equal(name, "4")

	obj := &user{}
	size, err := sql.QueryObject(true, obj)
	a.NotError(err).Equal(1, size)
	a.Equal(obj.ID, 4)

	cnt, err := sql.Count("count(*) AS cnt").QueryInt("cnt")
	a.NotError(err).
		Equal(cnt, 4)
}

func TestSelect_sqlite3(t *testing.T) {
	a := assert.New(t)
	e, err := orm.NewDB("sqlite3", "./test.db", "test_", dialect.Sqlite3())
	a.NotError(err)
	s := sqlbuilder.Select(e, e.Dialect()).Select("c1", "column2 as c2", "c3").
		From("table").
		And("c1=?", 1).
		Or("c2=@c2", sql.Named("c2", 2)).
		Limit(10, 0).
		Desc("c1")
	a.NotNil(s)
	query, args, err := s.SQL()
	a.NotError(err)
	a.Equal(args, []interface{}{1, sql.Named("c2", 2), 10, 0})
	sqltest.Equal(a, query, "select c1,column2 as c2,c3 from table where c1=? or c2=@c2 order by c1 desc limit ? offset ?")

	// count
	s.Count("count(*) as cnt")
	query, args, err = s.SQL()
	a.NotError(err)
	a.Equal(args, []interface{}{1, sql.Named("c2", 2)})
	sqltest.Equal(a, query, "select count(*) as cnt from table where c1=? or c2=@c2")

	// reset
	s.Reset()
	query, args, err = s.SQL()
	a.Error(err)

	s.Distinct().
		Select("c1,c2,c3").
		Join("left", "users as u", "a.id=u.id").
		Where("id=?", 5).
		Asc("id").
		From("table1 as t")
	query, args, err = s.SQL()
	a.NotError(err)
	a.Equal(args, []interface{}{5})
	sqltest.Equal(a, query, "select distinct c1,c2,c3 from table1 as t left join users as u on a.id=u.id where id=? order by id asc")

	// s.reset，没有  where
	s.Reset()
	s.Select("c1,c2").From("#tb1")
	query, args, err = s.SQL()
	a.NotError(err).Empty(args)
	sqltest.Equal(a, query, "select c1,c2 from #tb1")
}
