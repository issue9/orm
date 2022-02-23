// SPDX-License-Identifier: MIT

// Package types 提供部分存取数据库的类型
package types

import "strings"

func isNULL(v string) bool {
	const null = "NULL"
	return v == "" || strings.ToUpper(v) == null
}
