// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forward

import "database/sql"

// 将一组错误信息转换成一个标准的 error 接口。
type Errors []error

func (e Errors) Error() string {
	msg := "发生以下错误："
	for _, err := range e {
		msg += err.Error() + "\n"
	}

	return msg
}

// DB 与 Tx 的共有接口。
type Engine interface {
	// 获取与之关联的 Dialect 接口。
	Dialect() Dialect

	// 执行一条查询语句，并返回相应的 sql.Rows 实例。
	// 功能基本上等同于标准库 database/sql 的 DB.Query()
	// replace 指示是否替换掉语句中的占位符，语句中可以指定两种占位符：
	//  - # 表示一个表名前缀；
	//  - {} 表示一对 Quote 字符。
	//
	// 如：以下内容，在 replace 为 false 时，将原样输出，
	// 否则将被转换成以下字符串(以 mysql 为例，假设当前的 prefix 为 p_)
	//  select * from #user where {group}=1
	//  // 转换后
	//  select * from prefix_user where `group`=1
	Query(replace bool, query string, args ...interface{}) (*sql.Rows, error)

	// 功能等同于 database/sql 的 DB.Exec()。
	// replace 参数可参考 Engine.Query() 的说明。
	Exec(replace bool, query string, args ...interface{}) (sql.Result, error)

	// 功能等同于 database/sql 的 DB.Prepare()。
	// replace 参数可参考 Engine.Query() 的说明。
	Prepare(replace bool, query string) (*sql.Stmt, error)

	// 插入数据，若需一次性插入多条数据，请使用tx.Insert()。
	Insert(v interface{}) (sql.Result, error)

	// 删除符合条件的数据。
	// 查找条件以结构体定义的主键或是唯一约束(在没有主键的情况下)来查找，
	// 若两者都不存在，则将返回error
	Delete(v interface{}) (sql.Result, error)

	// 更新数据，零值不会被提交。
	// 查找条件以结构体定义的主键或是唯一约束(在没有主键的情况下)来查找，
	// 若两者都不存在，则将返回error
	Update(v interface{}, cols ...string) (sql.Result, error)

	// 查询一个符合条件的数据。
	// 查找条件以结构体定义的主键或是唯一约束(在没有主键的情况下)来查找，
	// 若两者都不存在，则将返回error
	// 若没有符合条件的数据，将不会对参数v做任何变动。
	Select(v interface{}) error

	// 查询符合 v 条件的记录数量。
	// v 中的所有非零字段都将参与查询。
	// 若需要复杂的查询方式，请构建 SQL 对象查询。
	Count(v interface{}) (int, error)

	// 创建一张表。
	Create(v interface{}) error

	// 删除一张表。
	Drop(v interface{}) error

	// 清空一张表。
	Truncate(v interface{}) error

	SQL() *SQL
}

// Dialect 接口用于描述与数据库相关的一些语言特性。
type Dialect interface {
	// 返回当前数据库的名称。
	Name() string

	// 返回符合当前数据库规范的引号对。
	QuoteTuple() (openQuote, closeQuote byte)

	// 替换语句中的?占位符
	ReplaceMarks(*string) error

	// 生成 `LIMIT N OFFSET M` 或是相同的语意的语句。
	// offset 值为一个可选参数，若不指定，则表示 `LIMIT N` 语句。
	// 返回的是对应数据库的 limit 语句以及语句中占位符对应的值。
	LimitSQL(sql *SQL, limit int, offset ...int) []interface{}

	// 输出非AI列的定义，必须包含末尾的分号
	NoAIColSQL(sql *SQL, m *Model) error

	// 输出AI列的定义，必须包含末尾的分号
	AIColSQL(sql *SQL, m *Model) error

	// 输出所有的约束定义，必须包含末尾的分号
	ConstraintsSQL(sql *SQL, m *Model)

	// 清空表内容，重置 AI。
	// aiColumn 需要被重置的自增列列名
	TruncateTableSQL(sql *SQL, tableName, aiColumn string)

	// 是否支持一次性插入多条语句
	SupportInsertMany() bool
}
