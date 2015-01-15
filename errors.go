// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"fmt"
)

// 执行SQL发生错误时，保存的信息。
type SQLError struct {
	Err        error  // 原来的错误信息
	DriverName string // 对应的数据为驱动名称

	OriginSQL string        // 原始的SQL语句，没有转换之前的
	SQL       string        // 转换后的SQL
	Args      []interface{} // SQL对应的参数
}

func newSQLError(err error, driverName, originSQL, sql string, args ...interface{}) error {
	return &SQLError{
		Err:        err,
		DriverName: driverName,
		OriginSQL:  originSQL,
		SQL:        sql,
		Args:       args,
	}
}

func (err *SQLError) Error() string {
	format := `SQLError:原始错误信息:%v;
driverName:%v;
原始语句:%v;
sql语句:%v;
对应参数:%v`
	return fmt.Sprintf(format, err.Err, err.DriverName, err.OriginSQL, err.SQL, err.Args)
}
