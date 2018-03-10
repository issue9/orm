// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
)

// CreateIndexStmt 创建索引的语句
type CreateIndexStmt struct {
	engine Engine
	table  string
	name   string   // 索引名称
	cols   []string // 索引列
}

// CreateIndex 声明一条 CrateIndexStmt 语句
func CreateIndex(e Engine) *CreateIndexStmt {
	return &CreateIndexStmt{
		engine: e,
	}
}

// Table 指定表名
func (stmt *CreateIndexStmt) Table(tbl string) *CreateIndexStmt {
	stmt.table = tbl
	return stmt
}

// Name 指定约束名
func (stmt *CreateIndexStmt) Name(col string) *CreateIndexStmt {
	stmt.name = col
	return stmt
}

// Columns 列名
func (stmt *CreateIndexStmt) Columns(col ...string) *CreateIndexStmt {
	if stmt.cols == nil {
		stmt.cols = make([]string, 0, len(col))
	}
	stmt.cols = append(stmt.cols, col...)

	return stmt
}

// SQL 生成 SQL 语句
func (stmt *CreateIndexStmt) SQL() (string, []interface{}, error) {
	sql := New("CREATE INDEX ").
		WriteString(stmt.name).
		WriteString(" ON ").
		WriteString(stmt.table).WriteByte('(')
	for _, col := range stmt.cols {
		sql.WriteString(col).WriteByte(',')
	}
	sql.TruncateLast(1).WriteByte(')')

	return sql.String(), nil, nil
}

// Reset 重置
func (stmt *CreateIndexStmt) Reset() {
	stmt.table = ""
	stmt.cols = stmt.cols[:0]
	stmt.name = ""
}

// Exec 执行 SQL 语句
func (stmt *CreateIndexStmt) Exec() (sql.Result, error) {
	return exec(stmt.engine, stmt)
}

// ExecContext 执行 SQL 语句
func (stmt *CreateIndexStmt) ExecContext(ctx context.Context) (sql.Result, error) {
	return execContext(ctx, stmt.engine, stmt)
}

// Prepare 预编译
func (stmt *CreateIndexStmt) Prepare() (*sql.Stmt, error) {
	return prepare(stmt.engine, stmt)
}

// PrepareContext 预编译
func (stmt *CreateIndexStmt) PrepareContext(ctx context.Context) (*sql.Stmt, error) {
	return prepareContext(ctx, stmt.engine, stmt)
}
