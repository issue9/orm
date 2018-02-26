// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import "bytes"

// SQL 定义 SQL 语句的基本接口
type SQL interface {
	// 获取 SQL 语句以及其关联的参数
	SQL() (query string, args []interface{}, err error)

	// 重置整个 SQL 语句。
	Reset()
}

// 对 bytes.Buffer 的一个简单封装。
// 当 Write* 系列函数出错时，直接 panic。
type stringBuilder bytes.Buffer

func newStringBuilder(str string) *stringBuilder {
	return (*stringBuilder)(bytes.NewBufferString(str))
}

func (b *stringBuilder) writeString(str string) {
	if _, err := (*bytes.Buffer)(b).WriteString(str); err != nil {
		panic(err)
	}
}

func (b *stringBuilder) writeByte(c byte) {
	if err := (*bytes.Buffer)(b).WriteByte(c); err != nil {
		panic(err)
	}
}

func (b *stringBuilder) reset() {
	(*bytes.Buffer)(b).Reset()
}

func (b *stringBuilder) string() string {
	return (*bytes.Buffer)(b).String()
}

func (b *stringBuilder) len() int {
	return (*bytes.Buffer)(b).Len()
}

func (b *stringBuilder) truncateLast(n int) {
	(*bytes.Buffer)(b).Truncate(b.len() - n)
}
