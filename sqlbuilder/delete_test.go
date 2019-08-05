// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/sqltest"
	"github.com/issue9/orm/v2/internal/test"
	"github.com/issue9/orm/v2/sqlbuilder"
)

var (
	_ sqlbuilder.SQLer       = &sqlbuilder.DeleteStmt{}
	_ sqlbuilder.WhereStmter = &sqlbuilder.DeleteStmt{}
)

func TestDelete_Exec(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		sql := sqlbuilder.Delete(t.DB).
			Table("users").
			Where("id=?", 1)
		_, err := sql.Exec()
		a.NotError(err)

		sql.Reset()
		sql.Table("users").
			Where("id=?").
			Or("name=?", "xx")
		_, err = sql.Exec()
		a.ErrorType(err, sqlbuilder.ErrArgsNotMatch)

		sql.Reset()
		_, err = sql.Exec()
		a.ErrorType(err, sqlbuilder.ErrTableIsEmpty)
	})
}

func TestWhereStmt_Delete(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		sql := sqlbuilder.Where().And("id=?", 1).
			Delete(t.DB).
			Table("users")
		_, err := sql.Exec()
		a.NotError(err)

		query, args, err := sql.SQL()
		a.NotError(err).
			Equal(args, []interface{}{1})
		sqltest.Equal(a, query, "DELETE FROM {users} WHERE id=?")
	})
}
