// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package dialect_test

import (
	"database/sql"
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/orm/v5/core"
	"github.com/issue9/orm/v5/dialect"
	"github.com/issue9/orm/v5/internal/sqltest"
	"github.com/issue9/orm/v5/internal/test"
	"github.com/issue9/orm/v5/sqlbuilder"
)

func TestPostgres_VersionSQL(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "", test.Postgres)

	suite.Run(func(t *test.Driver) {
		testDialectVersionSQL(t)
	})
}

func TestPostgres_SQLType(t *testing.T) {
	a := assert.New(t, false)

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
			col:     &core.Column{PrimitiveType: core.Bytes},
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
			col:     &core.Column{PrimitiveType: core.Bytes},
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

func TestPostgres_TruncateTableSQL(t *testing.T) {
	a := assert.New(t, false)

	suite := test.NewSuite(a, "", test.Postgres)

	suite.Run(func(t *test.Driver) {
		stmt := sqlbuilder.TruncateTable(t.DB).Table("tbl", "")
		a.NotNil(stmt)
		qs, err := t.DB.Dialect().TruncateTableSQL("tbl", "")
		a.NotError(err).Equal(1, len(qs))
		sqltest.Equal(a, qs[0], `TRUNCATE TABLE {tbl}`)

		qs, err = t.DB.Dialect().TruncateTableSQL("tbl", "id")
		a.NotError(err).Equal(1, len(qs))
		sqltest.Equal(a, qs[0], `TRUNCATE TABLE {tbl} RESTART IDENTITY`)
	})
}

func TestPostgres_Fix(t *testing.T) {
	a := assert.New(t, false)
	p := dialect.Postgres("driver_name")
	a.NotNil(p)

	data := []*struct {
		input, output string
		args          []any
	}{
		{
			input:  "abc",
			output: "abc",
		},
		{ // 未包含 ? 的情况下，不会触发 p.replace 可以有 $
			input:  "abc$",
			output: "abc$",
		},
		{
			input:  "abc @id abc",
			output: "abc $1 abc",
			args:   []any{sql.Named("id", 1)},
		},
		{
			input:  "@id1 abc @id2 abc @id3",
			output: "$1 abc $2 abc $3",
			args:   []any{sql.Named("id1", 1), sql.Named("id2", 1), sql.Named("id3", 1)},
		},
		{
			input:  "abc @id1 abc @id2 def",
			output: "abc $1 abc $2 def",
			args:   []any{sql.Named("id1", 1), sql.Named("id2", 1)},
		},
		{
			input:  "中文 @id1 abc @id2 def",
			output: "中文 $1 abc $2 def",
			args:   []any{sql.Named("id1", 1), sql.Named("id2", 1)},
		},
	}

	for _, item := range data {
		output, _, err := p.Fix(item.input, item.args)
		a.NotError(err)
		sqltest.Equal(a, output, item.output)
	}

	_, _, err := p.Fix("$a @id1 bc", []any{sql.Named("id1", 1)})
	a.Error(err)

	_, _, err = p.Fix("@id1 $abc$", []any{sql.Named("id1", 1)})
	a.Error(err)

	_, _, err = p.Fix("a @id1 bc$abc$abc", []any{sql.Named("id1", 1)})
	a.Error(err)

	_, _, err = p.Fix("@id1 中$文", []any{sql.Named("id1", 1)})
	a.Error(err)
}

func BenchmarkPostgres_Fix(b *testing.B) {
	a := assert.New(b, false)
	p := dialect.Postgres("postgres_driver_name")
	a.NotNil(p)

	s1 := "SELECT * FROM tbl WHERE uid>? AND group=? AND username LIKE ?"

	for i := 0; i < b.N; i++ {
		_, _, err := p.Fix(s1, nil)
		a.NotError(err)
	}
}

func TestPostgres_Types(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "", test.Postgres)

	suite.Run(func(t *test.Driver) {
		testTypes(t)
	})
}

func TestPostgres_TypesDefault(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "", test.Postgres)

	suite.Run(func(t *test.Driver) {
		testTypesDefault(t)
	})
}
