// SPDX-License-Identifier: MIT

// Package sqlbuilder 提供一套通过字符串拼接来构成 SQL 语句的工具
//
// sqlbuilder 提供了部分 *Hooker 的接口，用于自定义某一条语句的实现。
// 一般情况下， 如果有多个数据是遵循 SQL 标准的，只有个别有例外，
// 那么该例外的 Dialect 实现，可以同时实现 Hooker 接口， 自定义该语句的实现。
package sqlbuilder

import (
	"context"
	"database/sql"
	"errors"

	"github.com/issue9/orm/v4/core"
)

var (
	// ErrTableIsEmpty 未指定表名，任何 SQL 语句中，
	// 若未指定表名时，会返回此错误
	ErrTableIsEmpty = errors.New("表名为空")

	// ErrValueIsEmpty 在 Update 和 Insert 语句中，
	// 若未指定任何值，则返回此错误
	ErrValueIsEmpty = errors.New("值为空")

	// ErrColumnsIsEmpty 在 Insert 和 Select 语句中，
	// 若未指定任何列表，则返回此错误
	ErrColumnsIsEmpty = errors.New("未指定列")

	// ErrConstraintIsEmpty 约束名不能为空，某些需要操作约束的 SQL 会返回此值。
	ErrConstraintIsEmpty = errors.New("约束名不能为空")

	// ErrDupColumn 在 Update 中可能存在重复设置的列名。
	ErrDupColumn = errors.New("重复的列名")

	// ErrArgsNotMatch 在生成的 SQL 语句中，传递的参数与语句的占位符数量不匹配。
	ErrArgsNotMatch = errors.New("列与值的数量不匹配")

	// ErrUnknownConstraint 该约束类型不支持，或是当前环境下无法使用
	ErrUnknownConstraint = errors.New("不支持的约束类型")

	// ErrNotImplemented 部分数据库如果没有实现的功能，可以返回该错误
	ErrNotImplemented = errors.New("未实现该功能")

	// ErrConstraintType 约束类型错误
	ErrConstraintType = errors.New("约束类型错误，已经设置为其它约束")

	// ErrUnionColumnNotMatch 在 Union 中，各个 select 中的列长度不相同。
	ErrUnionColumnNotMatch = errors.New("union 列长度不相同")
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
		SQL() (query string, args []interface{}, err error)
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
func New(e core.Engine) *SQLBuilder { return &SQLBuilder{engine: e} }
