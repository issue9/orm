// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/orm/internal/sqltest"

	"github.com/issue9/assert"
	"github.com/issue9/orm"
	"github.com/issue9/orm/dialect"
	"github.com/issue9/orm/sqlbuilder"
)

var _ sqlbuilder.SQLer = &sqlbuilder.TruncateStmt{}

func TestTruncate_sqlite3(t *testing.T) {
	a := assert.New(t)
	e, err := orm.NewDB("sqlite3", "./test.db", "test_", dialect.Sqlite3())
	a.NotError(err)

	sql := sqlbuilder.Truncate(e, e.Dialect())
	a.NotNil(sql)
	sql.Table("#tb1")
	query, args, err := sql.SQL()
	a.NotError(err).Empty(args)
	sqltest.Equal(a, query, "delete from #tb1;delete from SQLITE_SEQUENCE WHERE name='#tb1';")

	sql.Reset()
	sql.Table("#tb1").Table("#tb2").AI("c1")
	query, args, err = sql.SQL()
	a.NotError(err).Empty(args)
	sqltest.Equal(a, query, "delete from #tb2;delete from SQLITE_SEQUENCE WHERE name='#tb2';")
}
