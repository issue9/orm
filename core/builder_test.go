// SPDX-License-Identifier: MIT

package core

import (
	"errors"
	"testing"

	"github.com/issue9/assert"
)

func TestSQLBuilder(t *testing.T) {
	a := assert.New(t)

	b := NewBuilder()
	b.WriteBytes('1')
	b.WriteString("23")

	str, err := b.String()
	a.NotError(err).Equal("123", str)
	a.Equal(3, b.Len())

	b.Reset()
	str, err = b.String()
	a.NotError(err).Equal(str, "")
	a.Equal(b.Len(), 0)

	b.WriteBytes('B', 'y', 't', 'e').
		WriteRunes('R', 'u', 'n', 'e').
		WriteString("String")
	str, err = b.String()
	a.NotError(err).Equal(str, "ByteRuneString")

	b.WriteBytes('1', '2')
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
	buf.err = errors.New("test")
	buf.WriteString("str1")
	str, err = buf.String()
	a.ErrorString(err, "test").
		Empty(str)
}
