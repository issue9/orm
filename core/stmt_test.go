// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package core

import (
	"database/sql"
	"testing"

	"github.com/issue9/assert/v4"
)

func TestNewStmt(t *testing.T) {
	a := assert.New(t, false)

	a.NotPanic(func() {
		stmt := NewStmt(nil, nil)
		a.NotNil(stmt)
	})

	a.NotPanic(func() {
		stmt := NewStmt(nil, map[string]int{"id": 0, "name": 1, "test": 2})
		a.NotNil(stmt)
	})

	a.NotPanic(func() {
		stmt := NewStmt(nil, map[string]int{"id": 2, "name": 1, "test": 0})
		a.NotNil(stmt)
	})

	a.Panic(func() {
		NewStmt(nil, map[string]int{"id": 1})
	})

	a.Panic(func() {
		NewStmt(nil, map[string]int{"id": 1, "name": 0, "test": 10})
	})
}

func TestStmt_buildArgs(t *testing.T) {
	a := assert.New(t, false)

	data := []*struct {
		orders map[string]int
		input  []any
		output []any
		err    bool
	}{
		{},
		{ // orders 为空，则原样返回内容
			input:  []any{1, 2, 3},
			output: []any{1, 2, 3},
		},
		{ // orders 为空，则原样返回内容
			orders: map[string]int{},
			input:  []any{1, 2, 3},
			output: []any{1, 2, 3},
		},
		{ // 参数数量不匹配
			orders: map[string]int{"id": 0},
			input:  []any{sql.Named("id", 1), 1, 2},
			err:    true,
		},
		{ // 输入参数有非 sql.Named 类型
			orders: map[string]int{"id": 0, "name": 1},
			input:  []any{sql.Named("id", 1), 1},
			err:    true,
		},
		{ // 参数并不在 orders 中
			orders: map[string]int{"id": 0, "name": 1},
			input:  []any{sql.Named("id", 1), sql.Named("not-exists-arg", "test")},
			err:    true,
		},
		{
			orders: map[string]int{"name": 1, "id": 0},
			input:  []any{sql.Named("id", 1), sql.Named("name", "test")},
			output: []any{1, "test"},
		},
	}

	for k, v := range data {
		stmt := NewStmt(nil, v.orders)
		output, err := stmt.buildArgs(v.input)
		if v.err {
			a.Error(err, "not error @ %d", k).
				Nil(output)
		} else {
			a.Equal(output, v.output, "not equal @%d,v1:%s,v2:%s", k, output, v.output)
		}
	}
}
