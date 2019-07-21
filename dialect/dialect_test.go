// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"database/sql"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/sqltest"
)

func TestMysqlLimitSQL(t *testing.T) {
	a := assert.New(t)

	query, ret := mysqlLimitSQL(5, 0)
	a.Equal(ret, []int{5, 0})
	sqltest.Equal(a, query, " LIMIT ? OFFSET ? ")

	query, ret = mysqlLimitSQL(5)
	a.Equal(ret, []int{5})
	sqltest.Equal(a, query, "LIMIT ?")

	// 带 sql.namedArg
	query, ret = mysqlLimitSQL(sql.Named("limit", 1), 2)
	a.Equal(ret, []interface{}{sql.Named("limit", 1), 2})
	sqltest.Equal(a, query, "LIMIT @limit offset ?")
}

func TestOracleLimitSQL(t *testing.T) {
	a := assert.New(t)

	query, ret := oracleLimitSQL(5, 0)
	a.Equal(ret, []int{0, 5})
	sqltest.Equal(a, query, " OFFSET ? ROWS FETCH NEXT ? ROWS ONLY ")

	query, ret = oracleLimitSQL(5)
	a.Equal(ret, []int{5})
	sqltest.Equal(a, query, "FETCH NEXT ? ROWS ONLY ")

	// 带 sql.namedArg
	query, ret = oracleLimitSQL(sql.Named("limit", 1), 2)
	a.Equal(ret, []interface{}{2, sql.Named("limit", 1)})
	sqltest.Equal(a, query, "offset ? rows fetch next @limit rows only")
}

func TestReplaceNamedArgs(t *testing.T) {
	a := assert.New(t)

	var data = []*struct {
		inputQuery  string
		inputArgs   []interface{}
		outputQuery string
		outputArgs  []interface{}
	}{
		{
			inputQuery:  "select * from table",
			outputQuery: "select * from table",
		},
		{
			inputQuery:  "select * from table where id=?",
			inputArgs:   []interface{}{1},
			outputQuery: "select * from table where id=?",
			outputArgs:  []interface{}{1},
		},
		{
			inputQuery:  "select * from table where id=@id",
			inputArgs:   []interface{}{sql.Named("id", 1)},
			outputQuery: "select * from table where id=?",
			outputArgs:  []interface{}{1},
		},
		{
			inputQuery:  "select * from table where id=@id and id=? and id=1",
			inputArgs:   []interface{}{sql.Named("id", 1), 2},
			outputQuery: "select * from table where id=? and id=? and id=1",
			outputArgs:  []interface{}{1, 2},
		},
		{
			inputQuery:  "select * from table where id=@id and id=? and id=1",
			inputArgs:   []interface{}{&sql.NamedArg{Name: "id", Value: 1}, 2},
			outputQuery: "select * from table where id=? and id=? and id=1",
			outputArgs:  []interface{}{1, 2},
		},
		{ // 参数名称是另一个参数名称的一部分
			inputQuery:  "select * from table where id=@id and id=@idMax and id=1",
			inputArgs:   []interface{}{sql.Named("id", 1), sql.Named("idMax", 2)},
			outputQuery: "select * from table where id=? and id=? and id=1",
			outputArgs:  []interface{}{1, 2},
		},
		{ // 参数名称是另一个参数名称的一部分
			inputQuery:  "select * from table where id=@idMax and id=@id and id=1",
			inputArgs:   []interface{}{sql.Named("id", 1), sql.Named("idMax", 2)},
			outputQuery: "select * from table where id=? and id=? and id=1",
			outputArgs:  []interface{}{1, 2},
		},
	}

	for _, item := range data {
		output := replaceNamedArgs(item.inputQuery, item.inputArgs)
		sqltest.Equal(a, output, item.outputQuery)
		a.Equal(item.inputArgs, item.outputArgs)
	}
}

func TestPrepareNamedArgs(t *testing.T) {
	a := assert.New(t)

	var data = []*struct {
		input  string
		query  string
		orders map[string]int
	}{
		{
			input:  "select * from table",
			query:  "select * from table",
			orders: map[string]int{},
		},
		{
			input:  "select * from table where id=@id",
			query:  "select * from table where id=?",
			orders: map[string]int{"id": 0},
		},
		{
			input:  "select * from table where id=@id and name like @name",
			query:  "select * from table where id=? and name like ?",
			orders: map[string]int{"id": 0, "name": 1},
		},
		{
			input:  "select * from table where {id}=@id and {name} like @name",
			query:  "select * from table where {id}=? and {name} like ?",
			orders: map[string]int{"id": 0, "name": 1},
		},
		{
			input:  "select * from table where {编号}=@编号 and {name} like @name",
			query:  "select * from table where {编号}=? and {name} like ?",
			orders: map[string]int{"编号": 0, "name": 1},
		},
	}

	for _, item := range data {
		q, o := PrepareNamedArgs(item.input)
		a.Equal(o, item.orders)
		sqltest.Equal(a, q, item.query)
	}
}
