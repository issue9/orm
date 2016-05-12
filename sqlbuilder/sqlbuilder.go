// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// 一个简单的 SQL 拼接工具。
package sqlbuilder

import (
	"bytes"
	"database/sql"
	"fmt"

	"github.com/issue9/orm/fetch"
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

// SQLBuilder 一个简单的 SQL 语句接接工具。
// NOTE: SQLBuilder 的所有函数调用，将直接拼接到字符串，
// 而不会做缓存，所以调用顺序必须与 SQL 语法相同。
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
func (sb *SQLBuilder) Reset() *SQLBuilder {
	sb.buffer.Reset()
	sb.args = sb.args[:0]
	sb.flag = 0
	sb.errors = sb.errors[:0]

	return sb
}

func (sb *SQLBuilder) isSetFlag(flag int8) bool {
	return sb.flag&flag > 0
}

func (sb *SQLBuilder) setFlag(flag int8) {
	sb.flag |= flag
}

// 是否在构建过程中触发错误信息。当出现此错误时，说明在构建 SQL
// 语句的过程中出现了错误，需要调用 Errors() 获取详细的错误信息。
//
// NOTE: 在构建完 SQL 语句，准备执行数据库操作之前，
// 都应该调用此函数确认是否存在错误。
func (sb *SQLBuilder) HasError() bool {
	return len(sb.errors) > 0
}

// 返回所有的错误内容。
func (sb *SQLBuilder) Errors() error {
	return Errors(sb.errors)
}

func (sb *SQLBuilder) WriteByte(c byte) *SQLBuilder {
	err := sb.buffer.WriteByte(c)
	if err != nil {
		sb.errors = append(sb.errors, err)
	}

	return sb
}

func (sb *SQLBuilder) WriteString(s string) *SQLBuilder {
	_, err := sb.buffer.WriteString(s)
	if err != nil {
		sb.errors = append(sb.errors, err)
	}

	return sb
}

// 去掉尾部的 n 个字符。
func (sb *SQLBuilder) TruncateLast(n int) *SQLBuilder {
	sb.buffer.Truncate(sb.buffer.Len() - n)
	return sb
}

// 启动一个 DELETE 语句。
func (sb *SQLBuilder) Delete(table string) *SQLBuilder {
	return sb.WriteString("DELETE FROM ").WriteString(table)
}

// 启动一个 SELECT 语句，并指定列名。可多次调用。
func (sb *SQLBuilder) Select(cols ...string) *SQLBuilder {
	if !sb.isSetFlag(flagColumn) {
		sb.WriteString("SELECT ")
		sb.setFlag(flagColumn)
	} else {
		sb.WriteByte(',')
	}

	for _, col := range cols {
		sb.WriteString(col)
		sb.WriteByte(',')
	}
	return sb.TruncateLast(1)
}

// 启动一个 INSERT 语句。
func (sb *SQLBuilder) Insert(table string) *SQLBuilder {
	return sb.WriteString("INSERT INTO ").WriteString(table)
}

// 启动一个 UPDATE 语句。
func (sb *SQLBuilder) Update(table string) *SQLBuilder {
	return sb.WriteString("UPDATE ").WriteString(table)
}

// 拼接表名字符串。当调用 Select() 之后，此方法用于指定表名。
func (sb *SQLBuilder) From(table string) *SQLBuilder {
	sb.WriteString(" FROM ")
	sb.WriteString(table)
	return sb
}

// 构建 WHERE 语句，op 只能是 AND 或是 OR
func (sb *SQLBuilder) where(op string, cond string, args ...interface{}) *SQLBuilder {
	if !sb.isSetFlag(flagWhere) {
		sb.setFlag(flagWhere)
		op = " WHERE "
	}

	sb.WriteString(op)
	sb.WriteString(cond)
	sb.args = append(sb.args, args...)

	return sb
}

// And 的别名。
func (sb *SQLBuilder) Where(cond string, args ...interface{}) *SQLBuilder {
	return sb.And(cond, args...)
}

func (sb *SQLBuilder) And(cond string, args ...interface{}) *SQLBuilder {
	return sb.where(" AND ", cond, args...)
}

func (sb *SQLBuilder) Or(cond string, args ...interface{}) *SQLBuilder {
	return sb.where(" OR ", cond, args...)
}

func (sb *SQLBuilder) orderBy(order, col string) *SQLBuilder {
	if !sb.isSetFlag(flagOrder) {
		sb.setFlag(flagOrder)
		sb.WriteString(" ORDER BY ")
	} else {
		sb.WriteByte(',')
	}

	sb.WriteString(col)
	sb.WriteString(order)

	return sb
}

func (sb *SQLBuilder) Desc(col string) *SQLBuilder {
	return sb.orderBy(" DESC ", col)
}

func (sb *SQLBuilder) Asc(col string) *SQLBuilder {
	return sb.orderBy(" ASC ", col)
}

func (sb *SQLBuilder) Limit(limit, offset int) *SQLBuilder {
	vals, err := sb.engine.Dialect().LimitSQL(sb.buffer, limit, offset)
	if err != nil {
		sb.errors = append(sb.errors, err)
	}

	args := make([]interface{}, 0, 2)
	for _, val := range vals {
		args = append(args, val)
	}

	sb.args = append(sb.args, args...)

	return sb
}

// 指定插入数据时的列名
func (sb *SQLBuilder) Keys(keys ...string) *SQLBuilder {
	sb.WriteByte('(')
	for _, key := range keys {
		sb.WriteString(key)
		sb.WriteByte(',')
	}
	sb.TruncateLast(1)
	sb.WriteByte(')')

	return sb
}

// 指定插入的数据，需要与 Keys 中的名称一一对应。
//
// NOTE: 若数据库支持多行插入，可多次调用，每次指定一行数据。
func (sb *SQLBuilder) Values(vals ...interface{}) *SQLBuilder {
	if !sb.isSetFlag(flagValues) {
		sb.WriteString("VALUES(")
		sb.setFlag(flagValues)
	} else {
		d := sb.engine.Dialect()
		if !d.SupportInsertMany() {
			sb.errors = append(sb.errors, fmt.Errorf("当前数据库[%v]不支持多行插入", d.Name()))
		}
		sb.WriteString(",(")
	}

	for _, v := range vals {
		sb.WriteString("?,")
		sb.args = append(sb.args, v)
	}
	sb.TruncateLast(1)

	sb.WriteByte(')')

	return sb
}

// 指定需要更新的数据。
// 仅针对 UPDATE 语句，INSERT 请使用 Keys() 和 Values() 两个函数指定。
func (sb *SQLBuilder) Set(k string, v interface{}) *SQLBuilder {
	if !sb.isSetFlag(flagSet) {
		sb.WriteString(" SET ")
		sb.setFlag(flagSet)
	} else {
		sb.WriteByte(',')
	}

	sb.WriteString(k)
	sb.WriteString("=?")

	sb.args = append(sb.args, v)

	return sb
}

// 拼接 SELECT 语句的 JOIN 部分。
func (sb *SQLBuilder) Join(typ, table, on string) *SQLBuilder {
	sb.WriteByte(' ')
	sb.WriteString(typ)
	sb.WriteString(" JOIN ")
	sb.WriteString(table)
	sb.WriteString(" ON ")
	sb.WriteString(on)

	return sb
}

// 返回 SQL 语句和其对应的值。
func (sb *SQLBuilder) String() (string, []interface{}, error) {
	if sb.HasError() {
		return "", nil, sb.Errors()
	}

	return sb.buffer.String(), sb.args, nil
}

// 返回预编译的实例及对应的值。
func (sb *SQLBuilder) Prepare() (*sql.Stmt, []interface{}, error) {
	if sb.HasError() {
		return nil, nil, sb.Errors()
	}

	stmt, err := sb.engine.Prepare(true, sb.buffer.String())
	if err != nil {
		return nil, nil, err
	}

	return stmt, sb.args, nil
}

// 执行 SQL 语句。
func (sb *SQLBuilder) Exec(replace bool) (sql.Result, error) {
	query, vals, err := sb.String()
	if err != nil {
		return nil, err
	}

	return sb.engine.Exec(replace, query, vals...)
}

// 执行 SQL 查询语句。仅对 SELECT 启作用。
func (sb *SQLBuilder) Query(replace bool) (*sql.Rows, error) {
	query, vals, err := sb.String()
	if err != nil {
		return nil, err
	}

	return sb.engine.Query(replace, query, vals...)
}

// 返回符合当前条件的所有记录。
//
// replace，是否需将对语句的占位符进行替换。
func (sb *SQLBuilder) QueryMap(replace bool) ([]map[string]interface{}, error) {
	rows, err := sb.Query(replace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return fetch.Map(false, rows)
}

// 返回符合当前条件的所有记录，功能与 QueryMap 相同，但 map 中的值全为字符串类型。
//
// replace，是否需将对语句的占位符进行替换。
func (sb *SQLBuilder) QueryMapString(replace bool) ([]map[string]string, error) {
	rows, err := sb.Query(replace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return fetch.MapString(false, rows)
}

// 将符合当前条件的所有记录依次写入 objs 中。
//
// replace，是否需将对语句的占位符进行替换。
func (sb *SQLBuilder) QueryObj(replace bool, objs interface{}) (int, error) {
	rows, err := sb.Query(replace)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	return fetch.Obj(objs, rows)
}

// 返回符合条件的记录中的某一列值。
func (sb *SQLBuilder) QueryColumn(replace bool, col string) ([]interface{}, error) {
	rows, err := sb.Query(replace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return fetch.Column(false, col, rows)
}

// 返回符合条件的记录中的某一列值，功能与 QueryColumn 相同，但返回的值均为字符串类型。
func (sb *SQLBuilder) QueryColumnString(replace bool, col string) ([]string, error) {
	rows, err := sb.Query(replace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return fetch.ColumnString(false, col, rows)
}
