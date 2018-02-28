// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package sqlbuilder 用于构建 SQL 语句
package sqlbuilder

import "errors"

var (
	// ErrTableIsEmpty 未指定表名，任何 SQL 语句中，
	// 若未指定表名时，会返回此错误
	ErrTableIsEmpty = errors.New("表名为空")

	// ErrValueIsEmpty 在 Update 和 Insert 语句中，
	// 若未指定任何值，则返回此错误
	ErrValueIsEmpty = errors.New("值为空")

	// ErrColumnsIsEmpty 在 Insert 和 Select 语句中，
	// 若未指定任何列表，则返回此错误
	ErrColumnsIsEmpty = errors.New("未指定列")

	// ErrArgsNotMatch 在生成的 SQL 语句中，传递的参数与语句的占位符数量不匹配。
	ErrArgsNotMatch = errors.New("列与值的数量不匹配")
)

// SQL 定义 SQL 语句的基本接口
type SQL interface {
	// 获取 SQL 语句以及其关联的参数
	SQL() (query string, args []interface{}, err error)

	// 重置整个 SQL 语句。
	Reset()
}
