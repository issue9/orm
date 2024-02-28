// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package types

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/orm/v5/core"
)

type (
	ints = SliceOf[int64]
	strs = SliceOf[string]
)

var (
	_ sql.Scanner         = &ints{}
	_ driver.Valuer       = ints{}
	_ core.PrimitiveTyper = &ints{}

	_ sql.Scanner         = &strs{}
	_ driver.Valuer       = strs{}
	_ core.PrimitiveTyper = &strs{}
)

func TestSlices_Scan(t *testing.T) {
	a := assert.New(t, false)

	// ints

	u := &ints{}
	a.NotError(u.Scan("[1,2,3]")).
		Equal(u, &ints{1, 2, 3})

	// 无效的类型
	u = &ints{}
	a.Error(u.Scan(1))

	u = &ints{}
	a.Error(u.Scan("2020"))

	u = &ints{}
	a.Error(u.Scan(map[string]string{}))

	u = &ints{}
	a.NotError(u.Scan(nil))

	// strs

	s := &strs{}
	a.NotError(s.Scan(`["1","2","3\""]`)).
		Equal(s, &strs{"1", "2", "3\""})

	// 无效的类型
	s = &strs{}
	a.Error(s.Scan(1))

	s = &strs{}
	a.Error(s.Scan("2020"))

	s = &strs{}
	a.Error(s.Scan(map[string]string{}))

	s = &strs{}
	a.NotError(s.Scan(nil))
}
