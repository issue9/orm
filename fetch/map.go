// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package fetch

import (
	"database/sql"
	"reflect"
)

// Map 将 rows 中的所有或一行数据导出到 map[string]interface{} 中。
// 若 once 值为 true，则只导出第一条数据。
//
// NOTE:
// 每个数据库对数据的处理方式是不一样的，比如如下语句
//  SELECT COUNT(*) as cnt FROM tbl1
// 将使用 Map() 导出到 []map[string]interface{} 中时，
// 在 mysql 中，cnt 有可能被处理成一个 []byte (打印输出时，像一个数组，容易造成困惑)，
// 而在 sqlite3 就有可能是个 int。
func Map(once bool, rows *sql.Rows) ([]map[string]interface{}, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// 临时缓存，用于保存从 rows 中读取出来的一行。
	buff := make([]interface{}, len(cols))
	for i := range cols {
		var value interface{}
		buff[i] = &value
	}

	var data []map[string]interface{}
	for rows.Next() {
		if err := rows.Scan(buff...); err != nil {
			return nil, err
		}

		line := make(map[string]interface{}, len(cols))
		for i, v := range cols {
			if buff[i] == nil {
				continue
			}
			value := reflect.Indirect(reflect.ValueOf(buff[i]))
			line[v] = value.Interface()
		}

		data = append(data, line)
		if once {
			return data, nil
		}
	}

	return data, nil
}

// MapString 将 rows 中的数据导出到一个 map[string]string 中。
// 功能上与 Map() 上一样，但 map 的键值固定为 string。
func MapString(once bool, rows *sql.Rows) (data []map[string]string, err error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	buf := make([]interface{}, len(cols))
	for k := range buf {
		var val string
		buf[k] = &val
	}

	for rows.Next() {
		if err = rows.Scan(buf...); err != nil {
			return nil, err
		}

		line := make(map[string]string, len(cols))
		for i, v := range cols {
			if buf[i] == nil {
				continue
			}
			line[v] = *(buf[i].(*string))
		}

		data = append(data, line)

		if once {
			return data, nil
		}
	}
	return data, nil
}
