// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/v2"
	"github.com/issue9/orm/v2/internal/sqltest"
	"github.com/issue9/orm/v2/model"
	"github.com/issue9/orm/v2/sqlbuilder"
)

var _ base = &postgres{}

func TestPostgres_CreateTableSQL(t *testing.T) {
	a := assert.New(t)
	ms := model.NewModels()
	m, err := ms.New(&user{})
	a.NotError(err).NotNil(m)

	sqls, err := Postgres().CreateTableSQL(m)
	a.NotError(err)
	a.Equal(2, len(sqls))
	sqltest.Equal(a, sqls[0], `CREATE TABLE IF NOT EXISTS {#user} (
		{id} BIGSERIAL NOT NULL,
		{name} varchar(20) NOT NULL,
		CONSTRAINT userpk PRIMARY KEY({id})
	)`)

	sqltest.Equal(a, sqls[1], `CREATE INDEX i_user_name ON {#user} (
		{name}
	)`)
}

func TestPostgres_sqlType(t *testing.T) {
	p := &postgres{}

	a := assert.New(t)
	buf := sqlbuilder.New("")
	col := &orm.Column{}
	a.Error(p.sqlType(buf, col))

	col.GoType = reflect.TypeOf(1)
	buf.Reset()
	a.NotError(p.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "BIGINT")

	col.Len1 = 5
	col.Len2 = 6
	buf.Reset()
	a.NotError(p.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "BIGINT")

	col.GoType = reflect.TypeOf("abc")
	buf.Reset()
	a.NotError(p.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "VARCHAR(5)")

	col.GoType = reflect.TypeOf(1.2)
	buf.Reset()
	a.NotError(p.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "NUMERIC(5,6)")

	col.GoType = reflect.TypeOf(sql.NullInt64{})
	buf.Reset()
	a.NotError(p.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "BIGINT")

	col.GoType = reflect.TypeOf(sql.RawBytes("123"))
	buf.Reset()
	a.NotError(p.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "BYTEA")
}

func TestPostgres_SQL(t *testing.T) {
	a := assert.New(t)
	p := Postgres()
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
	p := Postgres()
	a.NotNil(p)

	s1 := "SELECT * FROM tbl WHERE uid>? AND group=? AND username LIKE ?"

	for i := 0; i < b.N; i++ {
		p.SQL(s1)
	}
}
