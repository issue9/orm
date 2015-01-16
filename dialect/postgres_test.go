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

var _ base = &Postgres{}

var p = &Postgres{}

func TestPostgres_GetDBName(t *testing.T) {
	a := assert.New(t)

	a.Equal(p.GetDBName("user=abc dbname = dbname password=abc"), "dbname")
	a.Equal(p.GetDBName("dbname=\tdbname user=abc"), "dbname")
	a.Equal(p.GetDBName("dbname=dbname\tuser=abc"), "dbname")
	a.Equal(p.GetDBName("\tdbname=dbname user=abc"), "dbname")
	a.Equal(p.GetDBName("\tdbname = dbname user=abc"), "dbname")
}

func TestPostgres_SQLType(t *testing.T) {
	a := assert.New(t)
	buf := bytes.NewBufferString("")
	col := &core.Column{}
	a.Error(p.sqlType(buf, col))

	col.GoType = reflect.TypeOf(1)
	buf.Reset()
	a.NotError(p.sqlType(buf, col))
	chkSQLEqual(a, buf.String(), "BIGINT")

	col.Len1 = 5
	col.Len2 = 6
	buf.Reset()
	a.NotError(p.sqlType(buf, col))
	chkSQLEqual(a, buf.String(), "BIGINT")

	col.GoType = reflect.TypeOf("abc")
	buf.Reset()
	a.NotError(p.sqlType(buf, col))
	chkSQLEqual(a, buf.String(), "VARCHAR(5)")

	col.GoType = reflect.TypeOf(1.2)
	buf.Reset()
	a.NotError(p.sqlType(buf, col))
	chkSQLEqual(a, buf.String(), "DOUBLE(5,6)")

	col.GoType = reflect.TypeOf([]byte{'1', '2'})
	buf.Reset()
	a.NotError(p.sqlType(buf, col))
	chkSQLEqual(a, buf.String(), "VARCHAR(5)")

	col.GoType = reflect.TypeOf(sql.NullInt64{})
	buf.Reset()
	a.NotError(p.sqlType(buf, col))
	chkSQLEqual(a, buf.String(), "BIGINT")
}
