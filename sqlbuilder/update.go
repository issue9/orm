// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"

	"github.com/issue9/orm/core"
)

// UpdateStmt 更新语句
type UpdateStmt struct {
	engine   core.Engine
	table    string
	where    *WhereStmt
	values   map[string]interface{}
	increase map[string]interface{}
	decrease map[string]interface{}
}

// Update 声明一条 UPDATE 的 SQL 语句
func Update(e core.Engine, table string) *UpdateStmt {
	return &UpdateStmt{
		engine:   e,
		table:    table,
		where:    newWhereStmt(),
		values:   map[string]interface{}{},
		increase: map[string]interface{}{},
		decrease: map[string]interface{}{},
	}
}

// Table 指定表名
func (stmt *UpdateStmt) Table(table string) *UpdateStmt {
	stmt.table = table
	return stmt
}

// Set 设置值，若 col 相同，则会覆盖
func (stmt *UpdateStmt) Set(col string, val interface{}) *UpdateStmt {
	stmt.values[col] = val
	return stmt
}

// Increase 给列增加值
func (stmt *UpdateStmt) Increase(col string, val interface{}) *UpdateStmt {
	stmt.increase[col] = val
	return stmt
}

// Decrease 给钱减少值
func (stmt *UpdateStmt) Decrease(col string, val interface{}) *UpdateStmt {
	stmt.decrease[col] = val
	return stmt
}

// WhereStmt 实现 WhereStmter 接口
func (stmt *UpdateStmt) WhereStmt() *WhereStmt {
	return stmt.where
}

// Where 指定 where 语句
func (stmt *UpdateStmt) Where(cond string, args ...interface{}) *UpdateStmt {
	return stmt.And(cond, args...)
}

// And 指定 where ... AND ... 语句
func (stmt *UpdateStmt) And(cond string, args ...interface{}) *UpdateStmt {
	stmt.where.And(cond, args...)
	return stmt
}

// Or 指定 where ... OR ... 语句
func (stmt *UpdateStmt) Or(cond string, args ...interface{}) *UpdateStmt {
	stmt.where.Or(cond, args...)
	return stmt
}

// Reset 重置语句
func (stmt *UpdateStmt) Reset() {
	stmt.table = ""
	stmt.where.Reset()
	stmt.values = map[string]interface{}{}
	stmt.increase = map[string]interface{}{}
	stmt.decrease = map[string]interface{}{}
}

// SQL 获取 SQL 语句以及对应的参数
func (stmt *UpdateStmt) SQL() (string, []interface{}, error) {
	if stmt.table == "" {
		return "", nil, ErrTableIsEmpty
	}

	if len(stmt.values) == 0 && len(stmt.increase) == 0 && len(stmt.decrease) == 0 {
		return "", nil, ErrValueIsEmpty
	}

	buf := core.NewStringBuilder("UPDATE ")
	buf.WriteString(stmt.table)
	buf.WriteString(" SET ")

	args := make([]interface{}, 0, len(stmt.values))

	for col, val := range stmt.values {
		buf.WriteString(col)
		buf.WriteByte('=')
		if named, ok := val.(sql.NamedArg); ok && named.Name != "" {
			buf.WriteByte('@')
			buf.WriteString(named.Name)
		} else {
			buf.WriteByte('?')
		}
		buf.WriteByte(',')
		args = append(args, val)
	}

	for col, val := range stmt.increase {
		buf.WriteString(col)
		buf.WriteByte('=')
		buf.WriteString(col)
		buf.WriteByte('+')
		if named, ok := val.(sql.NamedArg); ok && named.Name != "" {
			buf.WriteByte('@')
			buf.WriteString(named.Name)
		} else {
			buf.WriteByte('?')
		}
		buf.WriteByte(',')
		args = append(args, val)
	}

	for col, val := range stmt.decrease {
		buf.WriteString(col)
		buf.WriteByte('=')
		buf.WriteString(col)
		buf.WriteByte('-')
		if named, ok := val.(sql.NamedArg); ok && named.Name != "" {
			buf.WriteByte('@')
			buf.WriteString(named.Name)
		} else {
			buf.WriteByte('?')
		}
		buf.WriteByte(',')
		args = append(args, val)
	}

	// 等所有的 SET 部分内容都完成了，去掉最后的逗号
	buf.TruncateLast(1)

	wq, wa, err := stmt.where.SQL()
	if err != nil {
		return "", nil, err
	}

	buf.WriteString(wq)
	args = append(args, wa...)
	return buf.String(), args, nil
}

// Exec 执行 SQL 语句
func (stmt *UpdateStmt) Exec() (sql.Result, error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return stmt.engine.Exec(query, args...)
}

// ExecContext 执行 SQL 语句
func (stmt *UpdateStmt) ExecContext(ctx context.Context) (sql.Result, error) {
	query, args, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return stmt.engine.ExecContext(ctx, query, args...)
}

// Prepare 预编译
func (stmt *UpdateStmt) Prepare() (*sql.Stmt, error) {
	query, _, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return stmt.engine.Prepare(query)
}

// PrepareContext 预编译
func (stmt *UpdateStmt) PrepareContext(ctx context.Context) (*sql.Stmt, error) {
	query, _, err := stmt.SQL()
	if err != nil {
		return nil, err
	}
	return stmt.engine.PrepareContext(ctx, query)
}
