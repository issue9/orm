// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

// SQL 定义 SQL 语句的基本接口
type SQL interface {
	// 获取 SQL 语句以及其关联的参数
	SQL() (query string, args []interface{}, err error)

	// 重置整个 SQL 语句。
	Reset()
}
