// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/core"
	"github.com/issue9/orm/v2/internal/test"
	"github.com/issue9/orm/v2/sqlbuilder"
)

var (
	_ sqlbuilder.DDLSQLer = &sqlbuilder.CreateIndexStmt{}
	_ sqlbuilder.DDLSQLer = &sqlbuilder.DropIndexStmt{}
)

func TestIndex(t *testing.T) {
	a := assert.New(t)
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
		a.ErrorType(createStmt.Exec(), sqlbuilder.ErrTableIsEmpty)

		createStmt.Reset()
		createStmt.Table("test")
		a.ErrorType(createStmt.Exec(), sqlbuilder.ErrColumnsIsEmpty)

		dropStmt.Reset()
		a.ErrorType(dropStmt.Exec(), sqlbuilder.ErrTableIsEmpty)

		dropStmt.Reset()
		dropStmt.Table("test")
		a.ErrorType(dropStmt.Exec(), sqlbuilder.ErrTableIsEmpty)
	})
}
