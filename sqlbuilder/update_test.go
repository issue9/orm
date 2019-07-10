// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/test"
	"github.com/issue9/orm/v2/sqlbuilder"
)

var (
	_ sqlbuilder.SQLer       = &sqlbuilder.UpdateStmt{}
	_ sqlbuilder.WhereStmter = &sqlbuilder.UpdateStmt{}
)

func TestUpdate_columnsHasDup(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		db := t.DB.DB
		d := t.DB.Dialect()

		u := sqlbuilder.Update(db, d).
			Table("users").
			Set("c1", "v1").
			Set("c1", "v1")
		_, err := u.Exec()
		a.ErrorType(err, sqlbuilder.ErrDupColumn)
	})
}

func TestUpdate(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		initDB(t)
		defer clearDB(t)

		db := t.DB.DB
		dialect := t.DB.Dialect()

		u := sqlbuilder.Update(db, dialect).Table("users")
		t.NotNil(u)

		u.Set("name", "name222").Where("id=?", 2)
		_, err := u.Exec()
		t.NotError(err)

		sel := sqlbuilder.Select(db, dialect).
			Select("name").
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
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		initDB(t)
		defer clearDB(t)

		db := t.DB.DB
		dialect := t.DB.Dialect()

		u := sqlbuilder.Update(db, dialect).
			Table("users").
			Increase("age", 5).
			Where("id=?", 1)
		t.NotNil(u)
		t.NotError(u.Exec())

		sel := sqlbuilder.Select(db, dialect).
			Select("age").
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
		t.NotError(u.Exec())
		sel.Reset()
		sel.Select("age").
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
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		initDB(t)
		defer clearDB(t)
		db := t.DB.DB
		dialect := t.DB.Dialect()

		u := sqlbuilder.Update(db, dialect).
			Table("users").
			Set("age", 100).
			Where("id=?", 1).
			OCC("version", 0)
		r, err := u.Exec()
		a.NotError(err).NotNil(r)

		sel := sqlbuilder.Select(db, dialect).
			Select("age").
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
		sel.Select("age").
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
