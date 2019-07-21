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

// Builder 对 bytes.Buffer 的一个简单封装。
// 当 Write* 系列函数出错时，直接 panic。
type Builder bytes.Buffer

// NewBuilder 声明一个新的 Builder 实例
func NewBuilder(str string) *Builder {
	return (*Builder)(bytes.NewBufferString(str))
}

func (b *Builder) buffer() *bytes.Buffer {
	return (*bytes.Buffer)(b)
}

// WriteString 写入一字符串
func (b *Builder) WriteString(str string) *Builder {
	if _, err := b.buffer().WriteString(str); err != nil {
		panic(err)
	}

	return b
}

func (b *Builder) writeByte(c byte) {
	if err := b.buffer().WriteByte(c); err != nil {
		panic(err)
	}
}

// WriteBytes 写入多个字符
func (b *Builder) WriteBytes(c ...byte) *Builder {
	for _, bb := range c {
		b.writeByte(bb)
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

// Reset 重置内容
func (b *Builder) Reset() *Builder {
	b.buffer().Reset()
	return b
}

// TruncateLast 去掉最后几个字符
func (b *Builder) TruncateLast(n int) *Builder {
	b.buffer().Truncate(b.Len() - n)

	return b
}

// String 获取表示的字符串
func (b *Builder) String() string {
	return b.buffer().String()
}

// Bytes 获取表示的字符串
func (b *Builder) Bytes() []byte {
	return b.buffer().Bytes()
}

// Len 获取长度
func (b *Builder) Len() int {
	return b.buffer().Len()
}

// Append 追加加一个 Builder 的内容
func (b *Builder) Append(v *Builder) *Builder {
	b.buffer().WriteString(v.String())
	return b
}
