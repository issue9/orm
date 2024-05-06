// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package fetch

import (
	"database/sql"

	"github.com/issue9/orm/v6/core"
)

// Column 导出 rows 中某列的所有或一行数据
//
// once 若为 true，则只导出第一条数据。
// colName 指定需要导出的列名，若指定了不存在的名称，返回 error。
//
// NOTE: 要求 T 的类型必须符合 [sql.Row.Scan] 的参数要求；
func Column[T any](once bool, colName string, rows *sql.Rows) ([]T, error) {
	// TODO: 应该约束 T 为 sql.Rows.Scan 允许的类型，但是以目前 Go 的语法无法做到。

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	index := -1 // colName 列在 rows.Columns() 中的索引号
	buff := make([]any, len(cols))
	for i, v := range cols {
		if colName == v { // 获取 index 的值
			index = i
			var zero T
			buff[i] = &zero
		} else {
			var value any
			buff[i] = &value
		}
	}

	if index == -1 {
		return nil, core.ErrColumnNotFound(colName)
	}

	var data []T
	for rows.Next() {
		if err := rows.Scan(buff...); err != nil {
			return nil, err
		}
		data = append(data, *buff[index].(*T))
		if once {
			return data, nil
		}
	}

	return data, nil
}
