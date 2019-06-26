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
		"mysql_engine":  []string{"innodb"},
		"mysql_charset": []string{"utf8"},
	}))
	a.True(sql.Len() > 0)
	sqltest.Equal(a, sql.String(), "engine=innodb character set=utf8")
}

func TestMysql_SQLType(t *testing.T) {
	a := assert.New(t)
	col := &sqlbuilder.Column{}
	var m = &mysql{}

	// col == nil
	typ, err := m.SQLType(nil)
	a.ErrorType(err, errColIsNil).Empty(typ)

	// col.GoType == nil
	typ, err = m.SQLType(col)
	a.ErrorType(err, errGoTypeIsNil).Empty(typ)

	// int
	col.GoType = reflect.TypeOf(1)
	typ, err = m.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "BIGINT NOT NULL")

	// bool
	col.GoType = reflect.TypeOf(true)
	typ, err = m.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "BOOLEAN NOT NULL")

	// int8
	col.GoType = reflect.TypeOf(int8(1))
	typ, err = m.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "SMALLINT NOT NULL")

	// int with len
	col.GoType = reflect.TypeOf(int(1))
	col.Length = []int{5, 6}
	typ, err = m.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "BIGINT(5) NOT NULL")

	// string:abc
	col.GoType = reflect.TypeOf("abc")
	typ, err = m.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "VARCHAR(5) NOT NULL")

	// string len=-1
	col.Length = []int{-1}
	typ, err = m.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "LONGTEXT NOT NULL")

	// float
	col.GoType = reflect.TypeOf(1.2)
	col.Length = []int{5, 6}
	typ, err = m.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "DOUBLE(5,6) NOT NULL")

	// float not length
	col.Length = []int{5}
	typ, err = m.SQLType(col)
	a.ErrorType(err, errMissLength).Empty(typ)

	// NullInt64
	col.GoType = reflect.TypeOf(sql.NullInt64{})
	typ, err = m.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "BIGINT(5) NOT NULL")

	// int64 AI
	col.GoType = reflect.TypeOf(int(64))
	col.AI = true
	typ, err = m.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "BIGINT(5) PRIMARY KEY AUTO_INCREMENT NOT NULL")

	// uint64 pk
	col.GoType = reflect.TypeOf(uint(64))
	typ, err = m.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "BIGINT(5) UNSIGNED PRIMARY KEY AUTO_INCREMENT NOT NULL")

	// Error
	col.GoType = reflect.TypeOf(struct{}{})
	typ, err = m.SQLType(col)
	a.Error(err).Empty(typ)
}
