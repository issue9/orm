// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"github.com/issue9/orm/forward"
	"github.com/issue9/orm/internal/stringbuilder"
)

// SelectStmt 查询语句
type SelectStmt struct {
	engine forward.Engine
	table  string
	where  *where
	cols   []string
	joins  []*join
	orders *stringbuilder.StringBuilder
}

type join struct {
	typ   string
	on    string
	table string
}

// Select 声明一条 Select 语句
func Select(e forward.Engine) *SelectStmt {
	return &SelectStmt{
		engine: e,
		where:  newWhere(),
	}
}

// Reset 重置语句
func (stmt *SelectStmt) Reset() {
	// TODO
}

// SQL 获取 SQL 语句及对应的参数
func (stmt *SelectStmt) SQL() (string, []interface{}, error) {
	// TODO
	return "", nil, nil
}

// Select 指定列名
func (stmt *SelectStmt) Select(cols ...string) *SelectStmt {
	if stmt.cols == nil {
		stmt.cols = cols
	} else {
		stmt.cols = append(stmt.cols, cols...)
	}
	return stmt
}

// From 指定表名
func (stmt *SelectStmt) From(table string) *SelectStmt {
	stmt.table = table

	return stmt
}

// Where 指定 where 语句
func (stmt *SelectStmt) Where(and bool, cond string, args ...interface{}) *SelectStmt {
	stmt.where.where(and, cond, args...)
	return stmt
}

// And 指定 where ... AND ... 语句
func (stmt *SelectStmt) And(cond string, args ...interface{}) *SelectStmt {
	stmt.where.and(cond, args...)
	return stmt
}

// Or 指定 where ... OR ... 语句
func (stmt *SelectStmt) Or(cond string, args ...interface{}) *SelectStmt {
	stmt.where.or(cond, args...)
	return stmt
}

// Join 添加一条 Join 语句
func (stmt *SelectStmt) Join(typ, table, on string) *SelectStmt {
	if stmt.joins == nil {
		stmt.joins = make([]*join, 0, 5)
	}

	stmt.joins = append(stmt.joins, &join{typ: typ, table: table, on: on})
	return stmt
}

// Desc 倒序查询
func (stmt *SelectStmt) Desc(col ...string) *SelectStmt {
	return stmt.orderBy(false, col...)
}

// Asc 正序查询
func (stmt *SelectStmt) Asc(col ...string) *SelectStmt {
	return stmt.orderBy(true, col...)
}

func (stmt *SelectStmt) orderBy(asc bool, col ...string) *SelectStmt {
	if stmt.orders == nil {
		stmt.orders = stringbuilder.New(" ORDER BY ")
	} else {
		stmt.orders.WriteByte(',')
	}

	for _, c := range col {
		stmt.orders.WriteString(c)
		stmt.orders.WriteByte(',')
	}
	stmt.orders.TruncateLast(1)

	if asc {
		stmt.orders.WriteString(" ASC ")
	} else {
		stmt.orders.WriteString(" DESC ")
	}

	return stmt
}

// Incr,Decr

// Limit,Group
