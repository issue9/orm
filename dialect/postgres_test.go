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
)

func TestPostgres_SQLType(t *testing.T) {
	p := &postgres{}

	a := assert.New(t)
	col := &orm.Column{}

	// col == nil
	typ, err := p.SQLType(nil)
	a.ErrorType(err, errColIsNil).Empty(typ)

	// col.GoType == nil
	typ, err = p.SQLType(col)
	a.ErrorType(err, errGoTypeIsNil).Empty(typ)

	col.GoType = reflect.TypeOf(1)
	typ, err = p.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "BIGINT NOT NULL")

	col.Len1 = 5
	col.Len2 = 6
	typ, err = p.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "BIGINT NOT NULL")

	col.GoType = reflect.TypeOf("abc")
	typ, err = p.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "VARCHAR(5) NOT NULL")

	col.GoType = reflect.TypeOf(1.2)
	typ, err = p.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "NUMERIC(5,6) NOT NULL")

	col.GoType = reflect.TypeOf(sql.NullInt64{})
	typ, err = p.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "BIGINT NOT NULL")

	col.GoType = reflect.TypeOf(sql.RawBytes("123"))
	typ, err = p.SQLType(col)
	a.NotError(err)
	sqltest.Equal(a, typ, "BYTEA NOT NULL")
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
