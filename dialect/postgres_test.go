// SPDX-License-Identifier: MIT

package dialect_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v4/core"
	"github.com/issue9/orm/v4/dialect"
	"github.com/issue9/orm/v4/internal/sqltest"
	"github.com/issue9/orm/v4/internal/test"
	"github.com/issue9/orm/v4/sqlbuilder"
)

func TestPostgres_VersionSQL(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a, test.Postgres)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testDialectVersionSQL(t)
	})
}

func TestPostgres_SQLType(t *testing.T) {
	a := assert.New(t)

	var data = []*sqlTypeTester{
		{ // col.PrimitiveType = Auto
			col: &core.Column{PrimitiveType: core.Auto},
			err: true,
		},
		{
			col:     &core.Column{PrimitiveType: core.Bool},
			SQLType: "BOOLEAN NOT NULL",
		},
		{
			col:     &core.Column{PrimitiveType: core.Int},
			SQLType: "BIGINT NOT NULL",
		},

		{
			col: &core.Column{
				PrimitiveType: core.Int8,
				AI:            true,
				Length:        []int{5, 6},
			},
			SQLType: "SERIAL NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Int8,
				Length:        []int{5, 6},
				Default:       1,
			},
			SQLType: "SMALLINT NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Int8,
				Length:        []int{5, 6},
				HasDefault:    true,
				Default:       1,
			},
			SQLType: "SMALLINT NOT NULL DEFAULT 1",
		},

		{
			col: &core.Column{
				PrimitiveType: core.Int64,
				AI:            true,
			},
			SQLType: "BIGSERIAL NOT NULL",
		},
		{
			col:     &core.Column{PrimitiveType: core.Int64},
			SQLType: "BIGINT NOT NULL",
		},

		{
			col: &core.Column{
				PrimitiveType: core.Uint32,
				AI:            true,
				Length:        []int{5, 6},
			},
			SQLType: "SERIAL NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Uint32,
				Length:        []int{5, 6},
			},
			SQLType: "INT NOT NULL",
		},

		{
			col: &core.Column{
				PrimitiveType: core.Int,
				Length:        []int{5, 6},
			},
			SQLType: "BIGINT NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Int,
				AI:            true,
			},
			SQLType: "BIGSERIAL NOT NULL",
		},

		{
			col: &core.Column{
				PrimitiveType: core.Float32,
				Length:        []int{5, 9},
			},
			SQLType: "REAL NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Float64,
				Length:        []int{5, 9},
			},
			SQLType: "DOUBLE PRECISION NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Decimal,
				Length:        []int{5, 9},
			},
			SQLType: "decimal(5,9) NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Decimal,
				Length:        []int{5},
			},
			err: true,
		},
		{
			col:     &core.Column{PrimitiveType: core.String},
			SQLType: "TEXT NOT NULL",
		},
		{
			col:     &core.Column{PrimitiveType: core.RawBytes},
			SQLType: "BYTEA NOT NULL",
		},

		{
			col:     &core.Column{PrimitiveType: core.String},
			SQLType: "TEXT NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.String,
				Length:        []int{-1, 111},
			},
			SQLType: "TEXT NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.String,
				Length:        []int{99, 111},
			},
			SQLType: "VARCHAR(99) NOT NULL",
		},

		{
			col: &core.Column{
				PrimitiveType: core.String,
				Length:        []int{99},
			},
			SQLType: "VARCHAR(99) NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.String,
				Length:        []int{11111111},
			},
			SQLType: "TEXT NOT NULL",
		},

		{
			col:     &core.Column{PrimitiveType: core.RawBytes},
			SQLType: "BYTEA NOT NULL",
		},

		{
			col:     &core.Column{PrimitiveType: core.Time},
			SQLType: "TIMESTAMP NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Time,
				Length:        []int{-1},
			},
			err: true,
		},
		{
			col: &core.Column{
				PrimitiveType: core.Time,
				Length:        []int{7},
			},
			err: true,
		},
	}

	testSQLType(a, dialect.Postgres("postgres_driver"), data)
}

func TestPostgres_TruncateTableStmtHooker(t *testing.T) {
	a := assert.New(t)

	suite := test.NewSuite(a, test.Postgres)
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
	})
}

func TestPostgres_SQL(t *testing.T) {
	a := assert.New(t)
	p := dialect.Postgres("driver_name")
	a.NotNil(p)

	eq := func(s1, s2 string) {
		ret, _, err := p.Fix(s1, nil)
		a.NotError(err)
		a.Equal(ret, s2)
	}

	err := func(s1 string) {
		ret, _, err := p.Fix(s1, nil)
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
	p := dialect.Postgres("postgres_driver_name")
	a.NotNil(p)

	s1 := "SELECT * FROM tbl WHERE uid>? AND group=? AND username LIKE ?"

	for i := 0; i < b.N; i++ {
		a.NotError(p.Fix(s1, nil))
	}
}

func TestPostgres_Types(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a, test.Postgres)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testTypes(t)
	})
}

func TestPostgres_TypesDefault(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a, test.Postgres)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testTypesDefault(t)
	})
}
