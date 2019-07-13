// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
	"strings"
)

type baseStmt struct {
	dialect  Dialect
	engine   Engine
	l, r     byte
	replacer *strings.Replacer
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

func newQueryStmt(e Engine, d Dialect, sql SQLer) *queryStmt {
	l, r := d.QuoteTuple()
	return &queryStmt{
		SQLer: sql,
		baseStmt: baseStmt{
			engine:   e,
			dialect:  d,
			l:        l,
			r:        r,
			replacer: strings.NewReplacer("{", string(l), "}", string(r)),
		},
	}
}

func newExecStmt(e Engine, d Dialect, sql SQLer) *execStmt {
	l, r := d.QuoteTuple()
	return &execStmt{
		SQLer: sql,
		baseStmt: baseStmt{
			engine:  e,
			dialect: d,
			l:       l,
			r:       r,
		},
	}
}

func newDDLStmt(e Engine, d Dialect, sql DDLSQLer) *ddlStmt {
	l, r := d.QuoteTuple()
	return &ddlStmt{
		DDLSQLer: sql,
		baseStmt: baseStmt{
			engine:  e,
			dialect: d,
			l:       l,
			r:       r,
		},
	}
}

func (stmt *baseStmt) Dialect() Dialect {
	return stmt.dialect
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
		if qs[k], err = stmt.Dialect().SQL(v); err != nil {
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

	if query, err = stmt.Dialect().SQL(query); err != nil {
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

	if query, err = stmt.Dialect().SQL(query); err != nil {
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

	if query, err = stmt.Dialect().SQL(query); err != nil {
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

	if query, err = stmt.Dialect().SQL(query); err != nil {
		return nil, err
	}

	return stmt.Engine().QueryContext(ctx, query, args...)
}