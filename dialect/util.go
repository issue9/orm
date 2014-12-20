// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"database/sql"
	"reflect"
	"time"
)

const (
	pkName = "pk" // 默认的主键约束名
)

var (
	nullString  = reflect.TypeOf(sql.NullString{})
	nullInt64   = reflect.TypeOf(sql.NullInt64{})
	nullBool    = reflect.TypeOf(sql.NullBool{})
	nullFloat64 = reflect.TypeOf(sql.NullFloat64{})
	timeType    = reflect.TypeOf(time.Time{})
)

// mysq系列数据库分页语法的实现。支持以下数据库：
// MySQL, H2, HSQLDB, Postgres, SQLite3
func mysqlLimitSQL(limit int, offset ...int) (string, []interface{}) {
	if len(offset) == 0 {
		return " LIMIT ? ", []interface{}{limit}
	}

	return " LIMIT ? OFFSET ? ", []interface{}{limit, offset[0]}
}

// oracle系列数据库分页语法的实现。支持以下数据库：
// Derby, SQL Server 2012, Oracle 12c, the SQL 2008 standard
func oracleLimitSQL(limit int, offset ...int) (string, []interface{}) {
	if len(offset) == 0 {
		return " FETCH NEXT ? ROWS ONLY ", []interface{}{limit}
	}

	return " OFFSET ? ROWS FETCH NEXT ? ROWS ONLY ", []interface{}{offset[0], limit}
}
