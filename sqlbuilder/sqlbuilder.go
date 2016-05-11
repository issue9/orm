// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bytes"
	"database/sql"
	"errors"

	"github.com/issue9/orm/forward"
)

// 一组标记位，用于标记某些可重复调用的函数，是否是第一次调用。
// 比如 OrderBy 在第一次调用时，需要填 `ORDER BY`字符串，之后
// 的只要跟着列名和相应的排序方式即可。
const (
	flagWhere  int8 = 1 << iota
	flagOrder       // ORDER BY
	flagColumn      // SELECT
	flagSet         // UPDATE 的 SET
	flagValues      // INSERT 的 VALUES
)

var ErrHasErrors = errors.New("语句中包含一个或多个错误")

// SQLBuilder 一个简单的 SQL 语句接接工具。
// NOTE: 调用顺序必须与 SQL 语句相同。
//
// DELETE
//  sql := New(engine).
//      Delete("table1").
//      Where("id>?", 5).
//      And("type=?", 2)
//  query, vals, err := sql.String()
type SQLBuilder struct {
	engine forward.Engine
	buffer *bytes.Buffer
	args   []interface{}
	flag   int8
	errors []error
}

// 声明一个 SQLBuilder 实例
func New(e forward.Engine) *SQLBuilder {
	return &SQLBuilder{
		engine: e,
		buffer: new(bytes.Buffer),
		args:   make([]interface{}, 0, 10),
		flag:   0,
		errors: make([]error, 0, 5),
	}
}

// 重置所有的数据为初始值，这样可以重复利用该 SQLBuilder 对象。
func (sql *SQLBuilder) Reset() *SQLBuilder {
	sql.buffer.Reset()
	sql.args = sql.args[:0]
	sql.flag = 0
	sql.errors = sql.errors[:0]

	return sql
}

func (sql *SQLBuilder) isSetFlag(flag int8) bool {
	return sql.flag&flag > 0
}

func (sql *SQLBuilder) setFlag(flag int8) {
	sql.flag |= flag
}

// 是否在构建过程中触发错误信息
func (sql *SQLBuilder) HasError() bool {
	return len(sql.errors) > 0
}

// 返回所有的错误内容
func (sql *SQLBuilder) Errors() []error {
	return sql.errors
}

func (sql *SQLBuilder) WriteByte(c byte) *SQLBuilder {
	err := sql.buffer.WriteByte(c)
	if err != nil {
		sql.errors = append(sql.errors, err)
	}

	return sql
}

func (sql *SQLBuilder) WriteString(s string) *SQLBuilder {
	_, err := sql.buffer.WriteString(s)
	if err != nil {
		sql.errors = append(sql.errors, err)
	}

	return sql
}

// 去掉尾部的 n 个字符
func (sql *SQLBuilder) TruncateLast(n int) *SQLBuilder {
	sql.buffer.Truncate(sql.buffer.Len() - n)
	return sql
}

// 启动一个 DELETE 语名。
func (sql *SQLBuilder) Delete(table string) *SQLBuilder {
	return sql.WriteString("DELETE FROM ").WriteString(table)
}

func (sql *SQLBuilder) Select(cols ...string) *SQLBuilder {
	if !sql.isSetFlag(flagColumn) {
		sql.WriteString("SELECT ")
		sql.setFlag(flagColumn)
	}

	for _, col := range cols {
		sql.WriteString(col)
		sql.WriteByte(',')
	}
	return sql.TruncateLast(1)
}

func (sql *SQLBuilder) Insert(table string) *SQLBuilder {
	return sql.WriteString("INSERT INTO ").WriteString(table)
}

func (sql *SQLBuilder) Update(table string) *SQLBuilder {
	return sql.WriteString("UPDATE ").WriteString(table)
}

// 拼接表名字符串。
func (sql *SQLBuilder) From(table string) *SQLBuilder {
	sql.WriteString(table)
	return sql
}

// 构建 WHERE 语句，op 只能是 AND 或是 OR
func (sql *SQLBuilder) where(op string, cond string, args ...interface{}) *SQLBuilder {
	if !sql.isSetFlag(flagWhere) {
		sql.setFlag(flagWhere)
		op = " WHERE "
	}

	sql.WriteString(op)
	sql.WriteString(cond)
	sql.args = append(sql.args, args...)

	return sql
}

func (sql *SQLBuilder) Where(cond string, args ...interface{}) *SQLBuilder {
	return sql.And(cond, args...)
}

func (sql *SQLBuilder) And(cond string, args ...interface{}) *SQLBuilder {
	return sql.where(" AND ", cond, args...)
}

func (sql *SQLBuilder) Or(cond string, args ...interface{}) *SQLBuilder {
	return sql.where(" OR ", cond, args...)
}

func (sql *SQLBuilder) orderBy(order, col string) *SQLBuilder {
	if !sql.isSetFlag(flagOrder) {
		sql.setFlag(flagOrder)
		sql.WriteString(" ORDER BY ")
	} else {
		sql.WriteByte(',')
	}

	sql.WriteString(col)
	sql.WriteString(order)

	return sql
}

func (sql *SQLBuilder) Desc(col string) *SQLBuilder {
	return sql.orderBy(" DESC ", col)
}

func (sql *SQLBuilder) Asc(col string) *SQLBuilder {
	return sql.orderBy(" ASC ", col)
}

func (sql *SQLBuilder) Limit(limit, offset int) *SQLBuilder {
	vals, err := sql.engine.Dialect().LimitSQL(sql.buffer, limit, offset)
	if err != nil {
		sql.errors = append(sql.errors, err)
	}

	args := make([]interface{}, 0, 2)
	for _, val := range vals {
		args = append(args, val)
	}

	sql.args = append(sql.args, args...)

	return sql
}

// 指定插入数据时的列名
func (sql *SQLBuilder) Keys(keys ...string) *SQLBuilder {
	sql.WriteByte('(')
	for _, key := range keys {
		sql.WriteString(key)
		sql.WriteByte(',')
	}
	sql.TruncateLast(1)
	sql.WriteByte(')')

	return sql
}

// 指定插入的数据，需要与 Keys 中的名称一一对应。
func (sql *SQLBuilder) Values(vals ...interface{}) *SQLBuilder {
	if !sql.isSetFlag(flagValues) {
		sql.WriteString("VALUES(")
		sql.setFlag(flagValues)
	} else {
		sql.WriteString(",(")
	}

	for _, v := range vals {
		sql.WriteString("?,")
		sql.args = append(sql.args, v)
	}
	sql.TruncateLast(1)

	sql.WriteByte(')')

	return sql
}

// 指定需要更新的数据
func (sql *SQLBuilder) Set(k string, v interface{}) *SQLBuilder {
	if !sql.isSetFlag(flagSet) {
		sql.WriteString(" SET ")
		sql.setFlag(flagSet)
	} else {
		sql.WriteByte(',')
	}

	sql.WriteString(k)
	sql.WriteString("=?")

	sql.args = append(sql.args, v)

	return sql
}

func (sql *SQLBuilder) Join(typ, table, on string) *SQLBuilder {
	sql.WriteString(typ)
	sql.WriteString(" JOIN ")
	sql.WriteString(table)
	sql.WriteByte(' ')
	sql.WriteString(on)

	return sql
}

func (sql *SQLBuilder) String() (string, []interface{}, error) {
	if sql.HasError() {
		return "", nil, ErrHasErrors
	}

	return sql.buffer.String(), sql.args, nil
}

func (sql *SQLBuilder) Prepare() (*sql.Stmt, []interface{}, error) {
	if sql.HasError() {
		return nil, nil, ErrHasErrors
	}

	stmt, err := sql.engine.Prepare(true, sql.buffer.String())
	if err != nil {
		return nil, nil, err
	}

	return stmt, sql.args, nil
}
