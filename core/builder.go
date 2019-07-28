// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package core

import "bytes"

// 作用于表名，列名等非关键字上的引号占位符。
// 在 Dialect.SQL 中会自动替换成该数据相应的符号。
const (
	QuoteLeft  = '{'
	QuoteRight = '}'
)

// Builder 用于构建 SQL 语句
//
// 出错时，错误信息会缓存，并在 String 和 Bytes 时返回，
// 或是通过 Err() 查看是否存在错误。
type Builder struct {
	buffer *bytes.Buffer
	err    error
}

// NewBuilder 声明一个新的 Builder 实例
func NewBuilder(str ...string) *Builder {
	b := &Builder{
		buffer: new(bytes.Buffer),
	}

	for _, s := range str {
		b.WriteString(s)
	}

	return b
}

// WriteString 写入一字符串
func (b *Builder) WriteString(str string) *Builder {
	if b.err != nil {
		return b
	}

	_, b.err = b.buffer.WriteString(str)
	return b
}

// WriteBytes 写入多个字符
func (b *Builder) WriteBytes(c ...byte) *Builder {
	for _, cc := range c {
		if b.err != nil {
			return b
		}

		b.err = b.buffer.WriteByte(cc)
	}
	return b
}

// WriteRunes 写入多个字符
func (b *Builder) WriteRunes(r ...rune) *Builder {
	for _, rr := range r {
		if b.err != nil {
			return b
		}

		_, b.err = b.buffer.WriteRune(rr)
	}
	return b
}

// Quote 给 str 左右添加 l 和 r 两个字符
func (b *Builder) Quote(str string, l, r byte) *Builder {
	return b.WriteBytes(l).WriteString(str).WriteBytes(r)
}

// QuoteKey 给 str 左右添加 QuoteLeft 和 QuoteRight 两个字符
func (b *Builder) QuoteKey(str string) *Builder {
	return b.Quote(str, QuoteLeft, QuoteRight)
}

// Reset 重置内容，同时也会将 err 设置为 nil
func (b *Builder) Reset() *Builder {
	b.buffer.Reset()
	b.err = nil
	return b
}

// TruncateLast 去掉最后几个字符
func (b *Builder) TruncateLast(n int) *Builder {
	b.buffer.Truncate(b.Len() - n)

	return b
}

// Err 返回错误内容
func (b *Builder) Err() error {
	return b.err
}

// String 获取表示的字符串
func (b *Builder) String() (string, error) {
	if b.err != nil {
		return "", b.Err()
	}
	return b.buffer.String(), nil
}

// Bytes 获取表示的字符串
func (b *Builder) Bytes() ([]byte, error) {
	if b.err != nil {
		return nil, b.Err()
	}
	return b.buffer.Bytes(), nil
}

// Len 获取长度
func (b *Builder) Len() int {
	return b.buffer.Len()
}

// Append 追加加一个 Builder 的内容
func (b *Builder) Append(v *Builder) *Builder {
	if b.err != nil {
		return b
	}

	str, err := v.String()
	if err == nil {
		b.buffer.WriteString(str)
		return b
	}
	b.err = err

	return b
}
