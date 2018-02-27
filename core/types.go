// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package core

import (
	"context"
	"database/sql"
)

// Engine 是 DB 与 Tx 的共有接口。
type Engine interface {
	// 获取与之关联的 Dialect 接口。
	Dialect() Dialect

	// 执行一条查询语句，并返回相应的 sql.Rows 实例。
	// 功能基本上等同于标准库 database/sql 的 DB.Query()
	//
	// 如：以下内容，在 replace 为 false 时，将原样输出，
	// 否则将被转换成以下字符串(以 mysql 为例，假设当前的 prefix 为 p_)
	//  select * from #user where {group}=1
	//  // 转换后
	//  select * from prefix_user where `group`=1
	Query(query string, args ...interface{}) (*sql.Rows, error)

	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

	// 功能等同于 database/sql 的 DB.Exec()。
	Exec(query string, args ...interface{}) (sql.Result, error)

	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// 功能等同于 database/sql 的 DB.Prepare()。
	Prepare(query string) (*sql.Stmt, error)

	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// Dialect 接口用于描述与数据库相关的一些语言特性。
type Dialect interface {
	// 返回当前数据库的名称。
	Name() string

	// 返回符合当前数据库规范的引号对。
	QuoteTuple() (openQuote, closeQuote byte)

	// 根据当前的数据库，对 SQL 作调整。
	//
	// 比如占位符 postgresql 可以使用 $1 等形式。
	// 以及部分驱动可能不支持最新的命名参数，也会做调整。
	SQL(sql *StringBuilder) error

	// 生成 `LIMIT N OFFSET M` 或是相同的语意的语句。
	//
	// offset 值为一个可选参数，若不指定，则表示 `LIMIT N` 语句。
	// 返回的是对应数据库的 limit 语句以及语句中占位符对应的值。
	LimitSQL(limit int, offset ...int) (string, []interface{})

	// 输出非 AI 列的定义，必须包含末尾的分号
	NoAIColSQL(sql *StringBuilder, m *Model) error

	// 输出 AI 列的定义，必须包含末尾的分号
	AIColSQL(sql *StringBuilder, m *Model) error

	// 输出所有的约束定义，必须包含末尾的分号
	ConstraintsSQL(sql *StringBuilder, m *Model)

	// 清空表内容，重置 AI。
	// aiColumn 需要被重置的自增列列名
	TruncateTableSQL(sql *StringBuilder, tableName, aiColumn string)

	// 是否支持一次性插入多条语句
	SupportInsertMany() bool
}
