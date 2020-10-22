// SPDX-License-Identifier: MIT

package flagtest

import (
	"testing"

	"github.com/issue9/assert"
)

func TestMain(m *testing.M) {
	Main(m)
}

func TestFlags(t *testing.T) {
	a := assert.New(t)

	a.NotNil(Flags)
}
