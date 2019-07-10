// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
)

// Rester 重置对象数据
//
// 实现该接口的对象，可以调用 Reset 方法重置对象的数据以达到重复利用数据的问题。
type Rester interface {
	Reset()
}

// SQLer 定义 SQL 语句的基本接口
type SQLer interface {
	SQL() (query string, args []interface{}, err error)
}

// DDLSQLer SQL 中 DDL 语句的基本接口
//
// 大部分数据的 DDL 操作是有多条语句组成，比如 CREATE TABLE
// 可能包含了额外的定义信息。
type DDLSQLer interface {
	DDLSQL() ([]string, error)
}

// WhereStmter 带 Where 语句的 SQL
type WhereStmter interface {
	WhereStmt() *WhereStmt
}

// Engine 数据库执行的基本接口。
//
// NOTE: 需要符合 SQL.DB 和 SQL.Tx 的定义。
type Engine interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)

	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)

	QueryRow(query string, args ...interface{}) *sql.Row

	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row

	Exec(query string, args ...interface{}) (sql.Result, error)

	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	Prepare(query string) (*sql.Stmt, error)

	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// Dialect 接口用于描述与数据库相关的一些语言特性。
//
// 除了 Dialect，同时还提供了部分 *Hooker 的接口，用于自定义某一条语句的实现。
// 一般情况下， 如果有多个数据是遵循 SQL 标准的，只有个别有例外，
// 那么该例外的 Dialect 实现，可以同时实现 Hooker 接口， 自定义该语句的实现。
type Dialect interface {
	// 返回符合当前数据库规范的引号对。
	QuoteTuple() (openQuote, closeQuote byte)

	// 将列转换成数据支持的类型
	SQLType(col *Column) (string, error)

	// 根据当前的数据库，对 SQL 作调整。
	//
	// 比如占位符 postgresql 可以使用 $1 等形式。
	SQL(sql string) (string, error)

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
	CreateTableOptionsSQL(sql *SQLBuilder, meta map[string][]string) error

	// 创建 AI 约束
	//CreateConstraintAI(name,col string)(string,error)
}
