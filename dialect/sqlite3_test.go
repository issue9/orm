// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/v2/internal/sqltest"
	"github.com/issue9/orm/v2/sqlbuilder"
)

func TestSqlite3_CreateTableOptions(t *testing.T) {
	a := assert.New(t)
	sql := sqlbuilder.New("")
	a.NotNil(sql)
	var s = &sqlite3{}

	// 空的 meta
	a.NotError(s.CreateTableOptionsSQL(sql, nil))
	a.Equal(sql.Len(), 0)

	// engine
	sql.Reset()
	a.NotError(s.CreateTableOptionsSQL(sql, map[string][]string{
		"sqlite3_rowid": []string{"false"},
	}))
	a.True(sql.Len() > 0)
	sqltest.Equal(a, sql.String(), "without rowid")
}

func TestSqlite3_SQLType(t *testing.T) {
	a := assert.New(t)
	var s = &sqlite3{}

	buf := sqlbuilder.New("")
	col := &sqlbuilder.Column{}

	// col == nil
	typ, err := s.SQLType(nil)
	a.ErrorType(err, errColIsNil).Empty(typ)

	// col.GoType == nil
	typ, err = s.SQLType(col)
	a.ErrorType(err, errGoTypeIsNil).Empty(typ)

	col.GoType = reflect.TypeOf(1)
	typ, err = s.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "INTEGER NOT NULL")

	col.Length = []int{5, 6}
	buf.Reset()
	typ, err = s.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "INTEGER NOT NULL")

	col.GoType = reflect.TypeOf("abc")
	buf.Reset()
	typ, err = s.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "TEXT NOT NULL")

	col.GoType = reflect.TypeOf(1.2)
	buf.Reset()
	typ, err = s.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "REAL NOT NULL")

	col.GoType = reflect.TypeOf(sql.NullInt64{})
	buf.Reset()
	typ, err = s.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "INTEGER NOT NULL")
}
