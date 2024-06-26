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

var _ sqlbuilder.DDLStmt = &sqlbuilder.CreateTableStmt{}

func TestCreateView(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		testCreateView(t)
	})
}

func testCreateView(d *test.Driver) {
	initDB(d)
	defer clearDB(d)
	sb := d.DB.SQLBuilder()

	viewName := "user_view"

	sel := sb.Select().
		Column("u.id as uid").
		Column("u.name").
		Column("i.address").
		Join("LEFT", "info", "i", "u.id=i.uid").
		From("users", "u")
	view := sel.View(viewName).
		Column("uid").
		Column("name", "address")
	d.NotError(view.Exec())

	exists, err := sqlbuilder.ViewExists(d.DB).View(viewName).Exists()
	d.NotError(err).True(exists)
	exists, err = sqlbuilder.TableExists(d.DB).Table("not-exists").Exists()
	d.NotError(err).False(exists)

	// 删除
	defer func() {
		dropView := sb.DropView().Name(viewName)
		d.NotError(dropView.Exec())
	}()

	// 创建同名视图
	view.Reset().Name(viewName).From(sel)
	d.Error(view.Exec(), "not err @%s", d.DriverName)

	// 以 replace 的方式创建
	view.Reset().Name(viewName).From(sel).Replace()
	d.NotError(view.Exec())
}
