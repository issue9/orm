// SPDX-License-Identifier: MIT

package core

import (
	"errors"
	"testing"

	"github.com/issue9/assert/v2"
)

func TestSQLBuilder(t *testing.T) {
	a := assert.New(t, false)

	b := NewBuilder()
	b.WBytes('1')
	b.WString("23")

	str, err := b.String()
	a.NotError(err).Equal("123", str)
	a.Equal(3, b.Len())

	b.Reset()
	str, err = b.String()
	a.NotError(err).Equal(str, "")
	a.Equal(b.Len(), 0)

	b.WBytes('B', 'y', 't', 'e').
		WRunes('R', 'u', 'n', 'e').
		WString("String")
	str, err = b.String()
	a.NotError(err).Equal(str, "ByteRuneString")

	b.WBytes('1', '2')
	b.TruncateLast(2)
	str, err = b.String()
	a.NotError(err).Equal(str, "ByteRuneString").Equal(14, b.Len())

	b.Reset()
	b.QuoteKey("key")
	bs, err := b.Bytes()
	a.NotError(err).Equal(bs, []byte("{key}"))

	buf := NewBuilder("buf-")
	buf.Append(b)
	bs, err = buf.Bytes()
	a.NotError(err).Equal(bs, "buf-{key}")

	// 带错误信息
	buf = NewBuilder()
	buf.buffer.Err = errors.New("test")
	buf.WString("str1")
	str, err = buf.String()
	a.ErrorString(err, "test").
		Empty(str)
}
