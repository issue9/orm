// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package sqltest_test

import (
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/orm/v5/internal/sqltest"
	"github.com/issue9/orm/v5/internal/test"
)

func TestMain(m *testing.M) {
	test.Main(m)
}

func TestEqual(t *testing.T) {
	a := assert.New(t, false)
	sqltest.Equal(a, "insert   INTO tb2 (c1, c2) values (?, ?) , (? ,@c2)", "insert into tb2 (c1,c2) values (?,?),(?,@c2)")
}
