// SPDX-License-Identifier: MIT

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/orm/v5/internal/sqltest"
	"github.com/issue9/orm/v5/internal/test"
	"github.com/issue9/orm/v5/sqlbuilder"
)

var _ sqlbuilder.ExecStmt = &sqlbuilder.DeleteStmt{}

func TestDelete_Exec(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a)

	suite.Run(func(t *test.Driver) {
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
		a.ErrorIs(err, sqlbuilder.ErrArgsNotMatch)

		sql.Reset()
		_, err = sql.Exec()
		a.ErrorIs(err, sqlbuilder.ErrTableIsEmpty)
	})
}

func TestWhereStmt_Delete(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a)

	suite.Run(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		sql := sqlbuilder.Where().And("id=?", 1).
			Delete(t.DB).
			Table("users")
		_, err := sql.Exec()
		a.NotError(err)

		query, args, err := sql.SQL()
		a.NotError(err).
			Equal(args, []any{1})
		sqltest.Equal(a, query, "DELETE FROM {users} WHERE id=?")
	})
}
