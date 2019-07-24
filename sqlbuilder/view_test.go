// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/test"
	"github.com/issue9/orm/v2/sqlbuilder"
)

func TestCreateView(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testCreateView(t)
	})
}

func testCreateView(d *test.Driver) {
	initDB(d)
	defer clearDB(d)

	viewName := "user_view"

	sel := sqlbuilder.Select(d.DB).
		Column("u.id as uid").
		Column("u.name").
		Column("i.address").
		Join("LEFT", "info", "i", "u.id=i.uid").
		From("users", "u")
	view := sel.View(viewName).
		Column("uid").
		Column("name", "address")
	d.NotError(view.Exec())

	// 创建同名视图
	view.Reset().Name(viewName).From(sel)
	d.Error(view.Exec(), "not err @%s", d.DriverName)

	// 以 replace 的方式创建
	view.Reset().Name(viewName).From(sel).Replace()
	d.NotError(view.Exec())

	// 删除
	dropView := sqlbuilder.DropView(d.DB).Name(viewName)
	d.NotError(dropView.Exec())
}
