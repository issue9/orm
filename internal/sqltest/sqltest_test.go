// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqltest

import (
	"testing"

	"github.com/issue9/assert"
)

func TestSQLTest(t *testing.T) {
	a := assert.New(t)
	Equal(a, "insert   INTO tb2 (c1, c2) values (?, ?) , (? ,@c2)", "insert into tb2 (c1,c2) values (?,?),(?,@c2)")
}
