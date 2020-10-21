// SPDX-License-Identifier: MIT

package dialect_test

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v3/core"
	"github.com/issue9/orm/v3/dialect"
	"github.com/issue9/orm/v3/internal/sqltest"
	"github.com/issue9/orm/v3/internal/test"
	"github.com/issue9/orm/v3/sqlbuilder"
)

func TestPostgres_VersionSQL(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testDialectVersionSQL(t)
	}, test.Postgres)
}

func TestPostgres_SQLType(t *testing.T) {
	a := assert.New(t)

	var data = []*sqlTypeTester{
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
			SQLType: "SMALLINT NOT NULL DEFAULT 1",
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

	testSQLType(a, dialect.Postgres("postgres", "postgres_driver"), data)
}

func TestPostgres_FormatSQL(t *testing.T) {
	a := assert.New(t)
	now := time.Now().In(time.UTC)

	var data = []*sqlFormatTester{
		{
			v:      1,
			format: "1",
		},
		{
			v:      int8(1),
			format: "1",
		},

		{ // Bool
			v:      true,
			format: "true",
		},
		{
			v:      false,
			format: "false",
		},

		{ // NullBool
			v:      sql.NullBool{Valid: true, Bool: true},
			format: "true",
		},
		{
			v:      sql.NullBool{Valid: true, Bool: false},
			format: "false",
		},
		{
			v:      sql.NullBool{Valid: false, Bool: true},
			format: "NULL",
		},

		{ // NullInt64
			v:      sql.NullInt64{Valid: true, Int64: 64},
			format: "64",
		},
		{
			v:      sql.NullInt64{Valid: true, Int64: -1},
			format: "-1",
		},
		{
			v:      sql.NullInt64{Valid: false, Int64: 64},
			format: "NULL",
		},

		{ // NullFloat64
			v:      sql.NullFloat64{Valid: true, Float64: 6.4},
			format: "6.4",
		},
		{
			v:      sql.NullFloat64{Valid: true, Float64: -1.64},
			format: "-1.64",
		},
		{
			v:      sql.NullFloat64{Valid: false, Float64: 6.4},
			format: "NULL",
		},

		{ // NullString
			v:      sql.NullString{Valid: true, String: "str"},
			format: "'str'",
		},
		{
			v:      sql.NullString{Valid: true, String: ""},
			format: "''",
		},
		{
			v:      sql.NullString{Valid: false, String: "str"},
			format: "NULL",
		},

		// time
		{ // 长度错误
			v:   now,
			l:   []int{1, 2},
			err: true,
		},
		{ // 长度错误
			v:   now,
			l:   []int{600},
			err: true,
		},
		{
			v:      now,
			format: "'" + now.Format("2006-01-02 15:04:05Z07:00") + "'",
		},
		{
			v:      now,
			l:      []int{1},
			format: "'" + now.Format("2006-01-02 15:04:05.9Z07:00") + "'",
		},
		{
			v:      now,
			l:      []int{3},
			format: "'" + now.Format("2006-01-02 15:04:05.999Z07:00") + "'",
		},
	}

	testFormatSQL(a, dialect.Postgres("postgres", "driverName"), data)
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
	}, test.Postgres)
}

func TestPostgres_SQL(t *testing.T) {
	a := assert.New(t)
	p := dialect.Postgres("postgres", "driver_name")
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
	p := dialect.Postgres("psql", "postgres_driver_name")
	a.NotNil(p)

	s1 := "SELECT * FROM tbl WHERE uid>? AND group=? AND username LIKE ?"

	for i := 0; i < b.N; i++ {
		a.NotError(p.Fix(s1, nil))
	}
}

func TestPostgres_Types(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testTypes(t)
	}, test.Postgres)
}

func TestPostgres_TypesDefault(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testTypesDefault(t)
	}, test.Postgres)
}
