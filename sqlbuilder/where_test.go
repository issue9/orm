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
	sqltest.Equal(a, sql, "where id=? and name like ? or type=?")

	w.Reset()
	a.Equal(0, w.buffer.Len())
	a.Equal(0, len(w.args))

	w.And("id=?", 5)
	sql, args, err = w.SQL()
	a.NotError(err).NotNil(args).NotEmpty(sql)
	a.Equal(args, []interface{}{5})
	sqltest.Equal(a, sql, "where id=?")

	w.Reset()
	a.Equal(0, w.buffer.Len())
	a.Equal(0, len(w.args))

	w.And("id=?", 5, 7)
	sql, args, err = w.SQL()
	a.Equal(err, ErrArgsNotMatch).Nil(args).Empty(sql)
}

func TestWhere_group(t *testing.T) {
	a := assert.New(t)
	w := newWhereStmt()

	w.And("id=?", 1).
		OrGroup("id=?", 2).
		And("id=?", 3).
		AndGroup("id=?", 4).
		EndGroup().
		And("id=?", 5).
		EndGroup().
		And("id=?", 6)

	query, args, err := w.SQL()
	a.NotError(err)
	a.Equal(args, []interface{}{1, 2, 3, 4, 5, 6})
	sqltest.Equal(a, query, "where id=? OR(id=? and id=? and(id=?) and id=?) and id=?")

	w.Reset()
	a.Panic(func() {
		w.EndGroup()
	})
}
