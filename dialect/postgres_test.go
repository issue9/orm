// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/orm/v2/sqlbuilder"
)

var (
	_ sqlbuilder.TruncateTableStmtHooker = &postgres{}
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
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(false)},
			SQLType: "BOOLEAN NOT NULL",
		},
		&test{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(1)},
			SQLType: "BIGINT NOT NULL",
		},

		&test{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(int8(1)),
				AI:     true,
				Length: []int{5, 6},
			},
			SQLType: "SERIAL NOT NULL",
		},
		&test{
			col: &sqlbuilder.Column{
				GoType:  reflect.TypeOf(int8(1)),
				Length:  []int{5, 6},
				Default: 1,
			},
			SQLType: "SMALLINT NOT NULL",
		},
		&test{
			col: &sqlbuilder.Column{
				GoType:     reflect.TypeOf(int8(1)),
				Length:     []int{5, 6},
				HasDefault: true,
				Default:    1,
			},
			SQLType: "SMALLINT NOT NULL DEFAULT '1'",
		},

		&test{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(sql.NullInt64{}),
				AI:     true,
			},
			SQLType: "BIGSERIAL NOT NULL",
		},
		&test{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.NullInt64{})},
			SQLType: "BIGINT NOT NULL",
		},

		&test{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(uint32(1)),
				AI:     true,
				Length: []int{5, 6},
			},
			SQLType: "SERIAL NOT NULL",
		},
		&test{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(uint32(1)),
				Length: []int{5, 6},
			},
			SQLType: "INT NOT NULL",
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
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(1.2),
				Length: []int{5, 9},
			},
			SQLType: "NUMERIC(5,9) NOT NULL",
		},
		&test{ // 长度必须为 2
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(1.2),
			},
			err: true,
		},
		&test{ // 长度必须为 2
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(1.2),
				Length: []int{1},
			},
			err: true,
		},
		&test{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(sql.NullFloat64{}),
				Length: []int{5, 9},
			},
			SQLType: "NUMERIC(5,9) NOT NULL",
		},
		&test{ // 长度必须为 2
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(sql.NullFloat64{}),
			},
			err: true,
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
				GoType: reflect.TypeOf(sql.NullString{}),
				Length: []int{-1, 111},
			},
			SQLType: "TEXT NOT NULL",
		},
		&test{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(sql.NullString{}),
				Length: []int{99, 111},
			},
			SQLType: "VARCHAR(99) NOT NULL",
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
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.RawBytes{})},
			SQLType: "BYTEA NOT NULL",
		},

		&test{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(time.Time{})},
			SQLType: "TIMESTAMP NOT NULL",
		},

		&test{ // 无法转换的类型
			col: &sqlbuilder.Column{GoType: reflect.TypeOf(struct{}{})},
			err: true,
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
