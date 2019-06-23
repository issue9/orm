// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
)

// AddColumnStmt 添加列
type AddColumnStmt struct {
	engine Engine
	table  string
	column *column
}

// DropColumnStmt 删除列
type DropColumnStmt struct {
	engine Engine
	table  string
	column string
}

// AddColumn 声明一条添加列的语句
func AddColumn(e Engine) *AddColumnStmt {
	return &AddColumnStmt{
		engine: e,
	}
}

// Table 指定表名。
// 重复指定，会覆盖之前的。
func (stmt *AddColumnStmt) Table(table string) *AddColumnStmt {
	stmt.table = table
	return stmt
}

// Column 添加列
//
// 参数信息可参考 CreateTableStmt.Column
func (stmt *AddColumnStmt) Column(name, typ string) *AddColumnStmt {
	stmt.column = &column{
		Name: name,
		Type: typ,
	}

	return stmt
}

// SQL 获取 SQL 语句以及对应的参数
func (stmt *AddColumnStmt) SQL() (string, []interface{}, error) {
	if stmt.table == "" {
		return "", nil, ErrTableIsEmpty
	}

	buf := New("ALTER TABLE ")
	buf.WriteString(stmt.table)
	buf.WriteString(" ADD ")
	buf.WriteString(stmt.column.Name)
	buf.WriteString(stmt.column.Type)
	return buf.String(), nil, nil
}

// Reset 重置
func (stmt *AddColumnStmt) Reset() {
	stmt.table = ""
	stmt.column = nil
}

// Exec 执行 SQL 语句
func (stmt *AddColumnStmt) Exec() (sql.Result, error) {
	return stmt.ExecContext(context.Background())
}

// ExecContext 执行 SQL 语句
func (stmt *AddColumnStmt) ExecContext(ctx context.Context) (sql.Result, error) {
	return execContext(ctx, stmt.engine, stmt)
}

// DropColumn 声明一条删除列的语句
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

	buf := New("ALTER TABLE ")
	buf.WriteString(stmt.table)
	buf.WriteString(" DROP COLUMN ")
	buf.WriteString(stmt.column)
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
