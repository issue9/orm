// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"fmt"
	"testing"

	"github.com/issue9/assert"
)

func TestContType(t *testing.T) {
	a := assert.New(t)

	a.Equal("KEY INDEX", fmt.Sprint(Index)).
		Equal("UNIQUE INDEX", Unique.String()).
		Equal("FOREIGN KEY", Fk.String()).
		Equal("CHECK", Check.String())

	var c1 ConType
	a.Equal("KEY INDEX", c1.String())

	c1 = 100
	a.Equal("<unknown>", c1.String())
}
