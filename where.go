// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"bytes"
	"errors"

	"github.com/issue9/orm/fetch"
)

const (
	and = iota
	or
)

type Where struct {
	db    db
	table string
	cond  *bytes.Buffer
	args  []interface{}
}

func newWhere(db db) *Where {
	return &Where{
		db:   db,
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
		panic("Where:where:错误的参数op")
	}
	w.cond.WriteString(cond)
	w.cond.WriteByte(')')
	return w
}

func (w *Where) And(cond string, args ...interface{}) *Where {
	return w.where(and, cond, args...)
}

func (w *Where) Or(cond string, args ...interface{}) *Where {
	return w.where(or, cond, args...)
}

func (w *Where) Table(tableName string) *Where {
	w.table = tableName
	return w
}

func (w *Where) Delete() error {
	if len(w.table) == 0 {
		return errors.New("Where:Delete:未指定表名")
	}

	sql := bytes.NewBufferString("DELETE FROM ")
	w.db.Dialect().Quote(sql, w.table)
	sql.WriteString(w.cond.String())
	_, err := w.db.Exec(sql.String(), w.args...)
	return err
}

func (w *Where) Update(data map[string]interface{}) error {
	if len(w.table) == 0 {
		return errors.New("Where:Update:未指定表名")
	}

	if len(data) == 0 {
		return errors.New("Where:Update:未指定需要更新的数据")
	}

	sql := bytes.NewBufferString("UPDATE ")
	w.db.Dialect().Quote(sql, w.table)
	sql.WriteString(" SET ")
	for k, v := range data {
		w.db.Dialect().Quote(sql, k)
		sql.WriteByte('=')
		AsString(sql, v)
		sql.WriteByte(',')
	}
	sql.Truncate(sql.Len() - 1) // 去掉最后一个逗号

	sql.WriteString(w.cond.String())
	_, err := w.db.Exec(sql.String(), w.args...)
	return err
}

func (w *Where) Select(cols ...string) ([]map[string]interface{}, error) {
	if len(w.table) == 0 {
		return nil, errors.New("Where:Select:未指定表名")
	}

	if len(cols) == 0 {
		cols = append(cols, "*")
	}

	sql := bytes.NewBufferString("SELECT ")
	for _, v := range cols {
		w.db.Dialect().Quote(sql, v)
		sql.WriteByte(',')
	}
	sql.Truncate(sql.Len() - 1)

	sql.WriteString(" FROM ")
	w.db.Dialect().Quote(sql, w.table)
	sql.WriteByte(' ')
	sql.WriteString(w.cond.String())

	rows, err := w.db.Query(sql.String(), w.args...)
	if err != nil {
		return nil, err
	}
	return fetch.Map(false, rows)
}
