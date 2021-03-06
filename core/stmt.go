// SPDX-License-Identifier: MIT

package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
)

// Stmt 实现自定义的 Stmt 实例
//
// 功能与 sql.Stmt 完全相同，但是实现了对 sql.NamedArgs 的支持。
type Stmt struct {
	*sql.Stmt
	orders map[string]int
}

// NewStmt 声明 Stmt 实例
//
// 如果 orders 为空，则 Stmt 的表现和 sql.Stmt 是完全相同的，
// 如果不为空，则可以处理 sql.NamedArg 类型的参数。
func NewStmt(stmt *sql.Stmt, orders map[string]int) *Stmt {
	ret := &Stmt{Stmt: stmt}

	if len(orders) == 0 {
		return ret
	}

	vals := make([]int, 0, len(orders))
	for _, v := range orders {
		vals = append(vals, v)
	}
	sort.Ints(vals)

	for k, v := range vals {
		if k != v {
			panic(fmt.Sprintf("orders 并不是连续的参数，缺少了 %d", k))
		}
	}

	ret.orders = orders
	return ret
}

// Close 关闭 Stmt 实例
func (stmt *Stmt) Close() error {
	stmt.orders = nil
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

var errArgsNotMatch = errors.New("参数数量不匹配")

func (stmt *Stmt) buildArgs(args []interface{}) ([]interface{}, error) {
	if len(stmt.orders) == 0 {
		return args, nil
	}

	if len(args) != len(stmt.orders) {
		return nil, errArgsNotMatch
	}

	ret := make([]interface{}, len(args))

	for index, arg := range args {
		named, ok := arg.(sql.NamedArg)
		if !ok {
			return nil, fmt.Errorf("第 %d 个参数并非是 sql.NamedArg 类型", index)
		}

		i, found := stmt.orders[named.Name]
		if !found {
			return nil, fmt.Errorf("参数 %s 并不存在于预编译内容中", named.Name)
		}
		ret[i] = named.Value
	}

	return ret, nil
}
