// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
)

// DropTableStmt 删除表语句
type DropTableStmt struct {
	engine Engine
	table  string
}

// DropTable 声明一条删除表的语句
func DropTable(e Engine) *DropTableStmt {
	return &DropTableStmt{
		engine: e,
	}
}

// Table 指定表名。
// 重复指定，会覆盖之前的。
func (stmt *DropTableStmt) Table(table string) *DropTableStmt {
	stmt.table = table
	return stmt
}

// DDLSQL 获取 SQL 语句以及对应的参数
func (stmt *DropTableStmt) DDLSQL() ([]string, error) {
	if stmt.table == "" {
		return nil, ErrTableIsEmpty
	}

	buf := New("DROP TABLE IF EXISTS {")
	buf.WriteString(stmt.table)
	return []string{buf.String()}, nil
}

// Reset 重置
func (stmt *DropTableStmt) Reset() {
	stmt.table = ""
}

// Exec 执行 SQL 语句
func (stmt *DropTableStmt) Exec() error {
	return stmt.ExecContext(context.Background())
}

// ExecContext 执行 SQL 语句
func (stmt *DropTableStmt) ExecContext(ctx context.Context) error {
	return ddlExecContext(ctx, stmt.engine, stmt)
}

// DropColumnStmt 删除表语句
type DropColumnStmt struct {
	engine Engine
	table  string
	column string
}

// DropColumn 声明一条删除表的语句
func DropColumn(e Engine) *DropColumnStmt {
	return &DropColumnStmt{
		engine: e,
	}
}

// Table 指定表名。
// 重复指定，会覆盖之前的。
func (stmt *DropColumnStmt) Table(table string) *DropColumnStmt {
	stmt.table = table
	return stmt
}

// Column 指定需要删除的列
func (stmt *DropColumnStmt) Column(col string) *DropColumnStmt {
	stmt.column = col
	return stmt
}

// SQL 获取 SQL 语句以及对应的参数
func (stmt *DropColumnStmt) SQL() (string, []interface{}, error) {
	if stmt.table == "" {
		return "", nil, ErrTableIsEmpty
	}

	buf := New("ALTER TABLE {")
	buf.WriteString(stmt.table)
	buf.WriteString("} DROP COLUMN {")
	buf.WriteString(stmt.column)
	buf.WriteString("};")
	return buf.String(), nil, nil
}

// Reset 重置
func (stmt *DropColumnStmt) Reset() {
	stmt.table = ""
	stmt.column = ""
}

// Exec 执行 SQL 语句
func (stmt *DropColumnStmt) Exec() (sql.Result, error) {
	return stmt.ExecContext(context.Background())
}

// ExecContext 执行 SQL 语句
func (stmt *DropColumnStmt) ExecContext(ctx context.Context) (sql.Result, error) {
	return execContext(ctx, stmt.engine, stmt)
}
