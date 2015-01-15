// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package builder

import (
	"testing"

	"github.com/issue9/assert"
)

func TestSQL_Where(t *testing.T) {
	a := assert.New(t)
	sql := NewSQL(nil)

	sql.Columns("c1").
		Table("user").
		Where("id=@id").
		And("group=1").
		Or("name like @name")

	wont := "select c1 from user where(id=@id) and(group=1) or(name like @name)"
	chkSQLEqual(a, wont, sql.selectSQL())

	// In
	sql.Reset().
		Columns("c1").
		Table("user").
		In("c1", 1, 2, "3")

	wont = "select c1 from user where(c1 in(1,2,'3'))"
	chkSQLEqual(a, wont, sql.selectSQL())

	// between
	sql.Reset().
		Columns("c1").
		Table("user").
		Between("c1", 5, 7)

	wont = "SELECT c1 from user where (c1 between 5 and 7)"
	chkSQLEqual(a, wont, sql.selectSQL())

	// is null
	sql.Reset().
		Columns("*").
		Table("user").
		IsNotNull("c1").
		OrIsNull("c2")

	wont = "SELECT * FROM user where(c1 is not null) or (c2 is null)"
	chkSQLEqual(a, wont, sql.selectSQL())
}
