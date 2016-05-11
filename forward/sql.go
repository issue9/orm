// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forward

import (
	"bytes"
	"database/sql"
	"errors"
)

const (
	flagWhere int8 = 1 << iota
	flagOrder
	flagColumn // SELECT 的标记
	flagSet    // UPDATE 的 SET 标记
	flagValues // INSERT 的 VALUES 标记
)

var ErrHasErrors = errors.New("语句中包含一个或多个错误")

// 一个简单的 SQL 语句接接工具。
//  sql := forward.NewSQL(e)
//  sql.Prepare().SQLString() // .Exec().Query()   ???
//
//  // DELETE
//  sql.Delete().From("#table1").Where("1=1").And("2=2").Or("3=3")
//
//  // UPDATE
//  sql.Update().Table("#table1").Set("{col}",1).Set("{col2}",2).Date(map[string]interface{}).Where(...)
//
//  // INSERT
//  sql.Insert().Table("table").Set(k,v).Data(map)
//
//  // SELECT
//  sql.Select(c1,c2),Select(c3,c4).From("t1 as t2").Where(...).Asc(...)
type SQL struct {
	engine Engine
	buffer *bytes.Buffer
	args   []interface{}
	flag   int8
	errors []error
}

func NewSQL(e Engine) *SQL {
	return &SQL{
		engine: e,
		buffer: new(bytes.Buffer),
		args:   make([]interface{}, 0, 10),
		flag:   0,
		errors: make([]error, 0, 5),
	}
}

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

// 是否在构建过程中触发错误信息
func (sql *SQL) HasError() bool {
	return len(sql.errors) > 0
}

// 返回所有的错误内容
func (sql *SQL) Errors() []error {
	return sql.errors
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

// 去掉尾部的 n 个字符
func (sql *SQL) TruncateLast(n int) *SQL {
	sql.buffer.Truncate(sql.buffer.Len() - n)
	return sql
}

func (sql *SQL) Delete(table string) *SQL {
	return sql.WriteString("DELETE FROM ").WriteString(table)
}

func (sql *SQL) Select(cols ...string) *SQL {
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

func (sql *SQL) Insert(table string) *SQL {
	return sql.WriteString("INSERT INTO ").WriteString(table)
}

func (sql *SQL) Update(table string) *SQL {
	return sql.WriteString("UPDATE ").WriteString(table)
}

// 拼接表名字符串。
func (sql *SQL) From(table string) *SQL {
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

func (sql *SQL) Limit(limit, offset int) *SQL {
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

// 指定插入的数据，需要与 Keys 中的名称一一对应。
func (sql *SQL) Values(vals ...interface{}) *SQL {
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

func (sql *SQL) Join(typ, table, on string) *SQL {
	sql.WriteString(typ)
	sql.WriteString(" JOIN ")
	sql.WriteString(table)
	sql.WriteByte(' ')
	sql.WriteString(on)

	return sql
}

func (sql *SQL) String() (string, []interface{}, error) {
	if sql.HasError() {
		return "", nil, ErrHasErrors
	}

	return sql.buffer.String(), sql.args, nil
}

func (sql *SQL) Prepare() (*sql.Stmt, []interface{}, error) {
	if sql.HasError() {
		return nil, nil, ErrHasErrors
	}

	stmt, err := sql.engine.Prepare(true, sql.buffer.String())
	if err != nil {
		return nil, nil, err
	}

	return stmt, sql.args, nil
}
