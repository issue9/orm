// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package engine [core.Engine] 的默认实现
package engine

import (
	"context"
	"database/sql"
	"strings"

	"github.com/issue9/orm/v6/core"
)

type coreEngine struct {
	engine      stdEngine
	dialect     core.Dialect
	tablePrefix string
	replacer    *strings.Replacer
	sqlLogger   func(string)
}

// [sql.DB] 与 [sql.Tx] 的最小接口
type stdEngine interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func defaultSQLLogger(string) {}

func New(e stdEngine, tablePrefix string, d core.Dialect) core.Engine {
	l, r := d.Quotes()

	return &coreEngine{
		engine:      e,
		dialect:     d,
		tablePrefix: tablePrefix,
		sqlLogger:   defaultSQLLogger,
		replacer: strings.NewReplacer(
			string(core.QuoteLeft), string(l),
			string(core.QuoteRight), string(r),
			"#", tablePrefix,
		),
	}
}

func (db *coreEngine) TablePrefix() string { return db.tablePrefix }

// Debug 指定调输出调试内容通道
//
// 如果 l 不为 nil，则每次 SQL 调用都会输出 SQL 语句，预编译的语句，仅在预编译时输出；
// 如果为 nil，则表示关闭调试。
func (db *coreEngine) Debug(l func(string)) {
	if l == nil {
		l = defaultSQLLogger
	}
	db.sqlLogger = l
}

func (db *coreEngine) Dialect() core.Dialect { return db.dialect }

func (db *coreEngine) QueryRow(query string, args ...any) *sql.Row {
	return db.QueryRowContext(context.Background(), query, args...)
}

func (db *coreEngine) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	db.sqlLogger(query)
	query, args, err := db.dialect.Fix(query, args)
	if err != nil {
		panic(err)
	}

	query = db.replacer.Replace(query)
	return db.engine.QueryRowContext(ctx, query, args...)
}

func (db *coreEngine) Query(query string, args ...any) (*sql.Rows, error) {
	return db.QueryContext(context.Background(), query, args...)
}

func (db *coreEngine) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	db.sqlLogger(query)
	query, args, err := db.Dialect().Fix(query, args)
	if err != nil {
		return nil, err
	}

	query = db.replacer.Replace(query)
	return db.engine.QueryContext(ctx, query, args...)
}

func (db *coreEngine) Exec(query string, args ...any) (sql.Result, error) {
	return db.ExecContext(context.Background(), query, args...)
}

func (db *coreEngine) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	db.sqlLogger(query)
	query, args, err := db.Dialect().Fix(query, args)
	if err != nil {
		return nil, err
	}

	query = db.replacer.Replace(query)
	return db.engine.ExecContext(ctx, query, args...)
}

func (db *coreEngine) Prepare(query string) (*core.Stmt, error) {
	return db.PrepareContext(context.Background(), query)
}

func (db *coreEngine) PrepareContext(ctx context.Context, query string) (*core.Stmt, error) {
	db.sqlLogger(query)
	query, orders, err := db.Dialect().Prepare(query)
	if err != nil {
		return nil, err
	}

	query = db.replacer.Replace(query)
	s, err := db.engine.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return core.NewStmt(s, orders), nil
}
