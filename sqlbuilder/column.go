// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"reflect"
)

// AddColumnStmt 添加列
type AddColumnStmt struct {
	engine  Engine
	dialect Dialect

	table  string
	column *Column
}

// AddColumn 声明一条添加列的语句
func AddColumn(e Engine, d Dialect) *AddColumnStmt {
	return &AddColumnStmt{
		engine:  e,
		dialect: d,
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
func (stmt *AddColumnStmt) Column(name string, goType reflect.Type, nullable, hasDefault bool, def interface{}, length ...int) *AddColumnStmt {
	col := newColumn(name, goType, false, nullable, hasDefault, def, length...)

	stmt.column = col

	return stmt
}

// DDLSQL 获取 SQL 语句以及对应的参数
func (stmt *AddColumnStmt) DDLSQL() ([]string, error) {
	if stmt.table == "" {
		return nil, ErrTableIsEmpty
	}

	typ, err := stmt.dialect.SQLType(stmt.column)
	if err != nil {
		return nil, err
	}

	buf := New("ALTER TABLE ").
		WriteString(stmt.table).
		WriteString(" ADD ").
		WriteString(stmt.column.Name).
		WriteByte(' ').
		WriteString(typ)

	return []string{buf.String()}, nil
}

// Reset 重置
func (stmt *AddColumnStmt) Reset() {
	stmt.table = ""
	stmt.column = nil
}

// Exec 执行 SQL 语句
func (stmt *AddColumnStmt) Exec() error {
	return stmt.ExecContext(context.Background())
}

// ExecContext 执行 SQL 语句
func (stmt *AddColumnStmt) ExecContext(ctx context.Context) error {
	return ddlExecContext(ctx, stmt.engine, stmt)
}

// DropColumnStmtHooker DropColumnStmt.DDLSQL 的钩子函数
type DropColumnStmtHooker interface {
	DropColumnStmtHook(*DropColumnStmt) ([]string, error)
}

// DropColumnStmt 删除列
type DropColumnStmt struct {
	engine  Engine
	dialect Dialect

	TableName  string
	ColumnName string
}

// DropColumn 声明一条删除列的语句
func DropColumn(e Engine, d Dialect) *DropColumnStmt {
	return &DropColumnStmt{
		engine:  e,
		dialect: d,
	}
}

// Table 指定表名。
// 重复指定，会覆盖之前的。
func (stmt *DropColumnStmt) Table(table string) *DropColumnStmt {
	stmt.TableName = table
	return stmt
}

// Column 指定需要删除的列
func (stmt *DropColumnStmt) Column(col string) *DropColumnStmt {
	stmt.ColumnName = col
	return stmt
}

// DDLSQL 获取 SQL 语句以及对应的参数
func (stmt *DropColumnStmt) DDLSQL() ([]string, error) {
	if stmt.TableName == "" {
		return nil, ErrTableIsEmpty
	}

	if hook, ok := stmt.dialect.(DropColumnStmtHooker); ok {
		return hook.DropColumnStmtHook(stmt)
	}

	buf := New("ALTER TABLE ")
	buf.WriteString(stmt.TableName)
	buf.WriteString(" DROP COLUMN ")
	buf.WriteString(stmt.ColumnName)
	return []string{buf.String()}, nil
}

// Reset 重置
func (stmt *DropColumnStmt) Reset() {
	stmt.TableName = ""
	stmt.ColumnName = ""
}

// Exec 执行 SQL 语句
func (stmt *DropColumnStmt) Exec() error {
	return stmt.ExecContext(context.Background())
}

// ExecContext 执行 SQL 语句
func (stmt *DropColumnStmt) ExecContext(ctx context.Context) error {
	return ddlExecContext(ctx, stmt.engine, stmt)
}
