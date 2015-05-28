// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"bytes"
)

type Dialect interface {
	// 返回符合当前数据库规范的引号对。
	QuoteTuple() (openQuote, closeQuote byte)

	// 给一个关键字加引号
	Quote(w *bytes.Buffer, colName string) error

	// 生成LIMIT N OFFSET M 或是相同的语意的语句。
	// offset值为一个可选参数，若不指定，则表示LIMIT N语句。
	// 返回的是对应数据库的limit语句以及语句中占位符对应的值。
	LimitSQL(w *bytes.Buffer, limit int, offset ...int) ([]int, error)

	// 输出非AI列的定义，必须包含末尾的分号
	NoAIColSQL(w *bytes.Buffer, m *Model) error

	// 输出AI列的定义，必须包含末尾的分号
	AIColSQL(w *bytes.Buffer, m *Model) error

	// 输出所有的约束定义，必须包含末尾的分号
	ConstraintsSQL(w *bytes.Buffer, m *Model) error

	// 清空表内容，重置AI。
	TruncateTableSQL(tableName string) string

	// 是否支持一次性插入多条语句
	SupportInsertMany() bool
}
