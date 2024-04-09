// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package sqlbuilder

import (
	"context"
	"database/sql"

	"github.com/issue9/orm/v6/core"
)

type (
	baseStmt struct {
		engine core.Engine

		// err 用于保存在生成语句中的错误信息
		//
		// 一旦有错误生成，那么后续的调用需要保证该 err 值不会被覆盖，
		// 即所有可能改变 err 的方法中，都要先判断 err 是否为空，
		// 如果不为空，则应该立即退出函数。
		err error
	}

	queryStmt struct {
		SQLer
		baseStmt
	}

	execStmt struct {
		SQLer
		baseStmt
	}

	ddlStmt struct {
		DDLSQLer
		baseStmt
	}

	multipleDDLStmt []DDLSQLer
)

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

func (stmt *baseStmt) Dialect() core.Dialect { return stmt.engine.Dialect() }

func (stmt *baseStmt) Engine() core.Engine { return stmt.engine }

func (stmt *baseStmt) Err() error { return stmt.err }

func (stmt *baseStmt) Reset() { stmt.err = nil }

func (stmt ddlStmt) Exec() error { return stmt.ExecContext(context.Background()) }

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

// CombineSQL 合并 [SQLer.SQL] 返回的 query 和 args 参数
func (stmt *execStmt) CombineSQL() (query string, err error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return "", err
	}

	return fillArgs(query, args)
}

func (stmt *execStmt) Exec() (sql.Result, error) { return stmt.ExecContext(context.Background()) }

func (stmt *execStmt) ExecContext(ctx context.Context) (sql.Result, error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return nil, err
	}

	return stmt.Engine().ExecContext(ctx, query, args...)
}

// Prepare 预编译语句
//
// 预编译语句，参数最好采用 [sql.NamedArg] 类型。
// 在生成语句时，参数顺序会发生变化，如果采用 ? 的形式，
// 用户需要自己处理参数顺序问题，而 [sql.NamedArg] 没有这些问题。
func (stmt *execStmt) Prepare() (*core.Stmt, error) { return stmt.PrepareContext(context.Background()) }

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

// CombineSQL 将 [SQLer.SQL] 中返回的参数替换掉 query 中的占位符，
// 形成一条完整的查询语句。
func (stmt *queryStmt) CombineSQL() (query string, err error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return "", err
	}

	return fillArgs(query, args)
}

func (stmt *queryStmt) PrepareContext(ctx context.Context) (*core.Stmt, error) {
	query, _, err := stmt.SQL()
	if err != nil {
		return nil, err
	}

	return stmt.Engine().PrepareContext(ctx, query)
}

func (stmt queryStmt) Query() (*sql.Rows, error) { return stmt.QueryContext(context.Background()) }

func (stmt *queryStmt) QueryContext(ctx context.Context) (*sql.Rows, error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return nil, err
	}

	return stmt.Engine().QueryContext(ctx, query, args...)
}

// MergeDDL 合并多个 [DDLSQLer] 对象
func MergeDDL(ddl ...DDLSQLer) DDLSQLer { return multipleDDLStmt(ddl) }

func (stmt multipleDDLStmt) DDLSQL() ([]string, error) {
	queries := make([]string, 0, len(stmt))

	for _, d := range stmt {
		q, e := d.DDLSQL()
		if e != nil {
			return nil, e
		}
		queries = append(queries, q...)
	}

	return queries, nil
}
