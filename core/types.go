// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package core

import (
	"database/sql"
)

// 通用但又没有统一标准的数据库功能接口。
//
// 有可能一个Dialect实例会被多个其它实例引用，
// 不应该在Dialect实例中保存状态值等内容。
type Dialect interface {
	// 对字段或是表名的引用字符。
	QuoteStr() (left, right string)

	// 是否支持返回LastInsertId()特性。
	SupportLastInsertId() bool

	// 从dataSourceName变量中获取数据库的名称。
	GetDBName(dataSourceName string) string

	// 生成LIMIT N OFFSET M 或是相同的语意的语句。
	// offset值为一个可选参数，若不指定，则表示LIMIT N语句。
	// 返回的是对应数据库的limit语句以及语句中占位符对应的值。
	LimitSQL(limit int, offset ...int) (sql string, args []interface{})

	// 根据一个Model创建或是更新表。
	// 表的创建虽然语法上大致上相同，但细节部分却又不一样，
	// 干脆整个过程完全交给Dialect去完成。
	CreateTable(db DB, m *Model) error
}

// 操作数据库的接口，用于统一普通数据库操作和事务操作。
type DB interface {
	// 当前操作数据库的名称。
	Name() string

	// 获取Stmts实例。
	GetStmts() *Stmts

	// 返回Dialect接口。
	Dialect() Dialect

	// 相当于sql.DB.Exec()。
	// 但是会将语句中的表名前缀和字段引号占位符替换掉。
	Exec(sql string, args ...interface{}) (sql.Result, error)

	// 相当于sql.DB.Query()。
	// 但是会将语句中的表名前缀和字段引号占位符替换掉。
	Query(sql string, args ...interface{}) (*sql.Rows, error)

	// 相当于sql.DB.QueryRow()。
	// 但是会将语句中的表名前缀和字段引号占位符替换掉。
	QueryRow(sql string, args ...interface{}) *sql.Row

	// 相当于sql.DB.Prepare()。
	// 但是会将语句中的表名前缀和字段引号占位符替换掉。
	Prepare(sql string) (*sql.Stmt, error)
}
