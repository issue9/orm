// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/v2"
	"github.com/issue9/orm/v2/internal/sqltest"
	"github.com/issue9/orm/v2/sqlbuilder"
)

var _ base = &mysql{}

func TestMysql_CreateTableOptions(t *testing.T) {
	a := assert.New(t)
	sql := sqlbuilder.New("")
	a.NotNil(sql)
	var m = &mysql{}

	// 空的 meta
	a.NotError(m.CreateTableOptionsSQL(sql, nil))
	a.Equal(sql.Len(), 0)

	// engine
	sql.Reset()
	a.NotError(m.CreateTableOptionsSQL(sql, map[string][]string{
		"mysql_engine":    []string{"innodb"},
		"mysql_character": []string{"utf8"},
	}))
	a.True(sql.Len() > 0)
	sqltest.Equal(a, sql.String(), "engine=innodb character set=utf8")
}

func TestMysql_sqlType(t *testing.T) {
	a := assert.New(t)
	buf := sqlbuilder.New("")
	col := &orm.Column{}
	var m = &mysql{}

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

	// NullInt64
	col.GoType = reflect.TypeOf(sql.NullInt64{})
	buf.Reset()
	a.NotError(m.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "BIGINT(5)")
}
