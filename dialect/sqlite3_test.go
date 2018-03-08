// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/internal/sqltest"
	"github.com/issue9/orm/model"
	"github.com/issue9/orm/sqlbuilder"
)

var _ base = &sqlite3{}

func TestSqlite3_sqlType(t *testing.T) {
	a := assert.New(t)
	var s = &sqlite3{}

	buf := sqlbuilder.New("")
	col := &model.Column{}
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
