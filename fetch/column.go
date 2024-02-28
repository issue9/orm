// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package fetch

import (
	"database/sql"
	"fmt"
	"reflect"
)

func columnNotExists(col string) error {
	return fmt.Errorf("指定的列名 %s 不存在", col)
}

// Column 导出 rows 中某列的所有或一行数据
//
// once 若为 true，则只导出第一条数据。
// colName 指定需要导出的列名，若指定了不存在的名称，返回 error。
func Column(once bool, colName string, rows *sql.Rows) ([]any, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	index := -1 // colName 列在 rows.Columns() 中的索引号
	buff := make([]any, len(cols))
	for i, v := range cols {
		var value any
		buff[i] = &value

		if colName == v { // 获取 index 的值
			index = i
		}
	}

	if index == -1 {
		return nil, columnNotExists(colName)
	}

	var data []any
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

// ColumnString 导出 rows 中某列的所有或是一行数据
//
// 功能等同于 Column() 函数，但是返回值是 []string 而不是 []interface{}。
func ColumnString(once bool, colName string, rows *sql.Rows) ([]string, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	index := -1 // colName 列在 rows.Columns() 中的索引号
	buff := make([]any, len(cols))
	for i, v := range cols {
		var value string
		buff[i] = &value

		if colName == v { // 获取 index 的值
			index = i
		}
	}

	if index == -1 {
		return nil, columnNotExists(colName)
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
