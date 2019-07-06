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
	_ sqlbuilder.TruncateTableStmtHooker  = &sqlite3{}
	_ sqlbuilder.DropColumnStmtHooker     = &sqlite3{}
	_ sqlbuilder.DropConstraintStmtHooker = &sqlite3{}
	_ sqlbuilder.AddConstraintStmtHooker  = &sqlite3{}
)

func TestSqlite3_CreateTableOptions(t *testing.T) {
	a := assert.New(t)
	builder := sqlbuilder.New("")
	a.NotNil(builder)
	var s = &sqlite3{}

	// 空的 meta
	a.NotError(s.CreateTableOptionsSQL(builder, nil))
	a.Equal(builder.Len(), 0)

	// engine
	builder.Reset()
	a.NotError(s.CreateTableOptionsSQL(builder, map[string][]string{
		"sqlite3_rowid": {"false"},
	}))
	a.True(builder.Len() > 0)
	sqltest.Equal(a, builder.String(), "without rowid")

	builder.Reset()
	a.Error(s.CreateTableOptionsSQL(builder, map[string][]string{
		"sqlite3_rowid": {"false", "false"},
	}))
}

func TestSqlite3_SQLType(t *testing.T) {
	a := assert.New(t)

	var data = []*test{
		{ // col == nil
			err: true,
		},
		{ // col.GoType == nil
			col: &sqlbuilder.Column{GoType: nil},
			err: true,
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(1)},
			SQLType: "INTEGER NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.NullBool{})},
			SQLType: "INTEGER NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(false)},
			SQLType: "INTEGER NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf([]byte{'a', 'b'})},
			SQLType: "BLOB NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.NullInt64{})},
			SQLType: "INTEGER NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.NullFloat64{})},
			SQLType: "REAL NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.NullString{})},
			SQLType: "TEXT NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType:   reflect.TypeOf(sql.NullString{}),
				Nullable: true,
			},
			SQLType: "TEXT",
		},
		{
			col: &sqlbuilder.Column{
				GoType:  reflect.TypeOf(sql.NullString{}),
				Default: "123",
			},
			SQLType: "TEXT NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType:     reflect.TypeOf(sql.NullString{}),
				Default:    "123",
				HasDefault: true,
			},
			SQLType: "TEXT NOT NULL DEFAULT '123'",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(1),
				Length: []int{5, 6},
			},
			SQLType: "INTEGER NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(1),
				AI:     true,
			},
			SQLType: "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf("")},
			SQLType: "TEXT NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(1.2)},
			SQLType: "REAL NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.NullInt64{})},
			SQLType: "INTEGER NOT NULL",
		},

		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(time.Time{})},
			SQLType: "DATETIME NOT NULL",
		},

		{
			col: &sqlbuilder.Column{GoType: reflect.TypeOf(struct{}{})},
			err: true,
		},
	}

	testData(a, Sqlite3(), data)
}
