// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/sqltest"
	"github.com/issue9/orm/v2/sqlbuilder"
)

var (
	_ sqlbuilder.TruncateTableStmtHooker  = &mysql{}
	_ sqlbuilder.DropIndexStmtHooker      = &mysql{}
	_ sqlbuilder.DropConstraintStmtHooker = &mysql{}
)

func TestMysql_CreateTableOptions(t *testing.T) {
	a := assert.New(t)
	builder := sqlbuilder.New("")
	a.NotNil(builder)
	var m = Mysql()

	// 空的 meta
	a.NotError(m.CreateTableOptionsSQL(builder, nil))
	a.Equal(builder.Len(), 0)

	// engine
	builder.Reset()
	a.NotError(m.CreateTableOptionsSQL(builder, map[string][]string{
		"mysql_engine":  {"innodb"},
		"mysql_charset": {"utf8"},
	}))
	a.True(builder.Len() > 0)
	sqltest.Equal(a, builder.String(), "engine=innodb character set=utf8")
}

func TestMysql_SQLType(t *testing.T) {
	a := assert.New(t)

	var data = []*sqltypeTester{
		{ // col == nil
			err: true,
		},
		{ // col.GoType == nil
			col: &sqlbuilder.Column{GoType: nil},
			err: true,
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(1)},
			SQLType: "BIGINT NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType:  reflect.TypeOf(1),
				Default: 5,
			},
			SQLType: "BIGINT NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType:     reflect.TypeOf(1),
				HasDefault: true,
				Default:    5,
			},
			SQLType: "BIGINT NOT NULL DEFAULT '5'",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(true)},
			SQLType: "BOOLEAN NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(time.Time{})},
			SQLType: "DATETIME NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(uint16(16))},
			SQLType: "MEDIUMINT UNSIGNED NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(int8(1))},
			SQLType: "SMALLINT NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf([]byte{'a', 'b', 'c'})},
			SQLType: "BLOB NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(int(1)),
				Length: []int{5, 6},
			},
			SQLType: "BIGINT(5) NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(""),
				Length: []int{5, 6},
			},
			SQLType: "VARCHAR(5) NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(""),
				Length: []int{-1},
			},
			SQLType: "LONGTEXT NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(1.2),
				Length: []int{5, 6},
			},
			SQLType: "DOUBLE(5,6) NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(1.2),
				Length: []int{5},
			},
			err: true,
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(sql.NullFloat64{}),
				Length: []int{5},
			},
			err: true,
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(sql.NullFloat64{}),
				Length: []int{5, 7},
			},
			SQLType: "DOUBLE(5,7) NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(sql.NullInt64{}),
				Length: []int{5},
			},
			SQLType: "BIGINT(5) NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(sql.NullString{}),
				Length: []int{5},
			},
			SQLType: "VARCHAR(5) NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.NullString{})},
			SQLType: "LONGTEXT NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.NullBool{})},
			SQLType: "BOOLEAN NOT NULL",
		},
		{ // sql.RawBytes 会被转换成 []byte
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.RawBytes{})},
			SQLType: "BLOB NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(int64(1)),
				AI:     true,
			},
			SQLType: "BIGINT PRIMARY KEY AUTO_INCREMENT NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(uint64(1)),
				AI:     true,
			},
			SQLType: "BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT NOT NULL",
		},
		{
			col: &sqlbuilder.Column{GoType: reflect.TypeOf(struct{}{})},
			err: true,
		},
	}

	testSQLType(a, Mysql(), data)
}
