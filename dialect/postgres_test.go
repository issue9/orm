// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/core"
	"github.com/issue9/orm/internal/sqltest"
)

var _ base = &postgres{}

func TestPostgres_sqlType(t *testing.T) {
	p := &postgres{}

	a := assert.New(t)
	buf := core.NewStringBuilder("")
	col := &core.Column{}
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
	sqltest.Equal(a, buf.String(), "DOUBLE(5,6)")

	col.GoType = reflect.TypeOf([]byte{'1', '2'})
	buf.Reset()
	a.NotError(p.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "VARCHAR(5)")

	col.GoType = reflect.TypeOf(sql.NullInt64{})
	buf.Reset()
	a.NotError(p.sqlType(buf, col))
	sqltest.Equal(a, buf.String(), "BIGINT")
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
