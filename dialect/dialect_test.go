// SPDX-License-Identifier: MIT

package dialect

import (
	"database/sql"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v4/internal/flagtest"
	"github.com/issue9/orm/v4/internal/sqltest"
)

func TestMain(m *testing.M) {
	flagtest.Main(m)
}

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

func TestPrepareNamedArgs(t *testing.T) {
	a := assert.New(t)

	var data = []*struct {
		input  string
		query  string
		orders map[string]int
		err    bool
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
		{
			input:  "INSERT INTO users({id},{name}) VALUES (@id,@name)",
			query:  "INSERT INTO users({id},{name}) VALUES (?,?)",
			orders: map[string]int{"id": 0, "name": 1},
		},
		{
			input:  "INSERT INTO users({id},{name}) VALUES (?,?)",
			query:  "INSERT INTO users({id},{name}) VALUES (?,?)",
			orders: map[string]int{},
		},
		{
			input:  "select * from table where id=? and id=@id and id=1",
			query:  "select * from table where id=? and id=? and id=1",
			orders: map[string]int{"id": 1},
		},
		{ // 参数名称是另一个参数名称的一部分
			input:  "select * from table where id=@id and id=? and id=@id2",
			query:  "select * from table where id=? and id=? and id=?",
			orders: map[string]int{"id": 0, "id2": 2},
		},
	}

	for _, item := range data {
		q, o, err := PrepareNamedArgs(item.input)

		if item.err {
			a.Error(err).Nil(o).Empty(q)
			continue
		}

		a.NotError(err).
			Equal(o, item.orders)
		sqltest.Equal(a, q, item.query)
	}

	a.Panic(func() {
		PrepareNamedArgs("INSERT INTO users({id},{name}) VALUES (@id,@id)")
	})
}

func TestFixQueryAndArgs(t *testing.T) {
	a := assert.New(t)

	data := []*struct {
		query       string
		args        []interface{}
		err         bool
		outputQuery string
		outputArgs  []interface{}
	}{
		{
			query:       "select * from table where id=? and id=@id and id=1",
			args:        []interface{}{1, sql.Named("id", 2)},
			outputQuery: "select * from table where id=? and id=? and id=1",
			outputArgs:  []interface{}{1, 2},
		},
		{
			query:       "select * from table where id=@id and id=? and id=@id2",
			args:        []interface{}{sql.Named("id2", 1), 1, sql.Named("id", 2)},
			outputQuery: "select * from table where id=? and id=? and id=?",
			outputArgs:  []interface{}{2, 1, 1},
		},
	}

	for _, item := range data {
		query, args, err := fixQueryAndArgs(item.query, item.args)
		a.NotError(err).
			Equal(args, item.outputArgs)
		sqltest.Equal(a, query, item.outputQuery)
	}

	a.Panic(func() {
		fixQueryAndArgs("select * from table where id=@id and id=? and id=@id2", []interface{}{sql.Named("id2", 1), 1, sql.Named("id3", 2)})
	})

	a.Panic(func() {
		fixQueryAndArgs("select * from table where id=@id and id=? and id=@id2", []interface{}{sql.Named("id2", 1), 1, sql.Named("not-exists", 2)})
	})

	a.Panic(func() {
		fixQueryAndArgs("select * from table where id=@id and id=? and id=@id", []interface{}{sql.Named("id", 1), 1, sql.Named("id2", 2)})
	})
}
