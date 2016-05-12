// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forward

import "database/sql"

type Errors []error

func (e Errors) Error() string {
	msg := "发生以下错误："
	for _, err := range e {
		msg += err.Error() + "\n"
	}

	return msg
}

// DB与Tx的共有接口。
type Engine interface {
	Dialect() Dialect

	Query(replace bool, query string, args ...interface{}) (*sql.Rows, error)

	Exec(replace bool, query string, args ...interface{}) (sql.Result, error)

	Prepare(replace bool, query string) (*sql.Stmt, error)

	Prefix() string
}

// Dialect接口用于描述与数据库相关的一些语言特性。
type Dialect interface {
	// 返回当前数据库的名称。
	Name() string

	// 返回符合当前数据库规范的引号对。
	QuoteTuple() (openQuote, closeQuote byte)

	// 替换语句中的?占位符
	ReplaceMarks(*string) error

	// 生成LIMIT N OFFSET M 或是相同的语意的语句。
	// offset值为一个可选参数，若不指定，则表示LIMIT N语句。
	// 返回的是对应数据库的limit语句以及语句中占位符对应的值。
	LimitSQL(sql *SQL, limit int, offset ...int) []interface{}

	// 输出非AI列的定义，必须包含末尾的分号
	NoAIColSQL(sql *SQL, m *Model) error

	// 输出AI列的定义，必须包含末尾的分号
	AIColSQL(sql *SQL, m *Model) error

	// 输出所有的约束定义，必须包含末尾的分号
	ConstraintsSQL(sql *SQL, m *Model)

	// 清空表内容，重置AI。
	// aiColumn 需要被重置的自增列列名
	TruncateTableSQL(sql *SQL, tableName, aiColumn string)

	// 是否支持一次性插入多条语句
	SupportInsertMany() bool
}
