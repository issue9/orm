// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// 一个简单的 SQL 拼接工具。
package forward

import (
	"bytes"
	"database/sql"
	"fmt"

	"github.com/issue9/orm/fetch"
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

// SQL 一个简单的 SQL 语句接接工具。
//
// NOTE: SQL 的所有函数调用，将直接拼接到字符串，
// 而不会做缓存，所以调用顺序必须与 SQL 语法相同。
//
// DELETE
//  sql := New(engine).
//      Delete("table1").
//      Where("id>?", 5).
//      And("type=?", 2)
//  query, vals, err := sql.String()
type SQL struct {
	engine Engine
	buffer *bytes.Buffer
	args   []interface{}
	flag   int8
	errors []error
}

// NewSQL 声明一个 SQL 实例
func NewSQL(e Engine) *SQL {
	return &SQL{
		engine: e,
		buffer: new(bytes.Buffer),
		args:   make([]interface{}, 0, 10),
		flag:   0,
		errors: make([]error, 0, 5),
	}
}

// Reset 重置所有的数据为初始值，这样可以重复利用该 SQL 对象。
func (sql *SQL) Reset() *SQL {
	sql.buffer.Reset()
	sql.args = sql.args[:0]
	sql.flag = 0
	sql.errors = sql.errors[:0]

	return sql
}

func (sql *SQL) isSetFlag(flag int8) bool {
	return sql.flag&flag > 0
}

func (sql *SQL) setFlag(flag int8) {
	sql.flag |= flag
}

// HasError 是否在构建过程中触发错误信息。当出现此错误时，说明在构建 SQL
// 语句的过程中出现了错误，需要调用 Errors() 获取详细的错误信息。
//
// NOTE: 在构建完 SQL 语句，准备执行数据库操作之前，
// 都应该调用此函数确认是否存在错误。
func (sql *SQL) HasError() bool {
	return len(sql.errors) > 0
}

// Buffer 获取与之关联的 bytes.Buffer 对像。
func (sql *SQL) Buffer() *bytes.Buffer {
	return sql.buffer
}

// Errors 返回所有的错误内容。
func (sql *SQL) Errors() error {
	return Errors(sql.errors)
}

func (sql *SQL) WriteByte(c byte) *SQL {
	err := sql.buffer.WriteByte(c)
	if err != nil {
		sql.errors = append(sql.errors, err)
	}

	return sql
}

func (sql *SQL) WriteString(s string) *SQL {
	_, err := sql.buffer.WriteString(s)
	if err != nil {
		sql.errors = append(sql.errors, err)
	}

	return sql
}

// TruncateLast 去掉尾部的 n 个字符。
func (sql *SQL) TruncateLast(n int) *SQL {
	sql.buffer.Truncate(sql.buffer.Len() - n)
	return sql
}

// Delete 启动一个 DELETE 语句。
func (sql *SQL) Delete(table string) *SQL {
	return sql.WriteString("DELETE FROM ").WriteString(table)
}

// Select 启动一个 SELECT 语句，并指定列名。可多次调用。
func (sql *SQL) Select(cols ...string) *SQL {
	if !sql.isSetFlag(flagColumn) {
		sql.WriteString("SELECT ")
		sql.setFlag(flagColumn)
	} else {
		sql.WriteByte(',')
	}

	for _, col := range cols {
		sql.WriteString(col)
		sql.WriteByte(',')
	}
	return sql.TruncateLast(1)
}

// Insert 启动一个 INSERT 语句。
func (sql *SQL) Insert(table string) *SQL {
	return sql.WriteString("INSERT INTO ").WriteString(table)
}

// Update 启动一个 UPDATE 语句。
func (sql *SQL) Update(table string) *SQL {
	return sql.WriteString("UPDATE ").WriteString(table)
}

// 拼接表名字符串。当调用 Select() 之后，此方法用于指定表名。
func (sql *SQL) From(table string) *SQL {
	sql.WriteString(" FROM ")
	sql.WriteString(table)
	return sql
}

// 构建 WHERE 语句，op 只能是 AND 或是 OR
func (sql *SQL) where(op string, cond string, args ...interface{}) *SQL {
	if !sql.isSetFlag(flagWhere) {
		sql.setFlag(flagWhere)
		op = " WHERE "
	}

	sql.WriteString(op)
	sql.WriteString(cond)
	sql.args = append(sql.args, args...)
	return sql
}

// And 的别名。
func (sql *SQL) Where(cond string, args ...interface{}) *SQL {
	return sql.And(cond, args...)
}

func (sql *SQL) And(cond string, args ...interface{}) *SQL {
	return sql.where(" AND ", cond, args...)
}

func (sql *SQL) Or(cond string, args ...interface{}) *SQL {
	return sql.where(" OR ", cond, args...)
}

func (sql *SQL) orderBy(order, col string) *SQL {
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

func (sql *SQL) Desc(col string) *SQL {
	return sql.orderBy(" DESC ", col)
}

func (sql *SQL) Asc(col string) *SQL {
	return sql.orderBy(" ASC ", col)
}

// offset 的值为多个时，只有第一个启作用
func (sql *SQL) Limit(limit int, offset ...int) *SQL {
	vals := sql.engine.Dialect().LimitSQL(sql, limit, offset...)
	sql.args = append(sql.args, vals...)
	return sql
}

// Keys 指定插入数据时的列名
func (sql *SQL) Keys(keys ...string) *SQL {
	sql.WriteByte('(')
	for _, key := range keys {
		sql.WriteString(key)
		sql.WriteByte(',')
	}
	sql.TruncateLast(1)
	sql.WriteByte(')')

	return sql
}

// Group 生成 GROUP BY 语句
func (sql *SQL) Group(col string) *SQL {
	sql.WriteString(" GROUP BY ")
	sql.WriteString(col)
	sql.WriteByte(' ')
	return sql
}

// Values 指定插入的数据，需要与 Keys 中的名称一一对应。
//
// NOTE: 若数据库支持多行插入，可多次调用，每次指定一行数据。
func (sql *SQL) Values(vals ...interface{}) *SQL {
	if !sql.isSetFlag(flagValues) {
		sql.WriteString("VALUES(")
		sql.setFlag(flagValues)
	} else {
		d := sql.engine.Dialect()
		if !d.SupportInsertMany() {
			sql.errors = append(sql.errors, fmt.Errorf("当前数据库[%v]不支持多行插入", d.Name()))
		}
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

// Set 指定需要更新的数据。
//
// 仅针对 UPDATE 语句，INSERT 请使用 Keys() 和 Values() 两个函数指定。
func (sql *SQL) Set(k string, v interface{}) *SQL {
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

// Incr 增加计数
func (sql *SQL) Incr(k string, v interface{}) *SQL {
	if !sql.isSetFlag(flagSet) {
		sql.WriteString(" SET ")
		sql.setFlag(flagSet)
	} else {
		sql.WriteByte(',')
	}

	sql.WriteString(k)
	sql.WriteString("=k+")
	sql.WriteString(fmt.Sprint(v))

	return sql
}

//  Decr 减少计数
func (sql *SQL) Decr(k string, v interface{}) *SQL {
	if !sql.isSetFlag(flagSet) {
		sql.WriteString(" SET ")
		sql.setFlag(flagSet)
	} else {
		sql.WriteByte(',')
	}

	sql.WriteString(k)
	sql.WriteString("=k-")
	sql.WriteString(fmt.Sprint(v))

	return sql
}

// Join 拼接 SELECT 语句的 JOIN 部分。
func (sql *SQL) Join(typ, table, on string) *SQL {
	sql.WriteByte(' ')
	sql.WriteString(typ)
	sql.WriteString(" JOIN ")
	sql.WriteString(table)
	sql.WriteString(" ON ")
	sql.WriteString(on)

	return sql
}

// String 返回 SQL 语句和其对应的值。
func (sql *SQL) String() (string, []interface{}, error) {
	if sql.HasError() {
		return "", nil, sql.Errors()
	}

	return sql.buffer.String(), sql.args, nil
}

// Prepare 返回预编译的实例及对应的值。
func (sql *SQL) Prepare(replace bool) (*sql.Stmt, []interface{}, error) {
	if sql.HasError() {
		return nil, nil, sql.Errors()
	}

	stmt, err := sql.engine.Prepare(replace, sql.buffer.String())
	if err != nil {
		return nil, nil, err
	}

	return stmt, sql.args, nil
}

// Exec 执行 SQL 语句。
func (sql *SQL) Exec(replace bool) (sql.Result, error) {
	query, vals, err := sql.String()
	if err != nil {
		return nil, err
	}

	return sql.engine.Exec(replace, query, vals...)
}

// Query 执行 SQL 查询语句。仅对 SELECT 启作用。
func (sql *SQL) Query(replace bool) (*sql.Rows, error) {
	query, vals, err := sql.String()
	if err != nil {
		return nil, err
	}

	return sql.engine.Query(replace, query, vals...)
}

// QueryMap 返回符合当前条件的所有记录。
//
// replace，是否需将对语句的占位符进行替换。
func (sql *SQL) QueryMap(replace bool) ([]map[string]interface{}, error) {
	rows, err := sql.Query(replace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return fetch.Map(false, rows)
}

// QueryMapString 返回符合当前条件的所有记录，功能与 QueryMap 相同，但 map 中的值全为字符串类型。
//
// replace，是否需将对语句的占位符进行替换。
func (sql *SQL) QueryMapString(replace bool) ([]map[string]string, error) {
	rows, err := sql.Query(replace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return fetch.MapString(false, rows)
}

// QueryObj 将符合当前条件的所有记录依次写入 objs 中。
//
// replace，是否需将对语句的占位符进行替换。
func (sql *SQL) QueryObj(replace bool, objs interface{}) (int, error) {
	rows, err := sql.Query(replace)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return fetch.Obj(objs, rows)
}

// QueryColumn 返回符合条件的记录中的某一列值。
func (sql *SQL) QueryColumn(replace bool, col string) ([]interface{}, error) {
	rows, err := sql.Query(replace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return fetch.Column(false, col, rows)
}

// QueryColumnString 返回符合条件的记录中的某一列值，
// 功能与 QueryColumn 相同，但返回的值均为字符串类型。
func (sql *SQL) QueryColumnString(replace bool, col string) ([]string, error) {
	rows, err := sql.Query(replace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return fetch.ColumnString(false, col, rows)
}
