// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect_test

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/dialect"
	"github.com/issue9/orm/v2/internal/sqltest"
	"github.com/issue9/orm/v2/internal/test"
	"github.com/issue9/orm/v2/sqlbuilder"
)

// 创建测试数据表的脚本
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
	CONSTRAINT usr_pk PRIMARY KEY (id),
	CONSTRAINT u_user_xx1 UNIQUE (mobile,username),
	CONSTRAINT u_user_email1 UNIQUE (email,username),
	CONSTRAINT unique_id UNIQUE (id),
	CONSTRAINT xxx_fk FOREIGN KEY (id) REFERENCES fk_table (id),
	CONSTRAINT xxx CHECK (created > 0)
	)`,
	`create index index_user_mobile on usr(mobile)`,
	`create unique index index_user_unique_email_id on usr(email,id)`,
}

func TestSqliteHooks(t *testing.T) {
	a := assert.New(t)
	s := dialect.Sqlite3()

	_, ok := s.(sqlbuilder.TruncateTableStmtHooker)
	a.True(ok)

	_, ok = s.(sqlbuilder.DropColumnStmtHooker)
	a.True(ok)

	_, ok = s.(sqlbuilder.DropConstraintStmtHooker)
	a.True(ok)

	_, ok = s.(sqlbuilder.AddConstraintStmtHooker)
	a.True(ok)
}

func TestSqlite3_VersionSQL(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		testDialectVersionSQL(t)
	}, "sqlite3")
}

func TestSqlite3_AddConstraintStmtHook(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		db := t.DB.DB

		for _, query := range sqlite3CreateTable {
			_, err := db.Exec(query)
			t.NotError(err)
		}

		// check 约束
		err := sqlbuilder.AddConstraint(db, t.DB.Dialect()).
			Table("fk_table").
			Check("id_great_zero", "id>0").
			Exec()
		t.NotError(err)

	}, "sqlite3")
}

func TestSqlite3_DropConstraintStmtHook(t *testing.T) {
	a := assert.New(t)

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		db := t.DB.DB

		for _, query := range sqlite3CreateTable {
			_, err := db.Exec(query)
			t.NotError(err)
		}

		testDialectDropConstraintStmtHook(t)
	}, "sqlite3")
}

func TestSqlite3_DropColumnStmtHook(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		db := t.DB.DB

		for _, query := range sqlite3CreateTable {
			_, err := db.Exec(query)
			t.NotError(err)
		}

		err := sqlbuilder.DropColumn(db, t.DB.Dialect()).
			Table("usr").
			Column("state").
			Exec()
		t.NotError(err)

		// 查询删除的列会出错
		_, err = db.Query("SELECT state FROM usr")
		t.Error(err)
	}, "sqlite3")
}

func TestSqlite3_CreateTableOptions(t *testing.T) {
	a := assert.New(t)
	builder := sqlbuilder.New("")
	a.NotNil(builder)
	var s = dialect.Sqlite3()

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

func TestSqlite3_TruncateTableStmtHooker(t *testing.T) {
	a := assert.New(t)
	s := dialect.Sqlite3()

	hook, ok := s.(sqlbuilder.TruncateTableStmtHooker)
	a.True(ok).NotNil(hook)

	stmt := sqlbuilder.TruncateTable(nil, s).Table("tbl", "")
	a.NotNil(stmt)
	qs, err := hook.TruncateTableStmtHook(stmt)
	a.NotError(err).Equal(1, len(qs))
	sqltest.Equal(a, qs[0], "DELETE FROM `tbl`")

	stmt = sqlbuilder.TruncateTable(nil, s).Table("tbl", "id")
	a.NotNil(stmt)
	qs, err = hook.TruncateTableStmtHook(stmt)
	a.NotError(err).Equal(2, len(qs))
	sqltest.Equal(a, qs[0], "DELETE FROM `tbl`")
	sqltest.Equal(a, qs[1], "DELETE FROM SQLITE_SEQUENCE WHERE name='tbl'")
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

	testSQLType(a, dialect.Sqlite3(), data)
}
