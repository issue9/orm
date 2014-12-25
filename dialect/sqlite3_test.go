// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"bytes"
	"database/sql"
	"reflect"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/core"
)

var _ base = &Sqlite3{}

var s = &Sqlite3{}

func TestSqlite3GetDBName(t *testing.T) {
	a := assert.New(t)

	a.Equal(s.GetDBName("./dbname.db"), "dbname")
	a.Equal(s.GetDBName("./dbname"), "dbname")
	a.Equal(s.GetDBName("abc/dbname.abc"), "dbname")
	a.Equal(s.GetDBName("dbname"), "dbname")
	a.Equal(s.GetDBName(""), "")
}

func TestSqlite3SQLType(t *testing.T) {
	a := assert.New(t)
	buf := bytes.NewBufferString("")
	col := &core.Column{}
	a.Error(s.sqlType(buf, col))

	col.GoType = reflect.TypeOf(1)
	buf.Reset()
	a.NotError(s.sqlType(buf, col))
	a.StringEqual(buf.String(), "INTEGER", style)

	col.Len1 = 5
	col.Len2 = 6
	buf.Reset()
	a.NotError(s.sqlType(buf, col))
	a.StringEqual(buf.String(), "INTEGER", style)

	col.GoType = reflect.TypeOf("abc")
	buf.Reset()
	a.NotError(s.sqlType(buf, col))
	a.StringEqual(buf.String(), "TEXT", style)

	col.GoType = reflect.TypeOf(1.2)
	buf.Reset()
	a.NotError(s.sqlType(buf, col))
	a.StringEqual(buf.String(), "REAL", style)

	col.GoType = reflect.TypeOf([]byte{'1', '2'})
	buf.Reset()
	a.NotError(s.sqlType(buf, col))
	a.StringEqual(buf.String(), "TEXT", style)

	col.GoType = reflect.TypeOf(sql.NullInt64{})
	buf.Reset()
	a.NotError(s.sqlType(buf, col))
	a.StringEqual(buf.String(), "INTEGER", style)
}
