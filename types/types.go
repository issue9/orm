// SPDX-License-Identifier: MIT

// Package types 提供部分存取数据库的类型
package types

import "strings"

const null = "NULL"

func isNULL(v string) bool {
	return v == "" || strings.ToUpper(v) == null
}
