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

func TestPostgresHooks(t *testing.T) {
	a := assert.New(t)
	_, ok := dialect.Postgres().(sqlbuilder.TruncateTableStmtHooker)
	a.True(ok)
}

func TestPostgres_VersionSQL(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		testDialectVersionSQL(t)
	}, "postgres")
}

func TestPostgres_SQLType(t *testing.T) {
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
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(false)},
			SQLType: "BOOLEAN NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(1)},
			SQLType: "BIGINT NOT NULL",
		},

		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(int8(1)),
				AI:     true,
				Length: []int{5, 6},
			},
			SQLType: "SERIAL NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType:  reflect.TypeOf(int8(1)),
				Length:  []int{5, 6},
				Default: 1,
			},
			SQLType: "SMALLINT NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType:     reflect.TypeOf(int8(1)),
				Length:     []int{5, 6},
				HasDefault: true,
				Default:    1,
			},
			SQLType: "SMALLINT NOT NULL DEFAULT '1'",
		},

		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(sql.NullInt64{}),
				AI:     true,
			},
			SQLType: "BIGSERIAL NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.NullInt64{})},
			SQLType: "BIGINT NOT NULL",
		},

		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(uint32(1)),
				AI:     true,
				Length: []int{5, 6},
			},
			SQLType: "SERIAL NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(uint32(1)),
				Length: []int{5, 6},
			},
			SQLType: "INT NOT NULL",
		},

		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(1),
				Length: []int{5, 6},
			},
			SQLType: "BIGINT NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(1),
				AI:     true,
			},
			SQLType: "BIGSERIAL NOT NULL",
		},

		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(1.2),
				Length: []int{5, 9},
			},
			SQLType: "NUMERIC(5,9) NOT NULL",
		},
		{ // 长度必须为 2
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(1.2),
			},
			err: true,
		},
		{ // 长度必须为 2
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(1.2),
				Length: []int{1},
			},
			err: true,
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(sql.NullFloat64{}),
				Length: []int{5, 9},
			},
			SQLType: "NUMERIC(5,9) NOT NULL",
		},
		{ // 长度必须为 2
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(sql.NullFloat64{}),
			},
			err: true,
		},

		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf("")},
			SQLType: "TEXT NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf([]byte{'a', 'b'})},
			SQLType: "BYTEA NOT NULL",
		},

		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.NullString{})},
			SQLType: "TEXT NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(sql.NullString{}),
				Length: []int{-1, 111},
			},
			SQLType: "TEXT NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(sql.NullString{}),
				Length: []int{99, 111},
			},
			SQLType: "VARCHAR(99) NOT NULL",
		},

		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(""),
				Length: []int{99},
			},
			SQLType: "VARCHAR(99) NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(""),
				Length: []int{11111111},
			},
			SQLType: "TEXT NOT NULL",
		},

		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.RawBytes{})},
			SQLType: "BYTEA NOT NULL",
		},

		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(time.Time{})},
			SQLType: "TIMESTAMP NOT NULL",
		},

		{ // 无法转换的类型
			col: &sqlbuilder.Column{GoType: reflect.TypeOf(struct{}{})},
			err: true,
		},
	}

	testSQLType(a, dialect.Postgres(), data)
}

func TestPostgres_TruncateTableStmtHooker(t *testing.T) {
	a := assert.New(t)
	s := dialect.Postgres()

	hook, ok := s.(sqlbuilder.TruncateTableStmtHooker)
	a.True(ok).NotNil(hook)

	stmt := sqlbuilder.TruncateTable(nil, s).Table("tbl", "")
	a.NotNil(stmt)
	qs, err := hook.TruncateTableStmtHook(stmt)
	a.NotError(err).Equal(1, len(qs))
	sqltest.Equal(a, qs[0], `TRUNCATE TABLE "tbl"`)

	stmt = sqlbuilder.TruncateTable(nil, s).Table("tbl", "id")
	a.NotNil(stmt)
	qs, err = hook.TruncateTableStmtHook(stmt)
	a.NotError(err).Equal(1, len(qs))
	sqltest.Equal(a, qs[0], `TRUNCATE TABLE "tbl" RESTART IDENTITY`)
}

func TestPostgres_SQL(t *testing.T) {
	a := assert.New(t)
	p := dialect.Postgres()
	a.NotNil(p)

	eq := func(s1, s2 string) {
		ret, err := p.SQL(s1)
		a.NotError(err)
		a.Equal(ret, s2)
	}

	err := func(s1 string) {
		ret, err := p.SQL(s1)
		a.Error(err).Empty(ret)
	}

	eq("abc", "abc")
	eq("abc$", "abc$") // 未包含?的情况下，不会触发ReplaceMarks，可以有$
	eq("abc?abc", "abc$1abc")
	eq("?abc?abc?", "$1abc$2abc$3")
	eq("abc?abc?def", "abc$1abc$2def")
	eq("中文?abc?def", "中文$1abc$2def")

	err("$a?bc")
	err("?$abc$")
	err("a?bc$abc$abc")
	err("?中$文")
}

func BenchmarkPostgres_SQL(b *testing.B) {
	a := assert.New(b)
	p := dialect.Postgres()
	a.NotNil(p)

	s1 := "SELECT * FROM tbl WHERE uid>? AND group=? AND username LIKE ?"

	for i := 0; i < b.N; i++ {
		a.NotError(p.SQL(s1))
	}
}
