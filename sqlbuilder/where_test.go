// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package sqlbuilder

import (
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/orm/v6/internal/sqltest"
)

var _ SQLer = &WhereStmt{}

func TestWhere(t *testing.T) {
	a := assert.New(t, false)
	w := Where()
	a.NotNil(w)

	w.And("id=?", 1)
	w.And("name like ?", "name")
	w.Or("type=?", 5)
	sql, args, err := w.SQL()
	a.NotError(err).NotNil(args).NotEmpty(sql)
	a.Equal(args, []any{1, "name", 5})
	sqltest.Equal(a, sql, "id=? and name like ? or type=?")

	w.Reset()
	a.Equal(0, w.builder.Len())
	a.Equal(0, len(w.args))

	w.And("id=?", 5)
	sql, args, err = w.SQL()
	a.NotError(err).NotNil(args).NotEmpty(sql)
	a.Equal(args, []any{5})
	sqltest.Equal(a, sql, "id=?")

	w.Reset()
	a.Equal(0, w.builder.Len())
	a.Equal(0, len(w.args))

	w.And("id=?", 5, 7)
	sql, args, err = w.SQL()
	a.ErrorString(err, "列与值不匹配").Nil(args).Empty(sql)

	// cond==""

	w.Reset()
	a.Equal(0, w.builder.Len())
	a.Equal(0, len(w.args))

	w.And("")
	sql, args, err = w.SQL()
	a.NotError(err).Empty(args)
	sqltest.Equal(a, sql, "")

	w.Reset()
	a.PanicString(func() {
		w.And("", 3)
	}, "列与值不匹配")
}

func TestWhereStmt_IsNull(t *testing.T) {
	a := assert.New(t, false)
	w := Where()

	w.AndIsNull("col1")
	query, args, err := w.SQL()
	a.NotError(err).Empty(args)
	sqltest.Equal(a, query, "{col1} is null")

	w.OrIsNull("col2")
	query, args, err = w.SQL()
	a.NotError(err).Empty(args)
	sqltest.Equal(a, query, "{col1} is null or {col2} is null")

	w.Reset()
	w.AndIsNotNull("col1")
	query, args, err = w.SQL()
	a.NotError(err).Empty(args)
	sqltest.Equal(a, query, "{col1} is not null")

	w.OrIsNotNull("col2")
	query, args, err = w.SQL()
	a.NotError(err).Empty(args)
	sqltest.Equal(a, query, "{col1} is not null or {col2} is not null")
}

func TestWhereStmt_Like(t *testing.T) {
	a := assert.New(t, false)
	w := Where()

	w.AndLike("col1", "%str1")
	query, args, err := w.SQL()
	a.NotError(err).
		Equal(args, []any{"%str1"})
	sqltest.Equal(a, query, "{col1} like ?")

	w.OrLike("col2", "str2%")
	query, args, err = w.SQL()
	a.NotError(err).
		Equal(args, []any{"%str1", "str2%"})
	sqltest.Equal(a, query, "{col1} like ?  or {col2} like ?")

	w.Reset()
	w.AndNotLike("col1", "%str1")
	query, args, err = w.SQL()
	a.NotError(err).
		Equal(args, []any{"%str1"})
	sqltest.Equal(a, query, "{col1} not like ?")

	w.OrNotLike("col2", "str2%")
	query, args, err = w.SQL()
	a.NotError(err).
		Equal(args, []any{"%str1", "str2%"})
	sqltest.Equal(a, query, "{col1} not like ? or {col2} not like ?")
}

func TestWhereStmt_Between(t *testing.T) {
	a := assert.New(t, false)
	w := Where()

	// AndBetween
	w.AndBetween("col1", 1, 2)
	query, args, err := w.SQL()
	a.NotError(err).
		Equal(args, []any{1, 2})
	sqltest.Equal(a, query, "{col1} between ? and ?")

	// OrBetween
	w.OrBetween("col2", 3, 4)
	query, args, err = w.SQL()
	a.NotError(err).
		Equal(args, []any{1, 2, 3, 4})
	sqltest.Equal(a, query, "{col1} between ? and ? or {col2} between ? and ?")

	// AndNotBetween
	w.Reset()
	w.AndNotBetween("col1", 1, 2)
	query, args, err = w.SQL()
	a.NotError(err).
		Equal(args, []any{1, 2})
	sqltest.Equal(a, query, "{col1} not between ? and ?")

	// OrBetween
	w.OrNotBetween("col2", 3, 4)
	query, args, err = w.SQL()
	a.NotError(err).
		Equal(args, []any{1, 2, 3, 4})
	sqltest.Equal(a, query, "{col1} not between ? and ? or {col2} not between ? and ?")
}

func TestWhereStmt_In(t *testing.T) {
	a := assert.New(t, false)
	w := Where()

	w.OrIn("col1", 1, 2, 3)
	query, args, err := w.SQL()
	a.NotError(err).
		Equal(args, []any{1, 2, 3})
	sqltest.Equal(a, query, "{col1} in(?,?,?)")

	w.AndIn("col2", "1", "2", "test")
	query, args, err = w.SQL()
	a.NotError(err).
		Equal(args, []any{1, 2, 3, "1", "2", "test"})
	sqltest.Equal(a, query, "{col1} in(?,?,?) and {col2} in(?,?,?)")

	w.Reset()
	w.OrNotIn("col1", 1, 2, 3)
	query, args, err = w.SQL()
	a.NotError(err).
		Equal(args, []any{1, 2, 3})
	sqltest.Equal(a, query, "{col1} not in(?,?,?)")

	w.AndNotIn("col2", "1", "2", "test")
	query, args, err = w.SQL()
	a.NotError(err).
		Equal(args, []any{1, 2, 3, "1", "2", "test"})
	sqltest.Equal(a, query, "{col1} not in(?,?,?) and {col2} not in(?,?,?)")
}

func TestWhereStmt_Group(t *testing.T) {
	a := assert.New(t, false)
	w := Where()

	w.AndGroup(func(ws *WhereStmt) {
		ws.And("id=?", 4)
	})

	w.AndGroup(func(ws *WhereStmt) {
		ws.And("id=?", 2).OrGroup(func(ws *WhereStmt) {
			ws.And("id=?", 3)
		}).And("id=?", 6)
	})

	w.And("id=?", 1).And("id=?", 5)

	query, args, err := w.SQL()
	a.NotError(err)
	a.Equal(args, []any{1, 5, 4, 2, 6, 3})
	sqltest.Equal(a, query, "id=? AND id=? and(id=?) AND (id=? and id=? OR(id=?))")

	// Reset
	w.Reset()
	w.And("id=?", 1)
	query, args, err = w.SQL()
	a.NotError(err)
	a.Equal(args, []any{1})
	sqltest.Equal(a, query, "id=?")
}

func TestWhereStmt_Cond(t *testing.T) {
	a := assert.New(t, false)
	w := Where()

	w.Cond(true, func(ww *WhereStmt) {
		ww.AndGroup(func(ws *WhereStmt) {
			ws.And("id=?", 4)
		})
	}).AndGroup(func(ws *WhereStmt) {
		ws.And("id=?", 2).OrGroup(func(ws *WhereStmt) {
			ws.And("id=?", 3)
		}).And("id=?", 6)
	})

	w.Cond(false, func(stmt *WhereStmt) {
		stmt.And("id=?", 1)
	}).And("id=?", 5)

	query, args, err := w.SQL()
	a.NotError(err)
	a.Equal(args, []any{5, 4, 2, 6, 3})
	sqltest.Equal(a, query, "id=? and(id=?) AND (id=? and id=? OR(id=?))")

	// Reset
	w.Reset()
	w.And("id=?", 1)
	query, args, err = w.SQL()
	a.NotError(err)
	a.Equal(args, []any{1})
	sqltest.Equal(a, query, "id=?")
}
