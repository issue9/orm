// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
	"reflect"
)

// AddColumnStmt 添加列
type AddColumnStmt struct {
	engine  Engine
	dialect Dialect

	table  string
	column *Column
}

// DropColumnStmt 删除列
type DropColumnStmt struct {
	engine Engine
	table  string
	column string
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

// SQL 获取 SQL 语句以及对应的参数
func (stmt *AddColumnStmt) SQL() (string, []interface{}, error) {
	if stmt.table == "" {
		return "", nil, ErrTableIsEmpty
	}

	typ, err := stmt.dialect.SQLType(stmt.column)
	if err != nil {
		return "", nil, err
	}

	buf := New("ALTER TABLE ").
		WriteString(stmt.table).
		WriteString(" ADD ").
		WriteString(stmt.column.Name).
		WriteByte(' ').
		WriteString(typ)

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
