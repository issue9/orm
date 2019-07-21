// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package core 核心功能
package core

import (
	"context"
	"database/sql"
)

const (
	defaultAINameSuffix = "_ai"
	defaultPKNameSuffix = "_pk"
)

// PKName 生成主键约束的名称
//
// 各个数据库对主键约束的规定并不统一，mysql 会忽略约束名，
// 为了统一，主键约束的名称统一由此函数生成，用户不能别外指定。
func PKName(table string) string {
	return table + defaultPKNameSuffix
}

// AIName 生成 AI 约束名称
//
// 自增约束的实现，各个数据库并不相同，诸如 mysql 直接加在列信息上，
// 而 postgres 会创建 sequence，需要指定 sequence 名称。
func AIName(table string) string {
	return table + defaultAINameSuffix
}

// Engine 数据库执行的基本接口。
//
// orm.DB 和 orm.Tx 应该实现此接口。
type Engine interface {
	Dialect() Dialect

	Query(query string, args ...interface{}) (*sql.Rows, error)

	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

	QueryRow(query string, args ...interface{}) *sql.Row

	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row

	Exec(query string, args ...interface{}) (sql.Result, error)

	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	Prepare(query string) (*Stmt, error)

	PrepareContext(ctx context.Context, query string) (*Stmt, error)
}

// Dialect 接口用于描述与数据库和驱动相关的一些语言特性。
//
// 除了 Dialect，同时还提供了部分 *Hooker 的接口，用于自定义某一条语句的实现。
// 一般情况下， 如果有多个数据是遵循 SQL 标准的，只有个别有例外，
// 那么该例外的 Dialect 实现，可以同时实现 Hooker 接口， 自定义该语句的实现。
type Dialect interface {
	// Dialect 的名称
	//
	// 自定义，不需要与 DriverName 相同。
	Name() string

	// 将列转换成数据支持的类型表达式
	SQLType(col *Column) (string, error)

	// 是否允许在事务中执行 DDL
	//
	// 比如在 postgresql 中，如果创建一个带索引的表，会采用在事务中，
	// 分多条语句创建表。
	// 而像 mysql 等不支持事务内 DDL 的数据库，则会采用普通的方式，
	// 依次提交语句。
	TransactionalDDL() bool

	// 根据当前的数据库，对 SQL 作调整。
	//
	// 比如替换 {} 符号；处理 sql.NamedArgs；
	// postgresql 需要将 ? 改成 $1 等形式。
	SQL(query string, args []interface{}) (string, []interface{}, error)

	// 查询服务器版本号的 SQL 语句。
	VersionSQL() string

	// 生成 `LIMIT N OFFSET M` 或是相同的语意的语句。
	//
	// offset 值为一个可选参数，若不指定，则表示 `LIMIT N` 语句。
	// 返回的是对应数据库的 limit 语句以及语句中占位符对应的值。
	//
	// limit 和 offset 可以是 SQL.NamedArg 类型。
	LimitSQL(limit interface{}, offset ...interface{}) (string, []interface{})

	// 自定义获取 LastInsertID 的获取方式。
	//
	// 类似于 postgresql 等都需要额外定义。
	//
	// 返回参数 SQL 表示额外的语句，如果为空，则执行的是标准的 SQL 插入语句；
	// append 表示在 SQL 不为空的情况下，SQL 与现有的插入语句的结合方式，
	// 如果为 true 表示直接添加在插入语句之后，否则为一条新的语句。
	LastInsertIDSQL(table, col string) (sql string, append bool)

	// 创建表时根据附加信息返回的部分 SQL 语句
	CreateTableOptionsSQL(sql *Builder, meta map[string][]string) error

	// 对预编译的内容进行处理。
	//
	// 目前大部分驱动都不支持 sql.NamedArgs，为了支持该功能，
	// 需要在预编译之前，对语句进行如下处理：
	// 1. 将 sql 中的 @xx 替换成 ?
	// 2. 将 sql 中的 @xx 在 sql 中的位置进行记录，并通过 orders 返回。
	Prepare(sql string) (query string, orders map[string]int)

	// 创建 AI 约束
	//CreateConstraintAI(name,col string)(string,error)
}
