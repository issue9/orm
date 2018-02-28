// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"testing"

	"github.com/issue9/assert"
)

var _ SQL = &where{}

func TestWhere(t *testing.T) {
	a := assert.New(t)
	w := newWhere()
	a.NotNil(w)

	w.and("id=?", 1)
	w.and("name like ?", "name")
	w.or("type=?", 5)
	sql, args, err := w.SQL()
	a.NotError(err).NotNil(args).NotEmpty(sql)
	a.Equal(args, []interface{}{1, "name", 5})
	chkSQLEqual(a, sql, "where id=? and name like ? or type=?")

	w.Reset()
	a.Equal(0, w.buffer.Len())
	a.Equal(0, len(w.args))

	w.and("id=?", 5)
	sql, args, err = w.SQL()
	a.NotError(err).NotNil(args).NotEmpty(sql)
	a.Equal(args, []interface{}{5})
	chkSQLEqual(a, sql, "where id=?")

	w.Reset()
	a.Equal(0, w.buffer.Len())
	a.Equal(0, len(w.args))

	w.and("id=?", 5, 7)
	sql, args, err = w.SQL()
	a.Equal(err, ErrArgsNotMatch).Nil(args).Empty(sql)
}
