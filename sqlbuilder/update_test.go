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
		a.NotError(err)

		sel := sqlbuilder.Select(db, dialect).
			Select("name").
			From("users").
			Where("id=?", 2)
		rows, err := sel.Query()
		t.NotError(err).NotNil(rows)
		a.True(rows.Next())
		var name string
		a.NotError(rows.Scan(&name))
		a.NotError(rows.Close())
		a.Equal(name, "name222")
	})
}

// TODO increment 测试

// TODO occ
