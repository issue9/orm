// SPDX-License-Identifier: MIT

package core

import (
	"strings"
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/orm/v4/internal/flagtest"
)

func TestMain(m *testing.M) {
	flagtest.Main(m)
}

func TestPKName(t *testing.T) {
	a := assert.New(t, false)
	a.True(strings.HasSuffix(PKName("xx"), PKNameSuffix))
}
