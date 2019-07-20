// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package stmt 实现自定义的 Stmt 实例
package stmt

import (
	"context"
	"database/sql"
	"fmt"
)

// Stmt 实现自定义的 Stmt 实例
//
// 功能与 sql.Stmt 完全相同。大部分的驱动都未实现命名参数，
// 只能自定义一个类似 sql.Stmt 的实现。
type Stmt struct {
	*sql.Stmt
	args map[string]int
}

// New 声明 Stmt 实例
func New(stmt *sql.Stmt, args map[string]int) *Stmt {
	return &Stmt{
		Stmt: stmt,
		args: args,
	}
}

// Close 关闭 Stmt 实例
func (stmt *Stmt) Close() error {
	stmt.args = nil
	return stmt.Stmt.Close()
}

// Exec 以指定的参数执行预编译的语句
func (stmt *Stmt) Exec(args ...interface{}) (sql.Result, error) {
	return stmt.ExecContext(context.Background(), args...)
}

// ExecContext 以指定的参数执行预编译的语句
func (stmt *Stmt) ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error) {
	args, err := stmt.buildArgs(args)
	if err != nil {
		return nil, err
	}
	return stmt.Stmt.ExecContext(ctx, args...)
}

// Query 以指定的参数执行预编译的语句
func (stmt *Stmt) Query(args ...interface{}) (*sql.Rows, error) {
	return stmt.QueryContext(context.Background(), args...)
}

// QueryContext 以指定的参数执行预编译的语句
func (stmt *Stmt) QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error) {
	args, err := stmt.buildArgs(args)
	if err != nil {
		return nil, err
	}
	return stmt.Stmt.QueryContext(context.Background(), args...)
}

// QueryRow 以指定的参数执行预编译的语句
func (stmt *Stmt) QueryRow(args ...interface{}) *sql.Row {
	return stmt.QueryRowContext(context.Background(), args...)
}

// QueryRowContext 以指定的参数执行预编译的语句
func (stmt *Stmt) QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row {
	args, err := stmt.buildArgs(args)
	if err != nil {
		panic(err)
	}

	return stmt.Stmt.QueryRowContext(context.Background(), args...)
}

func (stmt *Stmt) buildArgs(args []interface{}) ([]interface{}, error) {
	if len(stmt.args) == 0 {
		return args, nil
	}

	ret := make([]interface{}, len(args))

	for index, arg := range args {
		named, ok := arg.(sql.NamedArg)
		if !ok {
			return nil, fmt.Errorf("%d is not sql.namedArg", index)
		}

		i, found := stmt.args[named.Name]
		if !found {
			return nil, fmt.Errorf("%s not found", named.Name)
		}
		ret[i] = named.Value
	}

	return ret, nil
}
