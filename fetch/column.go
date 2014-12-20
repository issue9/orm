// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package fetch

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
)

// 导出rows中某列的所有或一行数据。
// once若为true，则只导出第一条数据。
// colName指定需要导出的列名，若不指定了不存在的名称，返回error。
func Column(once bool, colName string, rows *sql.Rows) ([]interface{}, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	index := -1 // colName列在rows.Columns()中的索引号
	buff := make([]interface{}, len(cols))
	for i, v := range cols {
		var value interface{}
		buff[i] = &value

		if colName == v { // 获取index的值
			index = i
		}
	}

	if index == -1 {
		return nil, errors.New("指定的名不存在")
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

// 导出rows中某列的所有或是一行数据。
// 功能等同于Columns()函数，但是返回值是[]string而不是[]interface{}。
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
		return nil, fmt.Errorf("指定的名[%v]不存在", colName)
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
