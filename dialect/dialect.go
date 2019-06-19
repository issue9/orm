// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package dialect 提供了部分数据库对 orm.Dialect 接口的实现。
package dialect

import (
	"database/sql"
	"reflect"
	"time"

	"github.com/issue9/orm/v2"
	"github.com/issue9/orm/v2/sqlbuilder"
)

const pkName = "pk" // 默认的主键约束名

var (
	nullString  = reflect.TypeOf(sql.NullString{})
	nullInt64   = reflect.TypeOf(sql.NullInt64{})
	nullBool    = reflect.TypeOf(sql.NullBool{})
	nullFloat64 = reflect.TypeOf(sql.NullFloat64{})
	rawBytes    = reflect.TypeOf(sql.RawBytes{})
	timeType    = reflect.TypeOf(time.Time{})
)

type base interface {
	orm.Dialect

	// 将 col 转换成 sql 类型，并写入 buf 中。
	sqlType(buf *sqlbuilder.SQLBuilder, col *orm.Column) error
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

// oracle系列数据库分页语法的实现。支持以下数据库：
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
