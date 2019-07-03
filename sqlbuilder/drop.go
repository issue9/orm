// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import "context"

// DropTableStmt 删除表语句
type DropTableStmt struct {
	engine Engine
	tables []string
}

// DropTable 声明一条删除表的语句
func DropTable(e Engine) *DropTableStmt {
	return &DropTableStmt{
		engine: e,
	}
}

// Table 指定表名。
//
// 多次指定，则会删除多个表
func (stmt *DropTableStmt) Table(table ...string) *DropTableStmt {
	if stmt.tables == nil {
		stmt.tables = table
		return stmt
	}

	stmt.tables = append(stmt.tables, table...)
	return stmt
}

// DDLSQL 获取 SQL 语句以及对应的参数
func (stmt *DropTableStmt) DDLSQL() ([]string, error) {
	if len(stmt.tables) == 0 {
		return nil, ErrTableIsEmpty
	}

	qs := make([]string, 0, len(stmt.tables))

	for _, table := range stmt.tables {
		buf := New("DROP TABLE IF EXISTS ")
		buf.WriteString(table)
		qs = append(qs, buf.String())
	}
	return qs, nil
}

// Reset 重置
func (stmt *DropTableStmt) Reset() {
	stmt.tables = stmt.tables[:0]
}

// Exec 执行 SQL 语句
func (stmt *DropTableStmt) Exec() error {
	return stmt.ExecContext(context.Background())
}

// ExecContext 执行 SQL 语句
func (stmt *DropTableStmt) ExecContext(ctx context.Context) error {
	return ddlExecContext(ctx, stmt.engine, stmt)
}

// DropConstraintStmtHooker DropConstraintStmt.DDLSQL 的钩子函数
type DropConstraintStmtHooker interface {
	DropConstraintStmtHook(*DropConstraintStmt) ([]string, error)
}

// DropConstraintStmt 删除约束
type DropConstraintStmt struct {
	engine  Engine
	dialect Dialect

	TableName string
	Name      string
}

// DropConstraint 声明一条删除表约束的语句
func DropConstraint(e Engine, d Dialect) *DropConstraintStmt {
	return &DropConstraintStmt{
		engine:  e,
		dialect: d,
	}
}

// Table 指定表名。
// 重复指定，会覆盖之前的。
func (stmt *DropConstraintStmt) Table(table string) *DropConstraintStmt {
	stmt.TableName = table
	return stmt
}

// Constraint 指定需要删除的列
func (stmt *DropConstraintStmt) Constraint(cont string) *DropConstraintStmt {
	stmt.Name = cont
	return stmt
}

// DDLSQL 获取 SQL 语句以及对应的参数
func (stmt *DropConstraintStmt) DDLSQL() ([]string, error) {
	if stmt.TableName == "" {
		return nil, ErrTableIsEmpty
	}

	if hook, ok := stmt.dialect.(DropConstraintStmtHooker); ok {
		return hook.DropConstraintStmtHook(stmt)
	}

	buf := New("ALTER TABLE {").
		WriteString(stmt.TableName).
		WriteString("} DROP CONSTRAINT {").
		WriteString(stmt.Name).
		WriteByte('}')
	return []string{buf.String()}, nil
}

// Reset 重置
func (stmt *DropConstraintStmt) Reset() {
	stmt.TableName = ""
	stmt.Name = ""
}

// Exec 执行 SQL 语句
func (stmt *DropConstraintStmt) Exec() error {
	return stmt.ExecContext(context.Background())
}

// ExecContext 执行 SQL 语句
func (stmt *DropConstraintStmt) ExecContext(ctx context.Context) error {
	return ddlExecContext(ctx, stmt.engine, stmt)
}
