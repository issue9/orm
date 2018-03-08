// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"fmt"
	"testing"

	"github.com/issue9/assert"
)

func TestConType_String(t *testing.T) {
	a := assert.New(t)

	a.Equal("<none>", none.String()).
		Equal("KEY INDEX", fmt.Sprint(index)).
		Equal("UNIQUE INDEX", unique.String()).
		Equal("FOREIGN KEY", fk.String()).
		Equal("CHECK", check.String())

	var c1 conType
	a.Equal("<none>", c1.String())

	c1 = 100
	a.Equal("<unknown>", c1.String())
}
