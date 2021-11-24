// SPDX-License-Identifier: MIT

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/orm/v4/core"
	"github.com/issue9/orm/v4/internal/test"
	"github.com/issue9/orm/v4/sqlbuilder"
)

var (
	_ sqlbuilder.DDLStmt = &sqlbuilder.CreateIndexStmt{}
	_ sqlbuilder.DDLStmt = &sqlbuilder.DropIndexStmt{}
)

func TestIndex(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		createStmt := sqlbuilder.CreateIndex(t.DB).
			Table("users").
			Name("index_key").
			Columns("id", "name")
		err := createStmt.Exec()
		t.NotError(err)

		// 同名约束名，应该会出错
		createStmt.Reset()
		err = createStmt.Table("users").
			Name("index_key").
			Columns("id", "name").
			Exec()
		t.Error(err)

		// 唯一约束
		createStmt.Reset()
		err = createStmt.Table("users").
			Name("index_unique_key").
			Type(core.IndexUnique).
			Columns("id", "name").
			Exec()
		t.NotError(err)

		dropStmt := sqlbuilder.DropIndex(t.DB).
			Table("users").
			Name("index_key")
		err = dropStmt.Exec()
		t.NotError(err)

		// 不存在的索引
		dropStmt.Reset()
		err = dropStmt.Table("users").
			Name("index_key").
			Exec()
		a.Error(err)

		dropStmt.Reset()
		err = dropStmt.Table("users").
			Name("index_unique_key").
			Exec()
		t.NotError(err, "cc")

		createStmt.Reset()
		a.ErrorIs(createStmt.Exec(), sqlbuilder.ErrTableIsEmpty)

		createStmt.Reset()
		createStmt.Table("test")
		a.ErrorIs(createStmt.Exec(), sqlbuilder.ErrColumnsIsEmpty)

		dropStmt.Reset()
		dropStmt.Table("test")
		a.ErrorIs(dropStmt.Exec(), sqlbuilder.ErrColumnsIsEmpty)
	})
}
