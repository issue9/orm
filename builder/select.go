// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package builder

import (
	"database/sql"
	"errors"
	"fmt"
	//"reflect"
	"strings"

	"github.com/issue9/orm/fetch"
)

// 指定列名。select语句可以用此方法指定需要获取的列。
func (s *SQL) Columns(cols ...string) *SQL {
	s.cols = append(s.cols, cols...)

	return s
}

var joinType = []string{" LEFT JOIN ", " RIGHT JOIN ", " INNER JOIN ", " FULL JOIN "}

// join功能
func (s *SQL) joinOn(typ int, table string, on string) *SQL {
	if typ < 0 && typ > 4 {
		s.errors = append(s.errors, fmt.Errorf("joinOn:错误的typ值:[%v]", typ))
	}

	s.join.WriteString(joinType[typ])
	s.join.WriteString(table)
	s.join.WriteString(" ON ")
	s.join.WriteString(on)

	return s
}

// LEFT JOIN ... ON ...
func (s *SQL) LeftJoin(table, on string) *SQL {
	return s.joinOn(0, table, on)
}

// RIGHT JOIN ... ON ...
func (s *SQL) RightJoin(table, on string) *SQL {
	return s.joinOn(1, table, on)
}

// INNER JOIN ... ON ...
func (s *SQL) InnerJoin(table, on string) *SQL {
	return s.joinOn(2, table, on)
}

// FULL JOIN ... ON ...
func (s *SQL) FullJoin(table, on string) *SQL {
	return s.joinOn(3, table, on)
}

var orderType = []string{" ASC ", " DESC "}

// 供Asc()和Desc()使用。
// sort: 0=asc,1=desc，其它值无效
func (s *SQL) orderBy(sort int, col string) *SQL {
	if sort != 0 && sort != 1 {
		s.errors = append(s.errors, fmt.Errorf("orderBy:错误的sort参数:[%v]", sort))
	}

	if s.order.Len() == 0 {
		s.order.WriteString(" ORDER BY ")
	} else {
		s.order.WriteString(", ")
	}

	s.order.WriteString(col)
	s.order.WriteString(orderType[sort])

	return s
}

// ORDER BY ... ASC
func (s *SQL) Asc(cols ...string) *SQL {
	for _, c := range cols {
		s.orderBy(0, c)
	}

	return s
}

// ORDER BY ... DESC
func (s *SQL) Desc(cols ...string) *SQL {
	for _, c := range cols {
		s.orderBy(1, c)
	}

	return s
}

// LIMIT ... OFFSET ...
func (s *SQL) Limit(limit interface{}, offset ...interface{}) *SQL {
	if len(offset) > 1 {
		s.errors = append(s.errors, errors.New("Limit:指定了太多的参数"))
	}

	s.limitSQL = s.db.Dialect().LimitSQL(limit, offset...)
	return s
}

// 产生SELECT语句
func (s *SQL) selectSQL() string {
	s.buf.Reset()

	s.buf.WriteString("SELECT ")
	s.buf.WriteString(strings.Join(s.cols, ","))
	s.buf.WriteString(" FROM ")
	s.buf.WriteString(s.tableName)
	s.buf.WriteString(s.join.String())
	s.buf.WriteString(s.cond.String())  // where
	s.buf.WriteString(s.order.String()) // NOTE(caixw):mysql中若要limit，order字段是必须提供的
	s.buf.WriteString(s.limitSQL)

	return s.buf.String()
}

func (s *SQL) Query(args map[string]interface{}) (*sql.Rows, error) {
	if s.HasErrors() {
		return nil, Errors(s.errors)
	}

	return s.db.Query(s.selectSQL(), args)
}

// 导出数据到map[string]interface{}
func (s *SQL) Fetch2Map(args map[string]interface{}) (map[string]interface{}, error) {
	rows, err := s.Query(args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data, err := fetch.Map(true, rows)
	if err != nil {
		return nil, err
	}

	return data[0], nil
}

// 导出所有数据到[]map[string]interface{}
func (s *SQL) Fetch2Maps(args map[string]interface{}) ([]map[string]interface{}, error) {
	rows, err := s.Query(args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return fetch.Map(false, rows)
}

// 返回指定列的第一行内容
func (s *SQL) FetchColumn(col string, args map[string]interface{}) (interface{}, error) {
	rows, err := s.Query(args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data, err := fetch.Column(true, col, rows)
	if err != nil {
		return nil, err
	}

	return data[0], nil
}

// 返回指定列的所有数据
func (s *SQL) FetchColumns(col string, args map[string]interface{}) ([]interface{}, error) {
	rows, err := s.Query(args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return fetch.Column(false, col, rows)
}

// 将当前select语句查询的数据导出到v中
// v可以是orm/fetch.Obj中允许的类型。
func (s *SQL) FetchObj(v interface{}, args map[string]interface{}) error {
	rows, err := s.Query(args)
	if err != nil {
		return err
	}
	defer rows.Close()

	return fetch.Obj(v, rows)
}
