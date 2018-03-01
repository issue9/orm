// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"

	"github.com/issue9/orm/core"
)

// DropTableStmt 删除表语句
type DropTableStmt struct {
	engine core.Engine
	table  string
}

// DropTable 声明一条删除表的语句
func DropTable(e core.Engine, table string) *DropTableStmt {
	return &DropTableStmt{
		engine: e,
		table:  table,
	}
}

// Table 指定表名。
// 重复指定，会覆盖之前的。
func (stmt *DropTableStmt) Table(table string) *DropTableStmt {
	stmt.table = table
	return stmt
}

// SQL 获取 SQL 语句以及对应的参数
func (stmt *DropTableStmt) SQL() (string, []interface{}, error) {
	if stmt.table == "" {
		return "", nil, ErrTableIsEmpty
	}

	buf := core.NewStringBuilder("DROP TABLE IF EXISTS ")
	buf.WriteString(stmt.table)
	return buf.String(), nil, nil
}

// Reset 重置
func (stmt *DropTableStmt) Reset() {
	stmt.table = ""
}

// Exec 执行 SQL 语句
func (stmt *DropTableStmt) Exec() (sql.Result, error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return stmt.engine.Exec(query, args...)
}

// ExecContext 执行 SQL 语句
func (stmt *DropTableStmt) ExecContext(ctx context.Context) (sql.Result, error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return stmt.engine.ExecContext(ctx, query, args...)
}

// Prepare 预编译
func (stmt *DropTableStmt) Prepare() (*sql.Stmt, error) {
	query, _, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return stmt.engine.Prepare(query)
}

// PrepareContext 预编译
func (stmt *DropTableStmt) PrepareContext(ctx context.Context) (*sql.Stmt, error) {
	query, _, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return stmt.engine.PrepareContext(ctx, query)
}
