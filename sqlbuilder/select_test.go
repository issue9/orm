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
	_ sqlbuilder.SQLer       = &sqlbuilder.SelectStmt{}
	_ sqlbuilder.WhereStmter = &sqlbuilder.SelectStmt{}
)

func TestSelect_Query(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		initDB(t)
		defer clearDB(t)
		db := t.DB.DB
		d := t.DB.Dialect()

		stmt := sqlbuilder.Select(db, d).Select("*").
			From("users").
			Where("id<?", 5).
			Desc("id")

		id, err := stmt.QueryInt("id")
		a.NotError(err).
			Equal(id, 4)

		name, err := stmt.QueryString("name")
		a.NotError(err).
			Equal(name, "4")

		obj := &user{}
		size, err := stmt.QueryObject(true, obj)
		a.NotError(err).Equal(1, size)
		a.Equal(obj.ID, 4)

		cnt, err := stmt.Count("count(*) AS cnt").QueryInt("cnt")
		a.NotError(err).
			Equal(cnt, 4)
	})
}
