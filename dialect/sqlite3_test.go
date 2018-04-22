// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm"
	"github.com/issue9/orm/internal/sqltest"
	"github.com/issue9/orm/sqlbuilder"
)

var _ base = &sqlite3{}

func TestSqlite3_CreateTableOptions(t *testing.T) {
	a := assert.New(t)
	sql := sqlbuilder.New("")
	a.NotNil(sql)
	var s = &sqlite3{}

	// 空的 meta
	mod, err := orm.NewModel(&model1{})
	a.NotError(err).NotNil(mod)
	s.createTableOptions(sql, mod)
	a.Equal(sql.Len(), 0)

	// engine
	sql.Reset()
	mod, err = orm.NewModel(&model2{})
	a.NotError(err).NotNil(mod)
	s.createTableOptions(sql, mod)
	a.True(sql.Len() > 0)
	sqltest.Equal(a, sql.String(), "without rowid")
}

func TestSqlite3_sqlType(t *testing.T) {
	a := assert.New(t)
	var s = &sqlite3{}

	buf := sqlbuilder.New("")
	col := &orm.Column{}
	a.Error(s.sqlType(buf, col))

	col.GoType = reflect.TypeOf(1)
	buf.Reset()
	a.NotError(s.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "INTEGER")

	col.Len1 = 5
	col.Len2 = 6
	buf.Reset()
	a.NotError(s.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "INTEGER")

	col.GoType = reflect.TypeOf("abc")
	buf.Reset()
	a.NotError(s.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "TEXT")

	col.GoType = reflect.TypeOf(1.2)
	buf.Reset()
	a.NotError(s.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "REAL")

	col.GoType = reflect.TypeOf([]byte{'1', '2'})
	buf.Reset()
	a.NotError(s.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "TEXT")

	col.GoType = reflect.TypeOf(sql.NullInt64{})
	buf.Reset()
	a.NotError(s.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "INTEGER")
}
