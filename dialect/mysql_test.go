// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/core"
	"github.com/issue9/orm/internal/sqltest"
)

var _ base = &mysql{}

var m = &mysql{}

func TestMysql_SQLType(t *testing.T) {
	a := assert.New(t)
	buf := core.NewStringBuilder("")
	col := &core.Column{}

	// col == nil
	a.Error(m.sqlType(buf, nil))

	// col.GoType == nil
	a.Error(m.sqlType(buf, col))

	// int
	col.GoType = reflect.TypeOf(1)
	buf.Reset()
	a.NotError(m.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "BIGINT")

	// int with len
	col.Len1 = 5
	col.Len2 = 6
	buf.Reset()
	a.NotError(m.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "BIGINT(5)")

	// string:abc
	col.GoType = reflect.TypeOf("abc")
	buf.Reset()
	a.NotError(m.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "VARCHAR(5)")

	// float
	col.GoType = reflect.TypeOf(1.2)
	buf.Reset()
	a.NotError(m.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "DOUBLE(5,6)")

	// []byte with len
	col.GoType = reflect.TypeOf([]byte{'1', '2'})
	buf.Reset()
	a.NotError(m.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "VARCHAR(5)")

	// NullInt64
	col.GoType = reflect.TypeOf(sql.NullInt64{})
	buf.Reset()
	a.NotError(m.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "BIGINT(5)")
}
