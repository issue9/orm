// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"

	"github.com/issue9/orm/types"
)

// DeleteStmt 表示删除操作的 SQL 语句
type DeleteStmt struct {
	engine types.Engine
	table  string
	where  *WhereStmt
}

// Delete 声明一条删除语句
func Delete(e types.Engine, table string) *DeleteStmt {
	return &DeleteStmt{
		engine: e,
		table:  table,
		where:  newWhereStmt(),
	}
}

// Table 指定表名
func (stmt *DeleteStmt) Table(table string) *DeleteStmt {
	stmt.table = table
	return stmt
}

// SQL 获取 SQL 语句，以及其参数对应的具体值
func (stmt *DeleteStmt) SQL() (string, []interface{}, error) {
	if stmt.table == "" {
		return "", nil, ErrTableIsEmpty
	}

	query, args, err := stmt.where.SQL()
	if err != nil {
		return "", nil, err
	}

	return "DELETE FROM " + stmt.table + query, args, nil
}

// Reset 重置语句
func (stmt *DeleteStmt) Reset() {
	stmt.table = ""
	stmt.where.Reset()
}

// WhereStmt 实现 WhereStmter 接口
func (stmt *DeleteStmt) WhereStmt() *WhereStmt {
	return stmt.where
}

// Where 指定 where 语句
func (stmt *DeleteStmt) Where(cond string, args ...interface{}) *DeleteStmt {
	return stmt.And(cond, args...)
}

// And 指定 where ... AND ... 语句
func (stmt *DeleteStmt) And(cond string, args ...interface{}) *DeleteStmt {
	stmt.where.And(cond, args...)
	return stmt
}

// Or 指定 where ... OR ... 语句
func (stmt *DeleteStmt) Or(cond string, args ...interface{}) *DeleteStmt {
	stmt.where.Or(cond, args...)
	return stmt
}

// Exec 执行 SQL 语句
func (stmt *DeleteStmt) Exec() (sql.Result, error) {
	return exec(stmt.engine, stmt)
}

// ExecContext 执行 SQL 语句
func (stmt *DeleteStmt) ExecContext(ctx context.Context) (sql.Result, error) {
	return execContext(ctx, stmt.engine, stmt)
}

// Prepare 预编译
func (stmt *DeleteStmt) Prepare() (*sql.Stmt, error) {
	return prepare(stmt.engine, stmt)
}

// PrepareContext 预编译
func (stmt *DeleteStmt) PrepareContext(ctx context.Context) (*sql.Stmt, error) {
	return prepareContext(ctx, stmt.engine, stmt)
}
