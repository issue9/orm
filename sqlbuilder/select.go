// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"

	"github.com/issue9/orm/core"
	"github.com/issue9/orm/fetch"
)

// SelectStmt 查询语句
type SelectStmt struct {
	engine    core.Engine
	table     string
	where     *WhereStmt
	cols      []string
	distinct  string
	forupdate bool

	joins  []*join
	orders *core.StringBuilder
	group  string

	havingQuery string
	havingVals  []interface{}

	limitQuery string
	limitVals  []interface{}
}

type join struct {
	typ   string
	on    string
	table string
}

// Select 声明一条 Select 语句
func Select(e core.Engine) *SelectStmt {
	return &SelectStmt{
		engine: e,
		where:  newWhereStmt(),
	}
}

// Distinct 声明一条 Select 语句的 Distinct
func (stmt *SelectStmt) Distinct(col string) *SelectStmt {
	stmt.distinct = col
	return stmt
}

// Reset 重置语句
func (stmt *SelectStmt) Reset() {
	stmt.table = ""
	stmt.where.Reset()
	stmt.cols = stmt.cols[:0]
	stmt.distinct = ""
	stmt.forupdate = false

	stmt.joins = stmt.joins[:]
	stmt.orders.Reset()
	stmt.group = ""

	stmt.havingQuery = ""
	stmt.havingVals = nil

	stmt.limitQuery = ""
	stmt.limitVals = nil
}

// SQL 获取 SQL 语句及对应的参数
func (stmt *SelectStmt) SQL() (string, []interface{}, error) {
	if stmt.table == "" {
		return "", nil, ErrTableIsEmpty
	}

	if len(stmt.cols) == 0 {
		return "", nil, ErrColumnsIsEmpty
	}

	buf := core.NewStringBuilder("SELECT ")
	args := make([]interface{}, 0, 10)

	if stmt.distinct != "" {
		buf.WriteString("DISTINCT ")
		buf.WriteString(stmt.distinct)
		buf.WriteByte(' ')
	}

	for _, c := range stmt.cols {
		buf.WriteString(c)
		buf.WriteByte(',')
	}
	buf.TruncateLast(1)

	buf.WriteString(" FROM ")
	buf.WriteString(stmt.table)

	// join
	if len(stmt.joins) > 0 {
		buf.WriteByte(' ')
		for _, join := range stmt.joins {
			buf.WriteString(join.typ)
			buf.WriteString(" JOIN ")
			buf.WriteString(join.table)
			buf.WriteString(" ON ")
			buf.WriteString(join.on)
			buf.WriteByte(',')
		}
		buf.TruncateLast(1)
	}

	// where
	wq, wa, err := stmt.where.SQL()
	if err != nil {
		return "", nil, err
	}
	buf.WriteString(wq)
	args = append(args, wa...)

	// group by
	if stmt.group != "" {
		buf.WriteString(stmt.group)
	}

	// having
	if stmt.havingQuery != "" {
		buf.WriteString(stmt.havingQuery)
		args = append(args, stmt.havingVals...)
	}

	// order by
	if stmt.orders != nil && stmt.orders.Len() > 0 {
		buf.WriteString(stmt.orders.String())
	}

	// limit
	if stmt.limitQuery != "" {
		buf.WriteString(stmt.limitQuery)
		args = append(args, stmt.limitVals...)
	}

	// for update
	if stmt.forupdate {
		buf.WriteString(" FOR UPDATE")
	}

	return buf.String(), args, nil
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

// Having 指定 having 语句
func (stmt *SelectStmt) Having(expr string, args ...interface{}) *SelectStmt {
	stmt.havingQuery = expr
	stmt.havingVals = args

	return stmt
}

// WhereStmt 实现 WhereStmter 接口
func (stmt *SelectStmt) WhereStmt() *WhereStmt {
	return stmt.where
}

// Where 指定 where 语句
func (stmt *SelectStmt) Where(cond string, args ...interface{}) *SelectStmt {
	return stmt.And(cond, args...)
}

// And 指定 where ... AND ... 语句
func (stmt *SelectStmt) And(cond string, args ...interface{}) *SelectStmt {
	stmt.where.And(cond, args...)
	return stmt
}

// Or 指定 where ... OR ... 语句
func (stmt *SelectStmt) Or(cond string, args ...interface{}) *SelectStmt {
	stmt.where.Or(cond, args...)
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
		stmt.orders = core.NewStringBuilder(" ORDER BY ")
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

// ForUpdate 添加 FOR UPDATE 语句部分
func (stmt *SelectStmt) ForUpdate() *SelectStmt {
	stmt.forupdate = true
	return stmt
}

// Group 添加 GROUP BY 语句
func (stmt *SelectStmt) Group(col string) *SelectStmt {
	stmt.group = " GROUP BY " + col + " "
	return stmt
}

// Limit 生成 SQL 的 Limit 语句
func (stmt *SelectStmt) Limit(limit int, offset ...int) *SelectStmt {
	query, vals := stmt.engine.Dialect().LimitSQL(limit, offset...)
	stmt.limitQuery = query
	stmt.limitVals = vals
	return stmt
}

// Query 查询
func (stmt *SelectStmt) Query() (*sql.Rows, error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return stmt.engine.Query(query, args...)
}

// QueryContext 查询
func (stmt *SelectStmt) QueryContext(ctx context.Context) (*sql.Rows, error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return stmt.engine.QueryContext(ctx, query, args...)
}

// QueryObj 将符合当前条件的所有记录依次写入 objs 中。
func (stmt *SelectStmt) QueryObj(objs interface{}) (int, error) {
	rows, err := stmt.Query()
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return fetch.Obj(objs, rows)
}

// Prepare 预编译
func (stmt *SelectStmt) Prepare() (*sql.Stmt, error) {
	query, _, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return stmt.engine.Prepare(query)
}

// PrepareContext 预编译
func (stmt *SelectStmt) PrepareContext(ctx context.Context) (*sql.Stmt, error) {
	query, _, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return stmt.engine.PrepareContext(ctx, query)
}
