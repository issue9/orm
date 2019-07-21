// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
)

type baseStmt struct {
	engine Engine
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

func newQueryStmt(e Engine, sql SQLer) *queryStmt {
	return &queryStmt{
		SQLer: sql,
		baseStmt: baseStmt{
			engine: e,
		},
	}
}

func newExecStmt(e Engine, sql SQLer) *execStmt {
	return &execStmt{
		SQLer: sql,
		baseStmt: baseStmt{
			engine: e,
		},
	}
}

func newDDLStmt(e Engine, sql DDLSQLer) *ddlStmt {
	return &ddlStmt{
		DDLSQLer: sql,
		baseStmt: baseStmt{
			engine: e,
		},
	}
}

func (stmt *baseStmt) Dialect() Dialect {
	return stmt.engine.Dialect()
}

func (stmt *baseStmt) Engine() Engine {
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

	for k, v := range qs {
		if qs[k], _, err = stmt.Dialect().SQL(v, nil); err != nil {
			return err
		}
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

func (stmt *execStmt) Prepare() (*sql.Stmt, error) {
	return stmt.PrepareContext(context.Background())
}

func (stmt *execStmt) PrepareContext(ctx context.Context) (*sql.Stmt, error) {
	query, _, err := stmt.SQL()
	if err != nil {
		return nil, err
	}

	return stmt.Engine().PrepareContext(ctx, query)
}

func (stmt *queryStmt) Prepare() (*sql.Stmt, error) {
	return stmt.PrepareContext(context.Background())
}

func (stmt *queryStmt) PrepareContext(ctx context.Context) (*sql.Stmt, error) {
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
