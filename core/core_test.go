// SPDX-License-Identifier: MIT

package core

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v3/internal/flagtest"
)

func TestMain(m *testing.M) {
	flagtest.Main(m)
}

func TestPKName(t *testing.T) {
	a := assert.New(t)
	a.Equal("xx_pk", PKName("xx"))
}

func TestAIName(t *testing.T) {
	a := assert.New(t)
	a.Equal("xx_ai", AIName("xx"))
}
