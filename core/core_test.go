// SPDX-License-Identifier: MIT

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
