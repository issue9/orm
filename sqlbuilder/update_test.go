// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/orm/v6/internal/test"
	"github.com/issue9/orm/v6/sqlbuilder"
)

var _ sqlbuilder.ExecStmt = &sqlbuilder.UpdateStmt{}

func TestUpdate_columnsHasDup(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		u := sqlbuilder.Update(t.DB).
			Table("users").
			Set("c1", "v1").
			Set("c1", "v2")
		_, err := u.Exec()
		a.ErrorString(err, "存在重复的列名")
	})
}

func TestUpdate(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		u := sqlbuilder.Update(t.DB).Table("users")
		t.NotNil(u)

		u.Set("name", "name222").Where("id=?", 2)
		_, err := u.Exec()
		t.NotError(err)

		sel := sqlbuilder.Select(t.DB).
			Column("name").
			From("users").
			Where("id=?", 2)
		rows, err := sel.Query()
		t.NotError(err).NotNil(rows)
		t.True(rows.Next())
		var name string
		t.NotError(rows.Scan(&name))
		t.NotError(rows.Close())
		t.Equal(name, "name222")
	})
}

func TestUpdateStmt_Increase(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		u := sqlbuilder.Update(t.DB).
			Table("users").
			Increase("age", 5).
			Where("id=?", 1)
		t.NotNil(u)
		_, err := u.Exec()
		t.NotError(err)

		sel := sqlbuilder.Select(t.DB).
			Column("age").
			From("users").
			Where("id=?", 1)
		rows, err := sel.Query()
		t.NotError(err).NotNil(rows)
		t.True(rows.Next())
		var val int
		t.NotError(rows.Scan(&val))
		t.NotError(rows.Close())
		t.Equal(val, 6)

		// decrease
		u.Reset()
		u.Table("users").
			Decrease("age", 3).
			Where("id=?", 1)
		t.NotNil(u)
		_, err = u.Exec()
		t.NotError(err)
		sel.Reset().
			Column("age").
			From("users").
			Where("id=?", 1)
		rows, err = sel.Query()
		t.NotError(err).NotNil(rows)
		t.True(rows.Next())
		t.NotError(rows.Scan(&val))
		t.NotError(rows.Close())
		t.Equal(val, 3)
	})
}

func TestUpdateStmt_OCC(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		u := sqlbuilder.Update(t.DB).
			Table("users").
			Set("age", 100).
			Where("id=?", 1).
			OCC("version", 0)
		r, err := u.Exec()
		a.NotError(err).NotNil(r)

		sel := sqlbuilder.Select(t.DB).
			Column("age").
			From("users").
			Where("id=?", 1)
		rows, err := sel.Query()
		t.NotError(err).NotNil(rows)
		t.True(rows.Next())
		var val int
		t.NotError(rows.Scan(&val))
		t.NotError(rows.Close())
		t.Equal(val, 100)

		// 乐观锁判断失败主
		u.Reset()
		u.Table("users").
			Set("age", 111).
			Where("id=?", 1).
			OCC("version", 0)
		r, err = u.Exec()
		a.NotError(err).NotNil(r)

		sel.Reset()
		sel.Column("age").
			From("users").
			Where("id=?", 1)
		rows, err = sel.Query()
		t.NotError(err).NotNil(rows)
		t.True(rows.Next())
		t.NotError(rows.Scan(&val))
		t.NotError(rows.Close())
		t.Equal(val, 100)
	})
}
