// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/internal/sqltest"
)

var _ SQL = &DeleteStmt{}

func TestDelete(t *testing.T) {
	a := assert.New(t)

	d := Delete("#table").
		Where(true, "id=?", 1).
		Or("id=?", 2).
		And("id=?", 3)
	query, args, err := d.SQL()
	a.NotError(err)
	a.Equal(args, []interface{}{1, 2, 3})
	sqltest.Equal(a, query, "delete from #table where id=? or id=? and id=?")

	d.Reset()
	a.Empty(d.table)
	query, args, err = d.Table("tb1").Where(true, "id=?").Or("id=?", 1).SQL()
	a.Equal(err, ErrArgsNotMatch) // 由 where 抛出
	a.Empty(query).Nil(args)
}
