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
	_ sqlbuilder.TruncateTableStmtHooker = &sqlite3{}
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

	sql.Reset()
	a.Error(s.CreateTableOptionsSQL(sql, map[string][]string{
		"sqlite3_rowid": []string{"false", "false"},
	}))
}

func TestSqlite3_SQLType(t *testing.T) {
	a := assert.New(t)

	var data = []*test{
		&test{ // col == nil
			err: true,
		},
		&test{ // col.GoType == nil
			col: &sqlbuilder.Column{GoType: nil},
			err: true,
		},
		&test{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(1)},
			SQLType: "INTEGER NOT NULL",
		},
		&test{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.NullBool{})},
			SQLType: "INTEGER NOT NULL",
		},
		&test{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(false)},
			SQLType: "INTEGER NOT NULL",
		},
		&test{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf([]byte{'a', 'b'})},
			SQLType: "BLOB NOT NULL",
		},
		&test{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.NullInt64{})},
			SQLType: "INTEGER NOT NULL",
		},
		&test{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.NullFloat64{})},
			SQLType: "REAL NOT NULL",
		},
		&test{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.NullString{})},
			SQLType: "TEXT NOT NULL",
		},
		&test{
			col: &sqlbuilder.Column{
				GoType:   reflect.TypeOf(sql.NullString{}),
				Nullable: true,
			},
			SQLType: "TEXT",
		},
		&test{
			col: &sqlbuilder.Column{
				GoType:  reflect.TypeOf(sql.NullString{}),
				Default: "123",
			},
			SQLType: "TEXT NOT NULL",
		},
		&test{
			col: &sqlbuilder.Column{
				GoType:     reflect.TypeOf(sql.NullString{}),
				Default:    "123",
				HasDefault: true,
			},
			SQLType: "TEXT NOT NULL DEFAULT '123'",
		},
		&test{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(1),
				Length: []int{5, 6},
			},
			SQLType: "INTEGER NOT NULL",
		},
		&test{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(1),
				AI:     true,
			},
			SQLType: "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL",
		},
		&test{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf("")},
			SQLType: "TEXT NOT NULL",
		},
		&test{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(1.2)},
			SQLType: "REAL NOT NULL",
		},
		&test{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.NullInt64{})},
			SQLType: "INTEGER NOT NULL",
		},

		&test{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(time.Time{})},
			SQLType: "DATETIME NOT NULL",
		},

		&test{
			col: &sqlbuilder.Column{GoType: reflect.TypeOf(struct{}{})},
			err: true,
		},
	}

	testData(a, Sqlite3(), data)
}
