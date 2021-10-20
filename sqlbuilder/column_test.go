// SPDX-License-Identifier: MIT

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v4/core"
	"github.com/issue9/orm/v4/internal/test"
	"github.com/issue9/orm/v4/sqlbuilder"
)

var (
	_ sqlbuilder.DDLStmt = &sqlbuilder.AddColumnStmt{}
	_ sqlbuilder.DDLStmt = &sqlbuilder.DropColumnStmt{}
)

func TestColumn(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		db := t.DB

		err := sqlbuilder.CreateTable(db).
			Table("users").
			AutoIncrement("id", core.Int64).
			Exec()
		a.NotError(err)
		defer func() {
			err = sqlbuilder.DropTable(db).Table("users").Exec()
			a.NotError(err)
		}()

		addStmt := sqlbuilder.AddColumn(db)
		err = addStmt.Table("users").
			Column("col1", core.Int, false, true, false, nil).
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
			AutoIncrement("id", core.Int64).
			Column("name", core.String, false, false, false, nil).
			Exec()
		a.NotError(err)
		defer func() {
			err = sqlbuilder.DropTable(db).Table("users").Exec()
			a.NotError(err)
		}()

		// 已存在
		addStmt := sqlbuilder.AddColumn(db)
		err = addStmt.Table("users").
			Column("id", core.Int, false, true, false, nil).
			Exec()
		a.Error(err, "%s@%s", err, t.DriverName)

		dropStmt := sqlbuilder.DropColumn(db)
		err = dropStmt.Table("users").
			Column("id").
			Exec()
		t.NotError(err, "%s@%s", err, t.DriverName)

		err = addStmt.Reset().
			Table("users").
			Column("id", core.Int, false, true, false, nil).
			Exec()
		a.NotError(err)
	})
}
