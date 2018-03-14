// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/issue9/orm/fetch"
)

// SelectStmt 查询语句
type SelectStmt struct {
	engine    Engine
	dialect   Dialect
	table     string
	where     *WhereStmt
	cols      []string
	distinct  bool
	forupdate bool

	// COUNT 查询的列内容
	countExpr string

	joins  []*join
	orders *SQLBuilder
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
func Select(e Engine, d Dialect) *SelectStmt {
	return &SelectStmt{
		engine:  e,
		dialect: d,
		where:   newWhereStmt(),
	}
}

// Distinct 声明一条 Select 语句的 Distinct
//
// 若指定了此值，则 Select() 所指定的列，均为 Distinct 之后的列。
func (stmt *SelectStmt) Distinct() *SelectStmt {
	stmt.distinct = true
	return stmt
}

// Reset 重置语句
func (stmt *SelectStmt) Reset() {
	stmt.table = ""
	stmt.where.Reset()
	stmt.cols = stmt.cols[:0]
	stmt.distinct = false
	stmt.forupdate = false

	stmt.countExpr = ""

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

	if len(stmt.cols) == 0 && stmt.countExpr == "" {
		return "", nil, ErrColumnsIsEmpty
	}

	buf := New("SELECT ")
	args := make([]interface{}, 0, 10)

	if stmt.countExpr == "" {
		if stmt.distinct {
			buf.WriteString("DISTINCT ")
		}
		for _, c := range stmt.cols {
			buf.WriteString(c)
			buf.WriteByte(',')
		}
		buf.TruncateLast(1)
	} else {
		buf.WriteString(stmt.countExpr)
	}

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
	buf.WriteString(" WHERE ")
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
	if stmt.countExpr == "" && stmt.limitQuery != "" {
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
		stmt.cols = make([]string, 0, len(cols))
	}

	stmt.cols = append(stmt.cols, cols...)
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
		stmt.orders = New("")
	}

	if stmt.orders.Len() == 0 {
		stmt.orders.WriteString(" ORDER BY ")
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
	query, vals := stmt.dialect.LimitSQL(limit, offset...)
	stmt.limitQuery = query
	stmt.limitVals = vals
	return stmt
}

// Count 指定 Count 表示式，如果指定了 count 表达式，则会造成 limit 失效。
//
// 传递空的 expr 参数，表示去除 count 表达式。
func (stmt *SelectStmt) Count(expr string) *SelectStmt {
	stmt.countExpr = expr
	return stmt
}

// Prepare 预编译
func (stmt *SelectStmt) Prepare() (*sql.Stmt, error) {
	return prepare(stmt.engine, stmt)
}

// PrepareContext 预编译
func (stmt *SelectStmt) PrepareContext(ctx context.Context) (*sql.Stmt, error) {
	return prepareContext(ctx, stmt.engine, stmt)
}

// Query 查询
func (stmt *SelectStmt) Query() (*sql.Rows, error) {
	return query(stmt.engine, stmt)
}

// QueryContext 查询
func (stmt *SelectStmt) QueryContext(ctx context.Context) (*sql.Rows, error) {
	return queryContext(ctx, stmt.engine, stmt)
}

// QueryObj 将符合当前条件的所有记录依次写入 objs 中。
//
// 关于 objs 的值类型，可以参考 github.com/issue9/orm/fetch.Obj 函数的相关介绍。
func (stmt *SelectStmt) QueryObj(objs interface{}) (int, error) {
	rows, err := stmt.Query()
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return fetch.Obj(objs, rows)
}

// QueryInt 查询指定列的第一行数据，并将其转换成 int
func (stmt *SelectStmt) QueryInt(colName string) (int64, error) {
	rows, err := stmt.Query()
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	cols, err := fetch.ColumnString(true, colName, rows)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(cols[0], 10, 64)
}
