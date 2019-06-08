// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
)

// InsertStmt 表示插入操作的 SQL 语句
type InsertStmt struct {
	engine  Engine
	dialect Dialect
	table   string
	cols    []string
	args    [][]interface{}
}

// Insert 声明一条插入语句
func Insert(e Engine, d Dialect) *InsertStmt {
	return &InsertStmt{
		engine:  e,
		dialect: d,
		cols:    make([]string, 0, 10),
		args:    make([][]interface{}, 0, 10),
	}
}

// Table 指定表名
func (stmt *InsertStmt) Table(table string) *InsertStmt {
	stmt.table = table
	return stmt
}

// KeyValue 指定键值对
//
// 当通过 Values() 指定多行数据时，再使用 KeyValue 会出错
func (stmt *InsertStmt) KeyValue(col string, val interface{}) *InsertStmt {
	if len(stmt.args) > 1 {
		panic("多列模式，不能调用 KeyValue 函数")
	}

	if len(stmt.args) == 0 {
		stmt.args = append(stmt.args, []interface{}{})
	}

	stmt.cols = append(stmt.cols, col)
	stmt.args[0] = append(stmt.args[0], val)

	return stmt
}

// Columns 指定插入的列，多次指定，之前的会被覆盖。
func (stmt *InsertStmt) Columns(cols ...string) *InsertStmt {
	stmt.cols = append(stmt.cols, cols...)
	return stmt
}

// Values 指定需要插入的值
//
// NOTE: vals 传入时，并不会被解压
func (stmt *InsertStmt) Values(vals ...interface{}) *InsertStmt {
	stmt.args = append(stmt.args, vals)
	return stmt
}

// Reset 重置语句
func (stmt *InsertStmt) Reset() {
	stmt.table = ""
	stmt.cols = stmt.cols[:0]
	stmt.args = stmt.args[:0]
}

// SQL 获取 SQL 的语句及参数部分
func (stmt *InsertStmt) SQL() (string, []interface{}, error) {
	if stmt.table == "" {
		return "", nil, ErrTableIsEmpty
	}

	if len(stmt.cols) == 0 {
		return "", nil, ErrColumnsIsEmpty
	}

	if len(stmt.args) == 0 {
		return "", nil, ErrValueIsEmpty
	}

	for _, vals := range stmt.args {
		if len(vals) != len(stmt.cols) {
			return "", nil, ErrArgsNotMatch
		}
	}

	buffer := New("INSERT INTO ")
	buffer.WriteString(stmt.table)

	buffer.WriteByte('(')
	for _, col := range stmt.cols {
		buffer.WriteString(col)
		buffer.WriteByte(',')
	}
	buffer.TruncateLast(1)
	buffer.WriteByte(')')

	args := make([]interface{}, 0, len(stmt.cols)*len(stmt.args))
	buffer.WriteString(" VALUES ")
	for _, vals := range stmt.args {
		buffer.WriteByte('(')
		for _, v := range vals {
			if named, ok := v.(sql.NamedArg); ok && named.Name != "" {
				buffer.WriteByte('@')
				buffer.WriteString(named.Name)
			} else {
				buffer.WriteByte('?')
			}
			buffer.WriteByte(',')
			args = append(args, v)
		}
		buffer.TruncateLast(1) // 去掉最后的逗号
		buffer.WriteString("),")
	}
	buffer.TruncateLast(1)

	return buffer.String(), args, nil
}

// Exec 执行 SQL 语句
func (stmt *InsertStmt) Exec() (sql.Result, error) {
	return stmt.ExecContext(context.Background())
}

// ExecContext 执行 SQL 语句
func (stmt *InsertStmt) ExecContext(ctx context.Context) (sql.Result, error) {
	return execContext(ctx, stmt.engine, stmt)
}

// Prepare 预编译
func (stmt *InsertStmt) Prepare() (*sql.Stmt, error) {
	return stmt.PrepareContext(context.Background())
}

// PrepareContext 预编译
func (stmt *InsertStmt) PrepareContext(ctx context.Context) (*sql.Stmt, error) {
	return prepareContext(ctx, stmt.engine, stmt)
}

// LastInsertID 执行 SQL 语句
//
// 并根据表名和自增列 ID 返回当前行的自增 ID 值。
func (stmt *InsertStmt) LastInsertID(table, col string) (int64, error) {
	return stmt.LastInsertIDContext(context.Background(), table, col)
}

// LastInsertIDContext 执行 SQL 语句
//
// 并根据表名和自增列 ID 返回当前行的自增 ID 值。
func (stmt *InsertStmt) LastInsertIDContext(ctx context.Context, table, col string) (int64, error) {
	sql, append := stmt.dialect.LastInsertID(stmt.table, col)
	if sql == "" {
		rslt, err := stmt.ExecContext(ctx)
		if err != nil {
			return 0, err
		}

		return rslt.LastInsertId()
	}

	query, args, err := stmt.SQL()
	if err != nil {
		return 0, err
	}
	if !append {
		_, err = stmt.ExecContext(ctx)
		if err != nil {
			return 0, err
		}
	} else {
		query += sql
	}

	var id int64
	err = stmt.engine.QueryRowContext(ctx, query, args...).Scan(&id)
	return id, err
}
