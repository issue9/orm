// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package sqlbuilder 提供一套通过字符串拼接来构成 SQL 语句的工具
//
// sqlbuilder 提供了部分 *Hooker 的接口，
// 用于处理大部分数据都有标准实现而只有某个数据库采用了非标准模式的。
package sqlbuilder

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/issue9/orm/v6/core"
)

type (
	// SQLBuilder 提供了 sqlbuilder 下的各类语句的创建方法
	SQLBuilder struct {
		engine core.Engine
	}

	// SQLer 定义 SQL 语句的基本接口
	SQLer interface {
		// SQL 将当前实例转换成 SQL 语句返回
		//
		// query 表示 SQL 语句，而 args 表示语句各个参数占位符对应的参数值。
		SQL() (query string, args []any, err error)
	}

	// DDLSQLer SQL 中 DDL 语句的基本接口
	//
	// 大部分数据的 DDL 操作是有多条语句组成，比如 CREATE TABLE
	// 可能包含了额外的定义信息。
	DDLSQLer interface {
		DDLSQL() ([]string, error)
	}

	ExecStmt interface {
		SQLer
		Prepare() (*core.Stmt, error)
		PrepareContext(ctx context.Context) (*core.Stmt, error)
		Exec() (sql.Result, error)
		ExecContext(ctx context.Context) (sql.Result, error)
	}

	DDLStmt interface {
		DDLSQLer
		Exec() error
		ExecContext(ctx context.Context) error
	}
)

// New 声明 SQLBuilder 实例
//
// tablePrefix 表名前缀；
func New(e core.Engine) *SQLBuilder {
	return &SQLBuilder{engine: e}
}

// SyntaxError 返回语法错误的信息
//
// typ 表示语句的类型，比如 SELECT、UPDATE 等；
// msg 为具体的错误信息；
func SyntaxError(typ string, msg any) error {
	return fmt.Errorf("在 %s 语句中存在语法错误 %s", typ, msg)
}
