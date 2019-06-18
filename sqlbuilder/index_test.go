// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/sqltest"
	"github.com/issue9/orm/v2/sqlbuilder"
)

var _ sqlbuilder.SQLer = &sqlbuilder.CreateIndexStmt{}

func TestCreateIndex(t *testing.T) {
	a := assert.New(t)
	sql := sqlbuilder.CreateIndex(nil)
	a.NotNil(sql)

	query, args, err := sql.Table("tbl1").Columns("c1", "c2").Name("c12").SQL()
	a.NotError(err).Nil(args)
	sqltest.Equal(a, query, "create index c12 on tbl1(c1,c2)")

	// 重置
	sql.Reset()
	query, args, err = sql.SQL()
	a.Error(err).Nil(args).Empty(query)

	sql = sqlbuilder.CreateIndex(nil)
	query, args, err = sql.Table("tbl1").Columns("c1", "c2").Type(sqlbuilder.IndexUnique).Name("c12").SQL()
	a.NotError(err).Nil(args)
	sqltest.Equal(a, query, "create unique index c12 on tbl1(c1,c2)")

	// 重置
	sql.Reset()
	query, args, err = sql.SQL()
	a.Error(err).Nil(args).Empty(query)
}

func TestIndex(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer clearDB(a, db)

	_, err := sqlbuilder.CreateIndex(db).
		Table("#user").
		Name("index_key").
		Columns("id", "name").
		Exec()
	a.NotError(err)

	_, err = sqlbuilder.DropIndex(db, db.Dialect()).
		Table("#user").
		Name("index_key").
		Exec()
	a.NotError(err)
}
