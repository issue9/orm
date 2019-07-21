// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package core

import (
	"testing"

	"github.com/issue9/assert"
)

func TestSQLBuilder(t *testing.T) {
	a := assert.New(t)

	b := NewBuilder("")
	b.WriteBytes('1')
	b.WriteString("23")

	a.Equal("123", b.String())
	a.Equal(3, b.Len())

	b.Reset()
	a.Equal(b.String(), "")
	a.Equal(b.Len(), 0)

	b.WriteBytes('3').WriteString("21")
	a.Equal(b.String(), "321")

	b.TruncateLast(1)
	a.Equal(b.String(), "32").Equal(2, b.Len())
}
