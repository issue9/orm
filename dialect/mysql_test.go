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

var _ base = &Mysql{}

var m = &Mysql{}

func TestMysql_GetDBName(t *testing.T) {
	a := assert.New(t)

	a.Equal(m.GetDBName("root:password@/dbname"), "dbname")
	a.Equal(m.GetDBName("root:@/dbname"), "dbname")
	a.Equal(m.GetDBName("root:password@tcp(localhost:3066)/dbname"), "dbname")
	a.Equal(m.GetDBName("root:password@unix(/tmp/mysql.lock)/dbname?loc=Local"), "dbname")
	a.Equal(m.GetDBName("root:/"), "")
}

func TestMysql_SQLType(t *testing.T) {
	a := assert.New(t)
	buf := bytes.NewBufferString("")
	col := &core.Column{}
	a.Error(m.sqlType(buf, col))

	col.GoType = reflect.TypeOf(1)
	buf.Reset()
	a.NotError(m.sqlType(buf, col))
	chkSQLEqual(a, buf.String(), "BIGINT")

	col.Len1 = 5
	col.Len2 = 6
	buf.Reset()
	a.NotError(m.sqlType(buf, col))
	chkSQLEqual(a, buf.String(), "BIGINT(5)")

	col.GoType = reflect.TypeOf("abc")
	buf.Reset()
	a.NotError(m.sqlType(buf, col))
	chkSQLEqual(a, buf.String(), "VARCHAR(5)")

	col.GoType = reflect.TypeOf(1.2)
	buf.Reset()
	a.NotError(m.sqlType(buf, col))
	chkSQLEqual(a, buf.String(), "DOUBLE(5,6)")

	col.GoType = reflect.TypeOf([]byte{'1', '2'})
	buf.Reset()
	a.NotError(m.sqlType(buf, col))
	chkSQLEqual(a, buf.String(), "VARCHAR(5)")

	col.GoType = reflect.TypeOf(sql.NullInt64{})
	buf.Reset()
	a.NotError(m.sqlType(buf, col))
	chkSQLEqual(a, buf.String(), "BIGINT(5)")
}
