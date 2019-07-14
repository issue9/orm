// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/sqltest"
)

var _ SQLer = &WhereStmt{}

func TestWhere(t *testing.T) {
	a := assert.New(t)
	w := newWhere('{', '}')
	a.NotNil(w)

	w.And("id=?", 1)
	w.And("name like ?", "name")
	w.Or("type=?", 5)
	sql, args, err := w.SQL()
	a.NotError(err).NotNil(args).NotEmpty(sql)
	a.Equal(args, []interface{}{1, "name", 5})
	sqltest.Equal(a, sql, "id=? and name like ? or type=?")

	w.Reset()
	a.Equal(0, w.builder.Len())
	a.Equal(0, len(w.args))

	w.And("id=?", 5)
	sql, args, err = w.SQL()
	a.NotError(err).NotNil(args).NotEmpty(sql)
	a.Equal(args, []interface{}{5})
	sqltest.Equal(a, sql, "id=?")

	w.Reset()
	a.Equal(0, w.builder.Len())
	a.Equal(0, len(w.args))

	w.And("id=?", 5, 7)
	sql, args, err = w.SQL()
	a.Equal(err, ErrArgsNotMatch).Nil(args).Empty(sql)
}

func TestWhere_Group(t *testing.T) {
	a := assert.New(t)
	w := newWhere('{', '}')

	w.AndGroup().And("id=?", 4)
	w.AndGroup().And("id=?", 2).
		OrGroup().And("id=?", 3).
		EndGroup().
		And("id=?", 6)

	w.And("id=?", 1).And("id=?", 5)

	query, args, err := w.SQL()
	a.NotError(err)
	a.Equal(args, []interface{}{1, 5, 4, 2, 6, 3})
	sqltest.Equal(a, query, "id=? AND id=? and(id=?) AND (id=? and id=? OR(id=?))")

	// Reset
	w.Reset()
	w.And("id=?", 1)
	query, args, err = w.SQL()
	a.NotError(err)
	a.Equal(args, []interface{}{1})
	sqltest.Equal(a, query, "id=?")
}
