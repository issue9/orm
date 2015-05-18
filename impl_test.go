// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"bytes"
	"testing"
	"time"

	"github.com/issue9/assert"
)

var _ engine = &DB{}
var _ engine = &Tx{}

func TestWriteString(t *testing.T) {
	a := assert.New(t)

	w := new(bytes.Buffer)
	a.NotNil(w)

	fn := func(v1 string, v2 interface{}) {
		w.Reset()
		WriteString(w, v2)
		a.Equal(v1, w.String())
	}

	fn("str", "str")         // string
	fn("str", []byte("str")) // []byte
	fn("str", []rune("str")) // []rune
	fn("1", 1)               // int
	fn("-1", -1)             // int
	fn("-1", -01)            // int
	fn("-10", -10)           // int
	fn("10", uint(10))       // uint
	fn("1", 1.0)             // float
	fn("-1", -1.0)           // float
	fn("1.1", 1.1)           // float
	fn("true", true)         // bool
	fn("false", false)       // bool
	//fn("2007-06-05 14:23:11", time.Date(2007, 06, 05, 14, 23, 11, 0, time.UTC)) // date
}
