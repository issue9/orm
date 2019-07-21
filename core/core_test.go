// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package core

import (
	"testing"

	"github.com/issue9/assert"
)

func TestPKName(t *testing.T) {
	a := assert.New(t)
	a.Equal("xx_pk", PKName("xx"))
}

func TestAIName(t *testing.T) {
	a := assert.New(t)
	a.Equal("xx_ai", AIName("xx"))
}
