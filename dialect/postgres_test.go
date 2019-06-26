// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/v2/sqlbuilder"
)

func TestPostgres_SQLType(t *testing.T) {
	a := assert.New(t)

	var data = []*test{
		&test{ // col == nil
			err: true,
		},
		&test{ // col.GoType == nil
			col: &sqlbuilder.Column{GoType: nil},
			err: true,
		},
		&test{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(1)},
			SQLType: "BIGINT NOT NULL",
		},
		&test{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(1),
				Length: []int{5, 6},
			},
			SQLType: "BIGINT NOT NULL",
		},
		&test{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(1),
				AI:     true,
			},
			SQLType: "BIGSERIAL NOT NULL",
		},
		&test{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf("")},
			SQLType: "TEXT NOT NULL",
		},
		&test{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf([]byte{'a', 'b'})},
			SQLType: "BYTEA NOT NULL",
		},
		&test{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.NullString{})},
			SQLType: "TEXT NOT NULL",
		},
		&test{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(""),
				Length: []int{99},
			},
			SQLType: "VARCHAR(99) NOT NULL",
		},
		&test{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(""),
				Length: []int{11111111},
			},
			SQLType: "TEXT NOT NULL",
		},
		&test{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(1.2),
				Length: []int{5, 9},
			},
			SQLType: "NUMERIC(5,9) NOT NULL",
		},
		&test{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.RawBytes{})},
			SQLType: "BYTEA NOT NULL",
		},
	}

	testData(a, Postgres(), data)
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
