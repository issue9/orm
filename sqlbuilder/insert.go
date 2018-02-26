// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import "errors"

// InsertStmt 表示插入操作的 SQL 语句
type InsertStmt struct {
	table string
	cols  []string
	args  [][]interface{}
}

// Insert 声明一条插入语句
func Insert(table string) *InsertStmt {
	return &InsertStmt{
		table: table,
		args:  make([][]interface{}, 0, 10),
	}
}

// Columns 指定插入的列，多次指定，之前的会被覆盖。
func (stmt *InsertStmt) Columns(cols ...string) *InsertStmt {
	stmt.cols = cols
	return stmt
}

// Values 指定需要插入的值
func (stmt *InsertStmt) Values(vals ...interface{}) *InsertStmt {
	stmt.args = append(stmt.args, vals)
	return stmt
}

// SQL 获取 SQL 的语句及参数部分
func (stmt *InsertStmt) SQL() (string, []interface{}, error) {
	if stmt.table == "" {
		return "", nil, errors.New("表名不能为空")
	}

	if len(stmt.cols) == 0 {
		return "", nil, errors.New("列名不能为空")
	}

	if len(stmt.args) == 0 {
		return "", nil, errors.New("没有指定数据")
	}

	for _, vals := range stmt.args {
		if len(vals) != len(stmt.cols) {
			return "", nil, errors.New("数所与列名数量不相同")
		}
	}

	buffer := newStringBuilder("INSERT INTO ")
	buffer.writeString(stmt.table)

	buffer.writeByte('(')
	for _, col := range stmt.cols {
		buffer.writeString(col)
		buffer.writeByte(',')
	}
	buffer.truncateLast(1)
	buffer.writeByte(')')

	buffer.writeString(" VALUES ")
	for index, vals := range stmt.args {
		// TODO
	}
}
