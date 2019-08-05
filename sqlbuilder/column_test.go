// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v3/core"
	"github.com/issue9/orm/v3/internal/test"
	"github.com/issue9/orm/v3/sqlbuilder"
)

var (
	_ sqlbuilder.DDLSQLer = &sqlbuilder.AddColumnStmt{}
	_ sqlbuilder.DDLSQLer = &sqlbuilder.DropColumnStmt{}
)

func TestColumn(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		db := t.DB

		err := sqlbuilder.CreateTable(db).
			Table("users").
			AutoIncrement("id", core.Int64Type).
			Exec()
		a.NotError(err)
		defer func() {
			err = sqlbuilder.DropTable(db).Table("users").Exec()
			a.NotError(err)
		}()

		addStmt := sqlbuilder.AddColumn(db)
		err = addStmt.Table("users").
			Column("col1", core.IntType, false, true, false, nil).
			Exec()
		a.NotError(err, "%s@%s", err, t.DriverName)

		dropStmt := sqlbuilder.DropColumn(db)
		err = dropStmt.Table("users").
			Column("col1").
			Exec()
		t.NotError(err, "%s@%s", err, t.DriverName)

		err = addStmt.Reset().Exec()
		a.ErrorType(err, sqlbuilder.ErrTableIsEmpty)

		err = addStmt.Reset().Table("users").Exec()
		a.ErrorType(err, sqlbuilder.ErrColumnsIsEmpty)

		err = dropStmt.Reset().Exec()
		a.ErrorType(err, sqlbuilder.ErrTableIsEmpty)
	})

	// 添加主键
	suite.ForEach(func(t *test.Driver) {
		db := t.DB

		err := sqlbuilder.CreateTable(db).
			Table("users").
			AutoIncrement("id", core.Int64Type).
			Column("name", core.StringType, false, false, false, nil).
			Exec()
		a.NotError(err)
		defer func() {
			err = sqlbuilder.DropTable(db).Table("users").Exec()
			a.NotError(err)
		}()

		// 已存在
		addStmt := sqlbuilder.AddColumn(db)
		err = addStmt.Table("users").
			Column("id", core.IntType, false, true, false, nil).
			Exec()
		a.Error(err, "%s@%s", err, t.DriverName)

		dropStmt := sqlbuilder.DropColumn(db)
		err = dropStmt.Table("users").
			Column("id").
			Exec()
		t.NotError(err, "%s@%s", err, t.DriverName)

		err = addStmt.Reset().
			Table("users").
			Column("id", core.IntType, false, true, false, nil).
			Exec()
		a.NotError(err)
	})
}
