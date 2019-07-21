// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"

	"github.com/issue9/orm/v2/core"
)

// SQLer 定义 SQL 语句的基本接口
type SQLer interface {
	SQL() (query string, args []interface{}, err error)
}

// DDLSQLer SQL 中 DDL 语句的基本接口
//
// 大部分数据的 DDL 操作是有多条语句组成，比如 CREATE TABLE
// 可能包含了额外的定义信息。
type DDLSQLer interface {
	DDLSQL() ([]string, error)
}

type baseStmt struct {
	engine core.Engine
}

type queryStmt struct {
	SQLer
	baseStmt
}

type execStmt struct {
	SQLer
	baseStmt
}

type ddlStmt struct {
	DDLSQLer
	baseStmt
}

func newQueryStmt(e core.Engine, sql SQLer) *queryStmt {
	return &queryStmt{
		SQLer: sql,
		baseStmt: baseStmt{
			engine: e,
		},
	}
}

func newExecStmt(e core.Engine, sql SQLer) *execStmt {
	return &execStmt{
		SQLer: sql,
		baseStmt: baseStmt{
			engine: e,
		},
	}
}

func newDDLStmt(e core.Engine, sql DDLSQLer) *ddlStmt {
	return &ddlStmt{
		DDLSQLer: sql,
		baseStmt: baseStmt{
			engine: e,
		},
	}
}

func (stmt *baseStmt) Dialect() core.Dialect {
	return stmt.engine.Dialect()
}

func (stmt *baseStmt) Engine() core.Engine {
	return stmt.engine
}

func (stmt ddlStmt) Exec() error {
	return stmt.ExecContext(context.Background())
}

func (stmt *ddlStmt) ExecContext(ctx context.Context) error {
	qs, err := stmt.DDLSQL()
	if err != nil {
		return err
	}

	for _, query := range qs {
		if _, err = stmt.Engine().ExecContext(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

func (stmt *execStmt) Exec() (sql.Result, error) {
	return stmt.ExecContext(context.Background())
}

func (stmt *execStmt) ExecContext(ctx context.Context) (sql.Result, error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return nil, err
	}

	return stmt.Engine().ExecContext(ctx, query, args...)
}

// Prepare 预编译语句
//
// 预编译语句，参数最好采用 sql.NamedArg 类型。
// 在生成语句时，参数顺序会发生变化，如果采用 ? 的形式，
// 用户需要自己处理参数顺序问题，而 sql.NamedArg 没有这些问题。
func (stmt *execStmt) Prepare() (*core.Stmt, error) {
	return stmt.PrepareContext(context.Background())
}

func (stmt *execStmt) PrepareContext(ctx context.Context) (*core.Stmt, error) {
	query, _, err := stmt.SQL()
	if err != nil {
		return nil, err
	}

	return stmt.Engine().PrepareContext(ctx, query)
}

func (stmt *queryStmt) Prepare() (*core.Stmt, error) {
	return stmt.PrepareContext(context.Background())
}

func (stmt *queryStmt) PrepareContext(ctx context.Context) (*core.Stmt, error) {
	query, _, err := stmt.SQL()
	if err != nil {
		return nil, err
	}

	return stmt.Engine().PrepareContext(ctx, query)
}

func (stmt queryStmt) Query() (*sql.Rows, error) {
	return stmt.QueryContext(context.Background())
}

func (stmt *queryStmt) QueryContext(ctx context.Context) (*sql.Rows, error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return nil, err
	}

	return stmt.Engine().QueryContext(ctx, query, args...)
}
