// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package core

import (
	"strings"

	"github.com/issue9/errwrap"
)

// 作用于表名，列名等非关键字上的引号占位符。
// 在执行会自动替换成该数据库相应的符号。
const (
	QuoteLeft  = '{'
	QuoteRight = '}'
)

// Builder 用于构建 SQL 语句
//
// 出错时，错误信息会缓存，并在 [Builder.String] 或 [Builder.Bytes] 时返回，
// 或是通过 [Builder.Err] 查看是否存在错误。
type Builder struct {
	buffer errwrap.Buffer
}

// NewBuilder 声明一个新的 [Builder] 实例
func NewBuilder(str ...string) *Builder {
	b := &Builder{}

	for _, s := range str {
		b.WString(s)
	}

	return b
}

// WString 写入一字符串
func (b *Builder) WString(str string) *Builder {
	b.buffer.WString(str)
	return b
}

// WBytes 写入多个字符
func (b *Builder) WBytes(c ...byte) *Builder {
	b.buffer.WBytes(c)
	return b
}

// WRunes 写入多个字符
func (b *Builder) WRunes(r ...rune) *Builder {
	b.buffer.WRunes(r)
	return b
}

// Quote 给 str 左右添加 l 和 r 两个字符
func (b *Builder) Quote(str string, l, r byte) *Builder { return b.WBytes(l).WString(str).WBytes(r) }

// QuoteKey 给 str 左右添加 [QuoteLeft] 和 [QuoteRight] 两个字符
func (b *Builder) QuoteKey(str string) *Builder { return b.Quote(str, QuoteLeft, QuoteRight) }

// QuoteColumn 为列名添加 [QuoteLeft] 和 [QuoteRight] 两个字符
//
// NOTE: 列名可能包含表名或是表名别名：table.col
func (b *Builder) QuoteColumn(col string) *Builder {
	if index := strings.IndexByte(col, '.'); index > 0 {
		return b.QuoteKey(col[:index]).WBytes('.').QuoteKey(col[index+1:])
	}
	return b.Quote(col, QuoteLeft, QuoteRight)
}

// Reset 重置内容，同时也会将 err 设置为 nil
func (b *Builder) Reset() *Builder {
	b.buffer.Reset()
	return b
}

// TruncateLast 去掉最后几个字符
func (b *Builder) TruncateLast(n int) *Builder {
	b.buffer.Truncate(b.Len() - n)
	return b
}

// Err 返回错误内容
func (b *Builder) Err() error { return b.buffer.Err }

// String 获取表示的字符串
func (b *Builder) String() (string, error) {
	if b.Err() != nil {
		return "", b.Err()
	}
	return b.buffer.String(), nil
}

// Bytes 获取表示的字符串
func (b *Builder) Bytes() ([]byte, error) {
	if b.Err() != nil {
		return nil, b.Err()
	}
	return b.buffer.Bytes(), nil
}

// Len 获取长度
func (b *Builder) Len() int { return b.buffer.Len() }

// Append 追加加一个 [Builder] 的内容
func (b *Builder) Append(v *Builder) *Builder {
	if b.Err() != nil {
		return b
	}

	str, err := v.String()
	if err == nil {
		return b.WString(str)
	}
	b.buffer.Err = err

	return b
}
