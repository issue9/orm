// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
)

// SQLer 定义 SQL 语句的基本接口
type SQLer interface {
	// 获取 SQL 语句以及其关联的参数
	SQL() (query string, args []interface{}, err error)

	// 重置整个 SQL 语句。
	Reset()
}

// WhereStmter 带 Where 语句的 SQL
type WhereStmter interface {
	WhereStmt() *WhereStmt
}

// Engine 数据库执行的基本接口。
type Engine interface {
	// 执行一条查询语句，并返回相应的 sql.Rows 实例。
	// 功能等同于标准库 database/sql 的 DB.Query()
	//
	// query 会被作相应的转换。以 mysql 为例，假设当前的 prefix 为 p_
	//  select * from #user where {group}=1
	//  // 转换后
	//  select * from prefix_user where `group`=1
	Query(query string, args ...interface{}) (*sql.Rows, error)

	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

	QueryRow(query string, args ...interface{}) *sql.Row

	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row

	Exec(query string, args ...interface{}) (sql.Result, error)

	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	Prepare(query string) (*sql.Stmt, error)

	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// Dialect 接口用于描述与数据库相关的一些语言特性。
type Dialect interface {
	// 生成 `LIMIT N OFFSET M` 或是相同的语意的语句。
	//
	// offset 值为一个可选参数，若不指定，则表示 `LIMIT N` 语句。
	// 返回的是对应数据库的 limit 语句以及语句中占位符对应的值。
	//
	// limit 和 offset 可以是 sql.NamedArg 类型。
	LimitSQL(limit interface{}, offset ...interface{}) (string, []interface{})

	// 自定义获取 LastInsertID 的获取方式。
	//
	// 类似于 postgresql 等都需要额外定义。
	//
	// 返回参数 sql 表示额外的语句，如果为空，则执行的是标准的 SQL 插入语句。
	// append 表示在 sql 不为空的情况下，sql 与现有的插入语句的结合方式，
	// 如果为 true 表示直接添加在插入语句之后，否则为一条新的语句。
	LastInsertID(table, col string) (sql string, append bool)
}

func exec(e Engine, stmt SQLer) (sql.Result, error) {
	return execContext(context.Background(), e, stmt)
}

func execContext(ctx context.Context, e Engine, stmt SQLer) (sql.Result, error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return e.ExecContext(ctx, query, args...)
}

func prepare(e Engine, stmt SQLer) (*sql.Stmt, error) {
	return prepareContext(context.Background(), e, stmt)
}

func prepareContext(ctx context.Context, e Engine, stmt SQLer) (*sql.Stmt, error) {
	query, _, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return e.PrepareContext(ctx, query)
}

func query(e Engine, stmt SQLer) (*sql.Rows, error) {
	return queryContext(context.Background(), e, stmt)
}

func queryContext(ctx context.Context, e Engine, stmt SQLer) (*sql.Rows, error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return e.QueryContext(ctx, query, args...)
}
