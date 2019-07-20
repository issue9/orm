// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package dialect 提供了部分数据库对 orm.Dialect 接口的实现。
package dialect

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	errColIsNil              = errors.New("参数 col 参数是个空值")
	errGoTypeIsNil           = errors.New("无效的 col.GoType 值")
	errMissLength            = errors.New("缺少长度")
	errTimeFractionalInvalid = errors.New("时间精度只能介于 [0,6] 之间")
)

func errUncovert(dest string) error {
	return fmt.Errorf("不支持的类型: %s", dest)
}

// mysql 系列数据库分页语法的实现。支持以下数据库：
// MySQL, H2, HSQLDB, Postgres, SQLite3
func mysqlLimitSQL(limit interface{}, offset ...interface{}) (string, []interface{}) {
	query := " LIMIT "

	if named, ok := limit.(sql.NamedArg); ok && named.Name != "" {
		query += "@" + named.Name
	} else {
		query += "?"
	}

	if len(offset) == 0 {
		return query + " ", []interface{}{limit}
	}

	query += " OFFSET "
	o := offset[0]
	if named, ok := o.(sql.NamedArg); ok && named.Name != "" {
		query += "@" + named.Name
	} else {
		query += "?"
	}

	return query + " ", []interface{}{limit, offset[0]}
}

// oracle 系列数据库分页语法的实现。支持以下数据库：
// Derby, SQL Server 2012, Oracle 12c, the SQL 2008 standard
func oracleLimitSQL(limit interface{}, offset ...interface{}) (string, []interface{}) {
	query := "FETCH NEXT "

	if named, ok := limit.(sql.NamedArg); ok && named.Name != "" {
		query += "@" + named.Name
	} else {
		query += "?"
	}
	query += " ROWS ONLY "

	if len(offset) == 0 {
		return query, []interface{}{limit}
	}

	o := offset[0]
	if named, ok := o.(sql.NamedArg); ok && named.Name != "" {
		query = "OFFSET @" + named.Name + " ROWS " + query
	} else {
		query = "OFFSET ? ROWS " + query
	}

	return query, []interface{}{offset[0], limit}
}

type namedArg struct {
	sql.NamedArg
	index int
}

func replaceNamedArgs(query string, args []interface{}) string {
	as := make([]namedArg, 0, len(args))

	for index, arg := range args {
		if named, ok := arg.(sql.NamedArg); ok {
			as = append(as, namedArg{
				NamedArg: named,
				index:    index,
			})
			continue
		}

		if named, ok := arg.(*sql.NamedArg); ok {
			as = append(as, namedArg{
				NamedArg: *named,
				index:    index,
			})
		}
	}

	// 将名称长的排到前面，确保可以正确替换
	sort.SliceStable(as, func(i, j int) bool {
		return len(as[i].Name) > len(as[j].Name)
	})

	for _, arg := range as {
		query = strings.Replace(query, "@"+arg.Name, "?", 1)
		args[arg.index] = arg.Value
	}

	return query
}
