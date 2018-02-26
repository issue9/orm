// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"testing"

	"github.com/issue9/assert"
)

func TestStringBuilder(t *testing.T) {
	a := assert.New(t)

	b := newStringBuilder("")
	b.writeByte('1')
	b.writeString("23")

	a.Equal("123", b.string())
	a.Equal(3, b.len())

	b.reset()
	a.Equal(b.string(), "")
	a.Equal(b.len(), 0)

	b.writeByte('3')
	b.writeString("21")
	a.Equal(b.string(), "321")

	b.truncateLast(1)
	a.Equal(b.string(), "32").Equal(2, b.len())
}
