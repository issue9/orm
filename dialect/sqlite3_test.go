// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"database/sql"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/sqltest"
	"github.com/issue9/orm/v2/sqlbuilder"

	_ "github.com/mattn/go-sqlite3"
)

var (
	_ sqlbuilder.TruncateTableStmtHooker  = &sqlite3{}
	_ sqlbuilder.DropColumnStmtHooker     = &sqlite3{}
	_ sqlbuilder.DropConstraintStmtHooker = &sqlite3{}
	_ sqlbuilder.AddConstraintStmtHooker  = &sqlite3{}
)

var sqlite3CreateTable = []string{`CREATE TABLE fk_table(
	id integer NOT NULL,
	name text not null,
	address text not null,
	constraint fk_table_pk PRIMARY KEY(id)
	)`,
	`CREATE TABLE usr (
	id integer NOT NULL,
	created integer NOT NULL,
	nickname text NOT NULL,
	state integer NOT NULL,
	username text NOT NULL,
	mobile text NOT NULL,
	email text NOT NULL,
	pwd text NOT NULL,
	CONSTRAINT users_pk PRIMARY KEY (id),
	CONSTRAINT u_user_xx1 UNIQUE (mobile,username),
	CONSTRAINT u_user_email1 UNIQUE (email,username),
	CONSTRAINT unique_id UNIQUE (id),
	CONSTRAINT xxx_fk FOREIGN KEY (id) REFERENCES fk_table (id),
	CONSTRAINT xxx CHECK (created > 0)
	)`,
	`create index index_user_mobile on usr(mobile)`,
	`create unique index index_user_unique_email_id on usr(email,id)`,
}

func TestSqlite3_AddDropConstraintStmtHook(t *testing.T) {
	a := assert.New(t)
	dbFile := "./orm_test.db"

	db, err := sql.Open("sqlite3", dbFile)
	a.NotError(err).NotNil(db)
	defer func() {
		a.NotError(db.Close())
		a.NotError(os.Remove(dbFile))
	}()

	for _, query := range sqlite3CreateTable {
		_, err = db.Exec(query)
		a.NotError(err)
	}

	s := &sqlite3{}

	// 不存在的约束，出错
	err = sqlbuilder.DropConstraint(db, s).
		Table("fk_table").
		Constraint("id_great_zero").
		Exec()
	a.Error(err)

	err = sqlbuilder.AddConstraint(db, s).
		Table("fk_table").
		Check("id_great_zero", "id>0").
		Exec()
	a.NotError(err)

	// 约束已经添加，可以正常删除
	err = sqlbuilder.DropConstraint(db, s).
		Table("fk_table").
		Constraint("id_great_zero").
		Exec()
	a.NotError(err)
}

func TestSqlite3_DropColumnStmtHook(t *testing.T) {
	a := assert.New(t)
	dbFile := "./orm_test.db"

	db, err := sql.Open("sqlite3", dbFile)
	a.NotError(err).NotNil(db)
	defer func() {
		a.NotError(db.Close())
		a.NotError(os.Remove(dbFile))
	}()

	for _, query := range sqlite3CreateTable {
		_, err = db.Exec(query)
		a.NotError(err)
	}

	s := &sqlite3{}
	err = sqlbuilder.DropColumn(db, s).
		Table("usr").
		Column("state").
		Exec()
	a.NotError(err)

	// 查询删除的列会出错
	_, err = db.Query("select state from usr")
	a.Error(err)
}

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

	testSQLType(a, Sqlite3(), data)
}
