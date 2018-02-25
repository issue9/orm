// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package fetch

import (
	"database/sql"
	"fmt"
	"reflect"
)

// Column 导出 rows 中某列的所有或一行数据。
// once 若为 true，则只导出第一条数据。
// colName 指定需要导出的列名，若不指定了不存在的名称，返回 error。
func Column(once bool, colName string, rows *sql.Rows) ([]interface{}, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	index := -1 // colName 列在 rows.Columns() 中的索引号
	buff := make([]interface{}, len(cols))
	for i, v := range cols {
		var value interface{}
		buff[i] = &value

		if colName == v { // 获取index的值
			index = i
		}
	}

	if index == -1 {
		return nil, fmt.Errorf("Column:指定的列名[%v]不存在", colName)
	}

	var data []interface{}
	for rows.Next() {
		if err := rows.Scan(buff...); err != nil {
			return nil, err
		}
		value := reflect.Indirect(reflect.ValueOf(buff[index]))
		data = append(data, value.Interface())
		if once {
			return data, nil
		}
	}

	return data, nil
}

// ColumnString 导出 rows 中某列的所有或是一行数据。
// 功能等同于 Columns() 函数，但是返回值是 []string 而不是 []interface{}。
func ColumnString(once bool, colName string, rows *sql.Rows) ([]string, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	index := -1 // colName列在rows.Columns()中的索引号
	buff := make([]interface{}, len(cols))
	for i, v := range cols {
		var value string
		buff[i] = &value

		if colName == v { // 获取index的值
			index = i
		}
	}

	if index == -1 {
		return nil, fmt.Errorf("Column:指定的列名[%v]不存在", colName)
	}

	var data []string
	for rows.Next() {
		if err := rows.Scan(buff...); err != nil {
			return nil, err
		}
		data = append(data, *(buff[index].(*string)))
		if once {
			return data, nil
		}
	}

	return data, nil
}
