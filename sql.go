// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"bytes"
	"database/sql"
	"errors"
	"strconv"
	"strings"

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

var (
	ErrEmptyTableName = errors.New("未指定表名")
)

// 以函数链的方式产生SQL语句。
type SQL struct {
	e        engine
	table    string
	cond     *bytes.Buffer
	condArgs []interface{}
	order    *bytes.Buffer
	limit    []int // 元素0为limit，元素1为offset
}

func newSQL(engine engine) *SQL {
	return &SQL{
		e:        engine,
		cond:     new(bytes.Buffer),
		condArgs: []interface{}{},
		order:    new(bytes.Buffer),
		limit:    make([]int, 0, 2), // 最长2个，同时保存limit和offset
	}
}

func (s *SQL) Reset() *SQL {
	s.table = ""
	s.cond.Reset()
	s.condArgs = s.condArgs[:0]
	s.order.Reset()
	s.limit = s.limit[:0]

	return s
}

// 指定Limit相关的值。
func (s *SQL) Limit(limit int, offset ...int) *SQL {
	if len(offset) > 1 {
		panic("offset参数指定了太多的值")
	}

	s.limit = append(s.limit, limit)

	if len(offset) > 0 {
		s.limit = append(s.limit, offset[0])
	}

	return s
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
	s.condArgs = append(s.condArgs, args...)

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
func (s *SQL) Delete(replace bool) (sql.Result, error) {
	if len(s.table) == 0 {
		return nil, ErrEmptyTableName
	}

	sql := bytes.NewBufferString("DELETE FROM ")
	s.e.Dialect().Quote(sql, s.table)
	sql.WriteString(s.cond.String())
	return s.e.Exec(replace, sql.String(), s.condArgs...)
}

// 更新符合当前条件的所有记录。
// replace，是否需将对语句的占位符进行替换。
func (s *SQL) Update(replace bool, data map[string]interface{}) (sql.Result, error) {
	if len(s.table) == 0 {
		return nil, ErrEmptyTableName
	}

	if len(data) == 0 {
		return nil, errors.New("SQL.Update:未指定需要更新的数据")
	}

	sql := bytes.NewBufferString("UPDATE ")
	s.e.Dialect().Quote(sql, s.table)
	vals := make([]interface{}, 0, len(data)+len(s.condArgs))
	sql.WriteString(" SET ")
	for k, v := range data {
		s.e.Dialect().Quote(sql, k)
		sql.WriteString("=?,")
		vals = append(vals, v)
	}
	sql.Truncate(sql.Len() - 1) // 去掉最后一个逗号

	sql.WriteString(s.cond.String())
	return s.e.Exec(replace, sql.String(), append(vals, s.condArgs...)...)
}

// 返回符合条件的记录数量。
// 若指定了Limit，则相应的条件也会计算在内。
func (s *SQL) Count(replace bool) (int, error) {
	rows, err := s.query(replace, "COUNT(*) AS cnt")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	cols, err := fetch.ColumnString(true, "cnt", rows)
	if err != nil {
		return 0, err
	}

	if len(cols) == 0 {
		return 0, nil
	}

	return strconv.Atoi(cols[0])
}

// 将符合当前条件的所有记录依次写入objs中。
// replace，是否需将对语句的占位符进行替换。
func (s *SQL) Select(replace bool, objs interface{}) (int, error) {
	rows, err := s.query(replace)
	if err != nil {
		return 0, err
	}
	cnt, err := fetch.Obj(objs, rows)
	rows.Close()
	return cnt, err
}

// 返回符合当前条件的所有记录。
// replace，是否需将对语句的占位符进行替换。
func (s *SQL) SelectMap(replace bool, cols ...string) ([]map[string]interface{}, error) {
	rows, err := s.query(replace, cols...)
	if err != nil {
		return nil, err
	}
	mapped, err := fetch.Map(false, rows)
	rows.Close()
	return mapped, err
}

// 返回符合当前条件的所有记录。
// replace，是否需将对语句的占位符进行替换。
func (s *SQL) SelectMapString(replace bool, cols ...string) ([]map[string]string, error) {
	rows, err := s.query(replace, cols...)
	if err != nil {
		return nil, err
	}
	mapped, err := fetch.MapString(false, rows)
	rows.Close()
	return mapped, err
}

func (s *SQL) query(replace bool, cols ...string) (*sql.Rows, error) {
	if len(s.table) == 0 {
		return nil, ErrEmptyTableName
	}

	args := make([]interface{}, 0, len(s.condArgs)) // Query对应的参数
	sql := bytes.NewBufferString("SELECT ")
	if len(cols) == 0 {
		sql.WriteString(" * ")
	} else {
		sql.WriteString(strings.Join(cols, ","))
	}

	// from
	sql.WriteString(" FROM ")
	s.e.Dialect().Quote(sql, s.table)

	// where
	sql.WriteString(s.cond.String())
	args = append(args, s.condArgs...)

	// order
	if s.order.Len() > 0 {
		sql.WriteString(" ORDER BY ")
		s.order.WriteTo(sql)
		sql.Truncate(sql.Len() - 1)
	}

	// limit
	if len(s.limit) > 0 {
		vals, err := s.e.Dialect().LimitSQL(sql, s.limit[0], s.limit[1:]...)
		if err != nil {
			return nil, err
		}

		args = append(args, vals[0])
		if len(vals) > 1 {
			args = append(args, vals[1])
		}
	}

	return s.e.Query(replace, sql.String(), args...)
}
