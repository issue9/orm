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

const (
	asc = iota
	desc
)

// 以函数链的方式产生SQL语句。
type SQL struct {
	e     engine
	table string
	cond  *bytes.Buffer
	args  []interface{}
	order *bytes.Buffer
}

func newSQL(engine engine) *SQL {
	return &SQL{
		e:     engine,
		cond:  new(bytes.Buffer),
		args:  []interface{}{},
		order: new(bytes.Buffer),
	}
}

func (s *SQL) where(op int, cond string, args ...interface{}) *SQL {
	switch {
	case s.cond.Len() == 0:
		s.cond.WriteString(" WHERE(")
	case op == and:
		s.cond.WriteString(" AND(")
	case op == or:
		s.cond.WriteString(" OR(")
	default:
		panic("where:错误的参数op")
	}
	s.cond.WriteString(cond)
	s.cond.WriteByte(')')
	s.args = append(s.args, args...)

	return s
}

// 将之后的语句以and的形式与当前的语句进行连接
func (s *SQL) And(cond string, args ...interface{}) *SQL {
	return s.where(and, cond, args...)
}

// 将之后的语句以or的形式与当前的语句进行连接
func (s *SQL) Or(cond string, args ...interface{}) *SQL {
	return s.where(or, cond, args...)
}

// order by ... asc
// 当cols参数为空时，不产生任何操作。
func (s *SQL) Asc(cols ...string) *SQL {
	if len(cols) == 0 {
		return s
	}

	for _, col := range cols {
		s.order.WriteString(col)
		s.order.WriteByte(',')
	}
	s.order.Truncate(s.order.Len() - 1)
	s.order.WriteString(" ASC,")
	return s
}

// order by ... desc
// 当cols参数为空时，不产生任何操作。
func (s *SQL) Desc(cols ...string) *SQL {
	if len(cols) == 0 {
		return s
	}

	for _, col := range cols {
		s.order.WriteString(col)
		s.order.WriteByte(',')
	}
	s.order.Truncate(s.order.Len() - 1)
	s.order.WriteString(" DESC,")
	return s
}

// 指定表名。
func (s *SQL) Table(tableName string) *SQL {
	s.table = tableName
	return s
}

// 将符合当前条件的所有记录删除。
// replace，是否需将对语句的占位符进行替换。
func (s *SQL) Delete(replace bool) error {
	if len(s.table) == 0 {
		return errors.New("SQL.Delete:未指定表名")
	}

	sql := pool.Get().(*bytes.Buffer)
	defer pool.Put(sql)

	sql.Reset()
	sql.WriteString("DELETE FROM ")
	s.e.Dialect().Quote(sql, s.table)
	sql.WriteString(s.cond.String())
	_, err := s.e.Exec(replace, sql.String(), s.args...)
	return err
}

// 更新符合当前条件的所有记录。
// replace，是否需将对语句的占位符进行替换。
func (s *SQL) Update(replace bool, data map[string]interface{}) error {
	if len(s.table) == 0 {
		return errors.New("SQL.Update:未指定表名")
	}

	if len(data) == 0 {
		return errors.New("SQL.Update:未指定需要更新的数据")
	}

	sql := pool.Get().(*bytes.Buffer)
	defer pool.Put(sql)

	sql.Reset()
	sql.WriteString("UPDATE ")
	s.e.Dialect().Quote(sql, s.table)
	vals := make([]interface{}, 0, len(data)+len(s.args))
	sql.WriteString(" SET ")
	for k, v := range data {
		s.e.Dialect().Quote(sql, k)
		sql.WriteString("=?,")
		vals = append(vals, v)
	}
	sql.Truncate(sql.Len() - 1) // 去掉最后一个逗号

	sql.WriteString(s.cond.String())
	_, err := s.e.Exec(replace, sql.String(), append(vals, s.args...)...)
	return err
}

// 将符合当前条件的所有记录依次写入objs中。
// replace，是否需将对语句的占位符进行替换。
func (s *SQL) Select(replace bool, objs interface{}) error {
	rows, err := s.buildSelectSQL(replace)
	if err != nil {
		return err
	}
	return fetch.Obj(objs, rows)
}

// 返回符合当前条件的所有记录。
// replace，是否需将对语句的占位符进行替换。
func (s *SQL) SelectMap(replace bool, cols ...string) ([]map[string]interface{}, error) {
	rows, err := s.buildSelectSQL(replace, cols...)
	if err != nil {
		return nil, err
	}
	return fetch.Map(false, rows)
}

func (s *SQL) buildSelectSQL(replace bool, cols ...string) (*sql.Rows, error) {
	if len(s.table) == 0 {
		return nil, errors.New("SQL:buildSelectSQL:未指定表名")
	}

	sql := pool.Get().(*bytes.Buffer)
	defer pool.Put(sql)

	sql.Reset()
	sql.WriteString("SELECT ")
	if len(cols) == 0 {
		sql.WriteString(" * ")
	} else {
		for _, v := range cols {
			s.e.Dialect().Quote(sql, v)
			sql.WriteByte(',')
		}
		sql.Truncate(sql.Len() - 1)
	}

	// from
	sql.WriteString(" FROM ")
	s.e.Dialect().Quote(sql, s.table)

	// where
	sql.WriteString(s.cond.String())

	// order
	if s.order.Len() > 0 {
		sql.WriteString(" ORDER BY ")
		s.order.WriteTo(sql)
		sql.Truncate(sql.Len() - 1)
	}

	return s.e.Query(replace, sql.String(), s.args...)
}
