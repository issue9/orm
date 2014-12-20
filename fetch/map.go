// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package fetch

import (
	"database/sql"
	"reflect"
)

// 将rows中的所有或一行数据导出到map[string]interface{}中。
// 若once值为true，则只导出第一条数据。
func Map(once bool, rows *sql.Rows) ([]map[string]interface{}, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// 临时缓存，用于保存从rows中读取出来的一行。
	buff := make([]interface{}, len(cols))
	for i, _ := range cols {
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

// 将rows中的数据导出到一个map[string]string中。
// 功能上与Map()上一样，但map的键值固定为string。
func MapString(once bool, rows *sql.Rows) (data []map[string]string, err error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	buf := make([]interface{}, len(cols))
	for k, _ := range buf {
		var val string
		buf[k] = &val
	}

	for rows.Next() {
		if err = rows.Scan(buf...); err != nil {
			return nil, err
		}

		line := make(map[string]string, len(cols))
		for i, v := range cols {
			line[v] = *(buf[i].(*string))
		}

		data = append(data, line)

		if once {
			return data, nil
		}
	}
	return data, nil
}
