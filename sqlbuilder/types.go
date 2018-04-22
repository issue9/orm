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

	Exec(query string, args ...interface{}) (sql.Result, error)

	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	Prepare(query string) (*sql.Stmt, error)

	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// Dialect 接口用于描述与数据库相关的一些语言特性。
type Dialect interface {
	// 返回符合当前数据库规范的引号对。
	QuoteTuple() (openQuote, closeQuote byte)

	// 生成 `LIMIT N OFFSET M` 或是相同的语意的语句。
	//
	// offset 值为一个可选参数，若不指定，则表示 `LIMIT N` 语句。
	// 返回的是对应数据库的 limit 语句以及语句中占位符对应的值。
	//
	// limit 和 offset 可以是 sql.NamedArg 类型。
	LimitSQL(limit interface{}, offset ...interface{}) (string, []interface{})

	// 是否允许在事务中执行 DDL
	//
	// 比如在 postgresql 中，如果创建一个带索引的表，会采用在事务中，
	// 分多条语句创建表。
	// 而像 mysql 等不支持事务内 DDL 的数据库，则会采用普通的方式，
	// 依次提交语句。
	TransactionalDDL() bool
}

func exec(e Engine, stmt SQLer) (sql.Result, error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return e.Exec(query, args...)
}

func execContext(ctx context.Context, e Engine, stmt SQLer) (sql.Result, error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return e.ExecContext(ctx, query, args...)
}

func prepare(e Engine, stmt SQLer) (*sql.Stmt, error) {
	query, _, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return e.Prepare(query)
}

func prepareContext(ctx context.Context, e Engine, stmt SQLer) (*sql.Stmt, error) {
	query, _, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return e.PrepareContext(ctx, query)
}

func query(e Engine, stmt SQLer) (*sql.Rows, error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return e.Query(query, args...)
}

func queryContext(ctx context.Context, e Engine, stmt SQLer) (*sql.Rows, error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return e.QueryContext(ctx, query, args...)
}
