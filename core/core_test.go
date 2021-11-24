// SPDX-License-Identifier: MIT

package core

import (
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/orm/v4/internal/flagtest"
)

func TestMain(m *testing.M) {
	flagtest.Main(m)
}

func TestPKName(t *testing.T) {
	a := assert.New(t, false)
	a.Equal("xx_pk", PKName("xx"))
}

func TestAIName(t *testing.T) {
	a := assert.New(t, false)
	a.Equal("xx_ai", AIName("xx"))
}
