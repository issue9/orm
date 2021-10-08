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
func PKName(table string) string { return table + defaultPKNameSuffix }

// AIName 生成 AI 约束名称
//
// 自增约束的实现，各个数据库并不相同，诸如 mysql 直接加在列信息上，
// 而 postgres 会创建 sequence，需要指定 sequence 名称。
//
// 参数 table 必须是完整的表名，如果有表名前缀，也需要带上。
func AIName(table string) string { return table + defaultAINameSuffix }

// Engine 数据库执行的基本接口
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

// Dialect 用于描述与数据库和驱动相关的一些特性
//
// Dialect 的实现者除了要实现 Dialect 之外，
// 还需要根据数据库的支持情况实现 sqlbuilder 下的部分 *Hooker 接口。
type Dialect interface {
	// DBName 当前关联的数据库名称
	//
	// 数据库名称和驱动名未必相同。比如 mysql 和 mariadb 可能采用相同的驱动名；
	DBName() string

	// DriverName 与当前实例关联的驱动名称
	//
	// 原则上驱动名和 Dialect 应该是一一对应的，但是也会有例外，比如：
	// github.com/lib/pq 和 github.com/jackc/pgx/v4/stdlib 功能上是相同的，
	// 仅注册的名称的不同。
	DriverName() string

	// TablePrefix 表名前缀
	//
	// 表名前缀由构建 Dialect 时传递给对象，可以由此方法返回。
	TablePrefix() string

	// SQLType 将列转换成数据支持的类型表达式
	//
	// 必须实现对所有 PrimitiveType 类型的转换。
	SQLType(*Column) (string, error)

	// TransactionalDDL 是否允许在事务中执行 DDL
	//
	// 比如在 postgresql 中，如果创建一个带索引的表，会采用在事务中，
	// 分多条语句创建表。
	// 而像 mysql 等不支持事务内 DDL 的数据库，则会采用普通的方式，
	// 依次提交语句。
	TransactionalDDL() bool

	// VersionSQL 查询服务器版本号的 SQL 语句
	VersionSQL() string

	// LimitSQL 生成 `LIMIT N OFFSET M` 或是相同的语意的语句片段
	//
	// offset 值为一个可选参数，若不指定，则表示 `LIMIT N` 语句。
	// 返回的是对应数据库的 limit 语句以及语句中占位符对应的值。
	//
	// limit 和 offset 可以是 SQL.NamedArg 类型。
	LimitSQL(limit interface{}, offset ...interface{}) (string, []interface{})

	// LastInsertIDSQL 自定义获取 LastInsertID 的获取方式
	//
	// 类似于 postgresql 等都需要额外定义。
	//
	// sql 表示额外的语句，如果为空，则执行的是标准的 SQL 插入语句；
	// append 表示在 sql 不为空的情况下，sql 与现有的插入语句的结合方式，
	// 如果为 true 表示直接添加在插入语句之后，否则为一条新的语句。
	LastInsertIDSQL(table, col string) (sql string, append bool)

	// CreateTableOptionsSQL 创建表时根据附加信息返回的部分 SQL 语句
	CreateTableOptionsSQL(sql *Builder, meta map[string][]string) error

	// TruncateTableSQL 生成清空数据表并重置自增列的语句
	//
	// ai 表示自增列的名称，可以为空，表示没有自去列。
	TruncateTableSQL(table, ai string) ([]string, error)

	// CreateViewSQL 生成创建视图的 SQL 语句
	CreateViewSQL(replace, temporary bool, name, selectQuery string, cols []string) ([]string, error)

	// DropIndexSQL 生成删除索引的语句
	//
	// table 为表名，部分数据库需要；
	// index 表示索引名；
	DropIndexSQL(table, index string) (string, error)

	// Fix 对 sql 语句作调整
	//
	// 比如替换 {} 符号；处理 sql.NamedArgs；
	// postgresql 需要将 ? 改成 $1 等形式。
	// 以及对 args 的参数作校正，比如 lib/pq 对 time.Time 处理有问题，也可以在此处作调整。
	//
	// NOTE: query 中不能同时存在 ? 和命名参数。因为如果是命名参数，则 args 的顺序可以是随意的。
	Fix(query string, args []interface{}) (string, []interface{}, error)

	// Prepare 对预编译的内容进行处理
	//
	// 目前大部分驱动都不支持 sql.NamedArgs，为了支持该功能，
	// 需要在预编译之前，对语句进行如下处理：
	//  1. 将 sql 中的 @xx 替换成 ?
	//  2. 将 sql 中的 @xx 在 sql 中的位置进行记录，并通过 orders 返回。
	// query 为处理后的 SQL 语句；
	// orders 为参数名在 query 中对应的位置，第一个位置为 0，依次增加。
	//
	// NOTE: query 中不能同时存在 ? 和命名参数。因为如果是命名参数，则 Exec 等的参数顺序可以是随意的。
	Prepare(sql string) (query string, orders map[string]int, err error)
}
