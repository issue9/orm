// SPDX-License-Identifier: MIT

package sqltest

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v3/internal/flagtest"
)

func TestMain(m *testing.M) {
	flagtest.Main(m)
}

func TestEqual(t *testing.T) {
	a := assert.New(t)
	Equal(a, "insert   INTO tb2 (c1, c2) values (?, ?) , (? ,@c2)", "insert into tb2 (c1,c2) values (?,?),(?,@c2)")
}
