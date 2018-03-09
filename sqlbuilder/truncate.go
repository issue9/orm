// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
)

// TruncateStmt 清除数据表中的数据并重置自增列
type TruncateStmt struct {
	engine  Engine
	dialect Dialect
	table   string
	aiCol   string
}

// Truncate 声明一个 TruncateStmt
func Truncate(e Engine, d Dialect) *TruncateStmt {
	return &TruncateStmt{
		engine:  e,
		dialect: d,
	}
}

// Table 指定表名
func (stmt *TruncateStmt) Table(tbl string) *TruncateStmt {
	stmt.table = tbl
	return stmt
}

// AI 指定 AI 列
func (stmt *TruncateStmt) AI(col string) *TruncateStmt {
	stmt.aiCol = col
	return stmt
}

// Reset 重置
func (stmt *TruncateStmt) Reset() {
	stmt.aiCol = ""
	stmt.table = ""
}

// SQL 获取 SQL 语句及对应的参数
func (stmt *TruncateStmt) SQL() (string, []interface{}, error) {
	return stmt.dialect.TruncateTableSQL(stmt.table, stmt.aiCol), nil, nil
}

// Exec 执行 SQL 语句
func (stmt *TruncateStmt) Exec() (sql.Result, error) {
	return exec(stmt.engine, stmt)
}

// ExecContext 执行 SQL 语句
func (stmt *TruncateStmt) ExecContext(ctx context.Context) (sql.Result, error) {
	return execContext(ctx, stmt.engine, stmt)
}

// Prepare 预编译
func (stmt *TruncateStmt) Prepare() (*sql.Stmt, error) {
	return prepare(stmt.engine, stmt)
}

// PrepareContext 预编译
func (stmt *TruncateStmt) PrepareContext(ctx context.Context) (*sql.Stmt, error) {
	return prepareContext(ctx, stmt.engine, stmt)
}
