// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/internal/sqltest"
)

var _ SQLer = &WhereStmt{}

func TestWhere(t *testing.T) {
	a := assert.New(t)
	w := newWhereStmt()
	a.NotNil(w)

	w.And("id=?", 1)
	w.And("name like ?", "name")
	w.Or("type=?", 5)
	sql, args, err := w.SQL()
	a.NotError(err).NotNil(args).NotEmpty(sql)
	a.Equal(args, []interface{}{1, "name", 5})
	sqltest.Equal(a, sql, "id=? and name like ? or type=?")

	w.Reset()
	a.Equal(0, w.buffer.Len())
	a.Equal(0, len(w.args))

	w.And("id=?", 5)
	sql, args, err = w.SQL()
	a.NotError(err).NotNil(args).NotEmpty(sql)
	a.Equal(args, []interface{}{5})
	sqltest.Equal(a, sql, "id=?")

	w.Reset()
	a.Equal(0, w.buffer.Len())
	a.Equal(0, len(w.args))

	w.And("id=?", 5, 7)
	sql, args, err = w.SQL()
	a.Equal(err, ErrArgsNotMatch).Nil(args).Empty(sql)
}

func TestWhere_addWhere(t *testing.T) {
	a := assert.New(t)
	w := newWhereStmt()

	w2 := newWhereStmt().And("id=?", 4)
	w1 := newWhereStmt().And("id=?", 2).Or("id=?", 3).OrWhere(w2)

	w.And("id=?", 1).
		AndWhere(w1).
		And("id=?", 5)

	query, args, err := w.SQL()
	a.NotError(err)
	a.Equal(args, []interface{}{1, 2, 3, 4, 5})
	sqltest.Equal(a, query, "id=? AND(id=? OR id=? OR(id=?)) and id=?")
}
