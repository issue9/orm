// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"database/sql"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/v2/internal/sqltest"
	"github.com/issue9/orm/v2/sqlbuilder"
)

type test struct {
	col     *sqlbuilder.Column
	err     bool
	SQLType string
}

type user struct {
	ID   int64  `orm:"name(id);ai"`
	Name string `orm:"name(name);index(i_user_name);len(20)"`
}

func (u *user) Meta() string {
	return "name(user)"
}

type model1 struct{}

func (m *model1) Meta() string {
	return "name(model1)"
}

type model2 struct{}

func (m *model2) Meta() string {
	return "check(chk_name,id>0);mysql_engine(innodb);mysql_charset(utf8);name(model2);sqlite3_rowid(false)"
}

func testData(a *assert.Assertion, d sqlbuilder.Dialect, data []*test) {
	for _, item := range data {
		typ, err := d.SQLType(item.col)
		if item.err {
			a.Error(err)
		} else {
			a.NotError(err)
		}
		sqltest.Equal(a, typ, item.SQLType)
	}
}

func TestMysqlLimitSQL(t *testing.T) {
	a := assert.New(t)

	query, ret := mysqlLimitSQL(5, 0)
	a.Equal(ret, []int{5, 0})
	sqltest.Equal(a, query, " LIMIT ? OFFSET ? ")

	query, ret = mysqlLimitSQL(5)
	a.Equal(ret, []int{5})
	sqltest.Equal(a, query, "LIMIT ?")

	// 带 sql.namedArg
	query, ret = mysqlLimitSQL(sql.Named("limit", 1), 2)
	a.Equal(ret, []interface{}{sql.Named("limit", 1), 2})
	sqltest.Equal(a, query, "LIMIT @limit offset ?")
}

func TestOracleLimitSQL(t *testing.T) {
	a := assert.New(t)

	query, ret := oracleLimitSQL(5, 0)
	a.Equal(ret, []int{0, 5})
	sqltest.Equal(a, query, " OFFSET ? ROWS FETCH NEXT ? ROWS ONLY ")

	query, ret = oracleLimitSQL(5)
	a.Equal(ret, []int{5})
	sqltest.Equal(a, query, "FETCH NEXT ? ROWS ONLY ")

	// 带 sql.namedArg
	query, ret = oracleLimitSQL(sql.Named("limit", 1), 2)
	a.Equal(ret, []interface{}{2, sql.Named("limit", 1)})
	sqltest.Equal(a, query, "offset ? rows fetch next @limit rows only")
}
