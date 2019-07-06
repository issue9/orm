// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/v2/internal/test"

	"github.com/issue9/orm/v2/internal/sqltest"
	"github.com/issue9/orm/v2/sqlbuilder"
)

var (
	_ sqlbuilder.DDLSQLer = &sqlbuilder.CreateIndexStmt{}
	_ sqlbuilder.DDLSQLer = &sqlbuilder.DropIndexStmt{}
)

func TestIndex_String(t *testing.T) {
	a := assert.New(t)

	a.Equal(sqlbuilder.IndexUnique.String(), "UNIQUE INDEX")
	a.Equal(sqlbuilder.IndexDefault.String(), "INDEX")
	a.Equal(sqlbuilder.Index(3).String(), "<unknown>")
	a.Equal(sqlbuilder.Index(-1).String(), "<unknown>")
}

func TestCreateIndex(t *testing.T) {
	a := assert.New(t)
	sql := sqlbuilder.CreateIndex(nil, nil)
	a.NotNil(sql)

	query, err := sql.Table("tbl1").
		Columns("c1", "c2").
		Name("c12").
		DDLSQL()
	a.NotError(err)
	sqltest.Equal(a, query[0], "create index c12 on tbl1(c1,c2)")

	sql.Reset()
	query, err = sql.DDLSQL()
	a.Error(err).Empty(query)

	sql = sqlbuilder.CreateIndex(nil, nil)
	query, err = sql.Table("tbl1").
		Columns("c1", "c2").
		Columns("c3", "c4").
		Type(sqlbuilder.IndexUnique).
		Name("c12").DDLSQL()
	a.NotError(err)
	sqltest.Equal(a, query[0], "create unique index c12 on tbl1(c1,c2,c3,c4)")

	// 缺少表名
	sql.Reset()
	query, err = sql.DDLSQL()
	a.Error(err).Empty(query)

	// 缺少列信息
	sql.Reset()
	sql.Table("tbl1")
	query, err = sql.DDLSQL()
	a.Error(err).Empty(query)
}

func TestIndex(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		initDB(t)
		defer clearDB(t)
		db := t.DB.DB
		dialect := t.DB.Dialect()

		err := sqlbuilder.CreateIndex(db, dialect).
			Table("users").
			Name("index_key").
			Columns("id", "name").
			Exec()
		t.NotError(err)

		err = sqlbuilder.DropIndex(db, dialect).
			Table("users").
			Name("index_key").
			Exec()
		t.NotError(err)
	})
}
