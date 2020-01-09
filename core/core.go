// SPDX-License-Identifier: MIT

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
// 为了统一，主键约束的名称统一由此函数生成，用户不能另外指定。
//
// 参数 table 必须是完整的表名，如果有表名前缀，也需要带上。
func PKName(table string) string {
	return table + defaultPKNameSuffix
}

// AIName 生成 AI 约束名称
//
// 自增约束的实现，各个数据库并不相同，诸如 mysql 直接加在列信息上，
// 而 postgres 会创建 sequence，需要指定 sequence 名称。
//
// 参数 table 必须是完整的表名，如果有表名前缀，也需要带上。
func AIName(table string) string {
	return table + defaultAINameSuffix
}

// Engine 数据库执行的基本接口。
//
// orm.DB 和 orm.Tx 应该实现此接口。
type Engine interface {
	TablePrefix() string

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

// Dialect 接口用于描述与数据库和驱动相关的一些特性。
//
// Dialect 的实现者除了要实现 Dialect 之外，
// 还需要根据数据库的支持情况实现 sqlbuilder 下的部分 *Hooker 接口。
type Dialect interface {
	// Dialect 的名称
	//
	// 可以直接采用驱动的 DriverName 值，方便简单。
	Name() string

	// 将列转换成数据支持的类型表达式
	SQLType(col *Column) (string, error)

	// 将 v 格式化为 SQL 对应的格式
	//
	// 比如在 mysql 中，true 会返回 1，而 postgres 中则返回 true
	//
	// 参数 length 表示部分需要固定长度的数据格式，比如浮点数，
	// 或是时间格式也需要精度。
	//
	// SQLFormat 应该优先调用 driver.Valuer 获取其原始类型的值，
	// 再作进一步的转换。
	SQLFormat(v interface{}, length ...int) (string, error)

	// 是否允许在事务中执行 DDL
	//
	// 比如在 postgresql 中，如果创建一个带索引的表，会采用在事务中，
	// 分多条语句创建表。
	// 而像 mysql 等不支持事务内 DDL 的数据库，则会采用普通的方式，
	// 依次提交语句。
	TransactionalDDL() bool

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

	// 根据当前的数据库，对 SQL 作调整。
	//
	// 比如替换 {} 符号；处理 sql.NamedArgs；
	// postgresql 需要将 ? 改成 $1 等形式。
	SQL(query string, args []interface{}) (string, []interface{}, error)

	// 对预编译的内容进行处理。
	//
	// 目前大部分驱动都不支持 sql.NamedArgs，为了支持该功能，
	// 需要在预编译之前，对语句进行如下处理：
	//  1. 将 sql 中的 @xx 替换成 ?
	//  2. 将 sql 中的 @xx 在 sql 中的位置进行记录，并通过 orders 返回。
	// query 为处理后的 SQL 语句；
	// orders 为参数名在 query 中对应的位置，第一个位置为 0，依次增加。
	Prepare(sql string) (query string, orders map[string]int, err error)
}
