// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package stringbuilder 实现了字符串拼接的一些基本操作
package stringbuilder

import "bytes"

// StringBuilder 对 bytes.Buffer 的一个简单封装。
// 当 Write* 系列函数出错时，直接 panic。
type StringBuilder bytes.Buffer

// New 声明一个新的 StringBuilder 实例
func New(str string) *StringBuilder {
	return (*StringBuilder)(bytes.NewBufferString(str))
}

func (b *StringBuilder) buffer() *bytes.Buffer {
	return (*bytes.Buffer)(b)
}

// WriteString 写入一字符串
func (b *StringBuilder) WriteString(str string) *StringBuilder {
	if _, err := b.buffer().WriteString(str); err != nil {
		panic(err)
	}

	return b
}

// WriteByte 写入一字符
func (b *StringBuilder) WriteByte(c byte) *StringBuilder {
	if err := b.buffer().WriteByte(c); err != nil {
		panic(err)
	}

	return b
}

// Reset 重置内容
func (b *StringBuilder) Reset() *StringBuilder {
	b.buffer().Reset()
	return b
}

// TruncateLast 去掉最后几个字符
func (b *StringBuilder) TruncateLast(n int) *StringBuilder {
	b.buffer().Truncate(b.Len() - n)

	return b
}

// String 获取表示的字符串
func (b *StringBuilder) String() string {
	return b.buffer().String()
}

// Len 获取长度
func (b *StringBuilder) Len() int {
	return b.buffer().Len()
}
