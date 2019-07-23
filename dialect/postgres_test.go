// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect_test

import (
	"reflect"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/core"
	"github.com/issue9/orm/v2/dialect"
	"github.com/issue9/orm/v2/internal/sqltest"
	"github.com/issue9/orm/v2/internal/test"
	"github.com/issue9/orm/v2/sqlbuilder"
)

func TestPostgres_VersionSQL(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
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
			col: &core.Column{GoType: nil},
			err: true,
		},
		{
			col:     &core.Column{GoType: core.BoolType},
			SQLType: "BOOLEAN NOT NULL",
		},
		{
			col:     &core.Column{GoType: core.IntType},
			SQLType: "BIGINT NOT NULL",
		},

		{
			col: &core.Column{
				GoType: core.Int8Type,
				AI:     true,
				Length: []int{5, 6},
			},
			SQLType: "SERIAL NOT NULL",
		},
		{
			col: &core.Column{
				GoType:  core.Int8Type,
				Length:  []int{5, 6},
				Default: 1,
			},
			SQLType: "SMALLINT NOT NULL",
		},
		{
			col: &core.Column{
				GoType:     core.Int8Type,
				Length:     []int{5, 6},
				HasDefault: true,
				Default:    1,
			},
			SQLType: "SMALLINT NOT NULL DEFAULT '1'",
		},

		{
			col: &core.Column{
				GoType: core.NullInt64Type,
				AI:     true,
			},
			SQLType: "BIGSERIAL NOT NULL",
		},
		{
			col:     &core.Column{GoType: core.NullInt64Type},
			SQLType: "BIGINT NOT NULL",
		},

		{
			col: &core.Column{
				GoType: core.Uint32Type,
				AI:     true,
				Length: []int{5, 6},
			},
			SQLType: "SERIAL NOT NULL",
		},
		{
			col: &core.Column{
				GoType: core.Uint32Type,
				Length: []int{5, 6},
			},
			SQLType: "INT NOT NULL",
		},

		{
			col: &core.Column{
				GoType: core.IntType,
				Length: []int{5, 6},
			},
			SQLType: "BIGINT NOT NULL",
		},
		{
			col: &core.Column{
				GoType: core.IntType,
				AI:     true,
			},
			SQLType: "BIGSERIAL NOT NULL",
		},

		{
			col: &core.Column{
				GoType: core.Float32Type,
				Length: []int{5, 9},
			},
			SQLType: "NUMERIC(5,9) NOT NULL",
		},
		{ // 长度必须为 2
			col: &core.Column{
				GoType: core.Float32Type,
			},
			err: true,
		},
		{ // 长度必须为 2
			col: &core.Column{
				GoType: core.Float64Type,
				Length: []int{1},
			},
			err: true,
		},
		{
			col: &core.Column{
				GoType: core.NullFloat64Type,
				Length: []int{5, 9},
			},
			SQLType: "NUMERIC(5,9) NOT NULL",
		},
		{ // 长度必须为 2
			col: &core.Column{
				GoType: core.NullFloat64Type,
			},
			err: true,
		},

		{
			col:     &core.Column{GoType: core.StringType},
			SQLType: "TEXT NOT NULL",
		},
		{
			col:     &core.Column{GoType: reflect.TypeOf([]byte{'a', 'b'})},
			SQLType: "BYTEA NOT NULL",
		},

		{
			col:     &core.Column{GoType: core.NullStringType},
			SQLType: "TEXT NOT NULL",
		},
		{
			col: &core.Column{
				GoType: core.NullStringType,
				Length: []int{-1, 111},
			},
			SQLType: "TEXT NOT NULL",
		},
		{
			col: &core.Column{
				GoType: core.NullStringType,
				Length: []int{99, 111},
			},
			SQLType: "VARCHAR(99) NOT NULL",
		},

		{
			col: &core.Column{
				GoType: core.StringType,
				Length: []int{99},
			},
			SQLType: "VARCHAR(99) NOT NULL",
		},
		{
			col: &core.Column{
				GoType: core.StringType,
				Length: []int{11111111},
			},
			SQLType: "TEXT NOT NULL",
		},

		{
			col:     &core.Column{GoType: core.RawBytesType},
			SQLType: "BYTEA NOT NULL",
		},

		{
			col:     &core.Column{GoType: core.TimeType},
			SQLType: "TIMESTAMP NOT NULL",
		},
		{
			col: &core.Column{
				GoType: core.TimeType,
				Length: []int{-1},
			},
			err: true,
		},
		{
			col: &core.Column{
				GoType: core.TimeType,
				Length: []int{7},
			},
			err: true,
		},

		{ // 无法转换的类型
			col: &core.Column{GoType: reflect.TypeOf(struct{}{})},
			err: true,
		},
	}

	testSQLType(a, dialect.Postgres(), data)
}

func TestPostgres_TruncateTableStmtHooker(t *testing.T) {
	a := assert.New(t)

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		hook, ok := t.DB.Dialect().(sqlbuilder.TruncateTableStmtHooker)
		a.True(ok).NotNil(hook)

		stmt := sqlbuilder.TruncateTable(t.DB).Table("tbl", "")
		a.NotNil(stmt)
		qs, err := hook.TruncateTableStmtHook(stmt)
		a.NotError(err).Equal(1, len(qs))
		sqltest.Equal(a, qs[0], `TRUNCATE TABLE {tbl}`)

		stmt = sqlbuilder.TruncateTable(t.DB).Table("tbl", "id")
		a.NotNil(stmt)
		qs, err = hook.TruncateTableStmtHook(stmt)
		a.NotError(err).Equal(1, len(qs))
		sqltest.Equal(a, qs[0], `TRUNCATE TABLE {tbl} RESTART IDENTITY`)
	}, "postgres")
}

func TestPostgres_SQL(t *testing.T) {
	a := assert.New(t)
	p := dialect.Postgres()
	a.NotNil(p)

	eq := func(s1, s2 string) {
		ret, _, err := p.SQL(s1, nil)
		a.NotError(err)
		a.Equal(ret, s2)
	}

	err := func(s1 string) {
		ret, _, err := p.SQL(s1, nil)
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
		a.NotError(p.SQL(s1, nil))
	}
}

func TestPostgres_Types(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testTypes(t)
	}, "postgres")
}
