// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"bytes"
	"database/sql"
	"errors"

	"github.com/issue9/orm/fetch"
)

const (
	and = iota
	or
)

// 用于产生条件语句。
type Where struct {
	e     engine
	table string
	cond  *bytes.Buffer
	args  []interface{}
}

func newWhere(engine engine) *Where {
	return &Where{
		e:    engine,
		cond: new(bytes.Buffer),
		args: []interface{}{},
	}
}

func (w *Where) where(op int, cond string, args ...interface{}) *Where {
	switch {
	case w.cond.Len() == 0:
		w.cond.WriteString(" WHERE(")
	case op == and:
		w.cond.WriteString(" AND(")
	case op == or:
		w.cond.WriteString(" OR(")
	default:
		panic("where:错误的参数op")
	}
	w.cond.WriteString(cond)
	w.cond.WriteByte(')')
	w.args = append(w.args, args...)

	return w
}

// 将之后的语句以and的形式与当前的语句进行连接
func (w *Where) And(cond string, args ...interface{}) *Where {
	return w.where(and, cond, args...)
}

// 将之后的语句以or的形式与当前的语句进行连接
func (w *Where) Or(cond string, args ...interface{}) *Where {
	return w.where(or, cond, args...)
}

// 指定表名。
func (w *Where) Table(tableName string) *Where {
	w.table = tableName
	return w
}

// 将符合当前条件的所有记录删除。
// replace，是否需将对语句的占位符进行替换。
func (w *Where) Delete(replace bool) error {
	if len(w.table) == 0 {
		return errors.New("Where.Delete:未指定表名")
	}

	sql := pool.Get().(*bytes.Buffer)
	defer pool.Put(sql)

	sql.Reset()
	sql.WriteString("DELETE FROM ")
	w.e.Dialect().Quote(sql, w.table)
	sql.WriteString(w.cond.String())
	_, err := w.e.Exec(replace, sql.String(), w.args...)
	return err
}

// 更新符合当前条件的所有记录。
// replace，是否需将对语句的占位符进行替换。
func (w *Where) Update(replace bool, data map[string]interface{}) error {
	if len(w.table) == 0 {
		return errors.New("Where.Update:未指定表名")
	}

	if len(data) == 0 {
		return errors.New("Where.Update:未指定需要更新的数据")
	}

	sql := pool.Get().(*bytes.Buffer)
	defer pool.Put(sql)

	sql.Reset()
	sql.WriteString("UPDATE ")
	w.e.Dialect().Quote(sql, w.table)
	vals := make([]interface{}, 0, len(data)+len(w.args))
	sql.WriteString(" SET ")
	for k, v := range data {
		w.e.Dialect().Quote(sql, k)
		sql.WriteString("=?,")
		vals = append(vals, v)
	}
	sql.Truncate(sql.Len() - 1) // 去掉最后一个逗号

	sql.WriteString(w.cond.String())
	_, err := w.e.Exec(replace, sql.String(), append(vals, w.args...)...)
	return err
}

// 将符合当前条件的所有记录依次写入objs中。
// replace，是否需将对语句的占位符进行替换。
func (w *Where) Select(replace bool, objs interface{}) error {
	rows, err := w.buildSelectSQL(replace)
	if err != nil {
		return err
	}
	return fetch.Obj(objs, rows)
}

// 返回符合当前条件的所有记录。
// replace，是否需将对语句的占位符进行替换。
func (w *Where) SelectMap(replace bool, cols ...string) ([]map[string]interface{}, error) {
	rows, err := w.buildSelectSQL(replace, cols...)
	if err != nil {
		return nil, err
	}
	return fetch.Map(false, rows)
}

func (w *Where) buildSelectSQL(replace bool, cols ...string) (*sql.Rows, error) {
	if len(w.table) == 0 {
		return nil, errors.New("Where:buildSelectSQL:未指定表名")
	}

	sql := pool.Get().(*bytes.Buffer)
	defer pool.Put(sql)

	sql.Reset()
	sql.WriteString("SELECT ")
	if len(cols) == 0 {
		sql.WriteString(" * ")
	} else {
		for _, v := range cols {
			w.e.Dialect().Quote(sql, v)
			sql.WriteByte(',')
		}
		sql.Truncate(sql.Len() - 1)
	}

	sql.WriteString(" FROM ")
	w.e.Dialect().Quote(sql, w.table)
	sql.WriteString(w.cond.String())

	return w.e.Query(replace, sql.String(), w.args...)
}
