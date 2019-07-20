// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
	"reflect"
	"time"
)

// CreateTableStmt.Column 用到的数据类型。
var (
	BoolType    = reflect.TypeOf(true)
	IntType     = reflect.TypeOf(int(1))
	Int8Type    = reflect.TypeOf(int8(1))
	Int16Type   = reflect.TypeOf(int16(1))
	Int32Type   = reflect.TypeOf(int32(1))
	Int64Type   = reflect.TypeOf(int64(1))
	UintType    = reflect.TypeOf(uint(1))
	Uint8Type   = reflect.TypeOf(uint8(1))
	Uint16Type  = reflect.TypeOf(uint16(1))
	Uint32Type  = reflect.TypeOf(uint32(1))
	Uint64Type  = reflect.TypeOf(uint64(1))
	Float32Type = reflect.TypeOf(float32(1))
	Float64Type = reflect.TypeOf(float64(1))
	StringType  = reflect.TypeOf("")

	NullStringType  = reflect.TypeOf(sql.NullString{})
	NullInt64Type   = reflect.TypeOf(sql.NullInt64{})
	NullBoolType    = reflect.TypeOf(sql.NullBool{})
	NullFloat64Type = reflect.TypeOf(sql.NullFloat64{})
	RawBytesType    = reflect.TypeOf(sql.RawBytes{})
	TimeType        = reflect.TypeOf(time.Time{})

	//UintptrType=reflect.TypeOf(uintptr(1))
	//Complex64Type=reflect.TypeOf(complex64(1,1))
	//Complex128Type=reflect.TypeOf(complex128(1,1))
)

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
// 是 sql.DB 与 sql.Tx 的共有接口。
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
	CreateTableOptionsSQL(sql *SQLBuilder, meta map[string][]string) error

	// 创建 AI 约束
	//CreateConstraintAI(name,col string)(string,error)
}
