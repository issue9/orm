// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package sqlbuilder 用于构建 SQL 语句
package sqlbuilder

import (
	"context"
	"database/sql"
	"errors"

	"github.com/issue9/orm/core"
)

var (
	// ErrTableIsEmpty 未指定表名，任何 SQL 语句中，
	// 若未指定表名时，会返回此错误
	ErrTableIsEmpty = errors.New("表名为空")

	// ErrValueIsEmpty 在 Update 和 Insert 语句中，
	// 若未指定任何值，则返回此错误
	ErrValueIsEmpty = errors.New("值为空")

	// ErrColumnsIsEmpty 在 Insert 和 Select 语句中，
	// 若未指定任何列表，则返回此错误
	ErrColumnsIsEmpty = errors.New("未指定列")

	// ErrDupColumn 在 Update 中可能存在重复设置的列名。
	ErrDupColumn = errors.New("重复的列名")

	// ErrArgsNotMatch 在生成的 SQL 语句中，传递的参数与语句的占位符数量不匹配。
	ErrArgsNotMatch = errors.New("列与值的数量不匹配")
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

type execer interface {
	Exec() (sql.Result, error)
	ExecContext(ctx context.Context) (sql.Result, error)
	Prepare() (*sql.Stmt, error)
	PrepareContext(ctx context.Context) (*sql.Stmt, error)
}

type queryer interface {
	Query() (*sql.Rows, error)
	QueryContext(ctx context.Context) (*sql.Rows, error)
	Prepare() (*sql.Stmt, error)
	PrepareContext(ctx context.Context) (*sql.Stmt, error)
}

func exec(e core.Engine, stmt SQLer) (sql.Result, error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return e.Exec(query, args...)
}

func execContext(ctx context.Context, e core.Engine, stmt SQLer) (sql.Result, error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return e.ExecContext(ctx, query, args...)
}

func prepare(e core.Engine, stmt SQLer) (*sql.Stmt, error) {
	query, _, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return e.Prepare(query)
}

func prepareContext(ctx context.Context, e core.Engine, stmt SQLer) (*sql.Stmt, error) {
	query, _, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return e.PrepareContext(ctx, query)
}

func query(e core.Engine, stmt SQLer) (*sql.Rows, error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return e.Query(query, args...)
}

func queryContext(ctx context.Context, e core.Engine, stmt SQLer) (*sql.Rows, error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return e.QueryContext(ctx, query, args...)
}
