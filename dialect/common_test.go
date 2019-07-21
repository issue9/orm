// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect_test

import (
	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/core"
	"github.com/issue9/orm/v2/internal/sqltest"
	"github.com/issue9/orm/v2/internal/test"
	"github.com/issue9/orm/v2/sqlbuilder"
)

type sqltypeTester struct {
	col     *core.Column
	err     bool
	SQLType string
}

func testSQLType(a *assert.Assertion, d core.Dialect, data []*sqltypeTester) {
	for _, item := range data {
		typ, err := d.SQLType(item.col)
		if item.err {
			a.Error(err)
		} else {
			a.NotError(err)
		}
		sqltest.Equal(a, typ, item.SQLType)
	}
}

func testDialectVersionSQL(t *test.Test) {
	rows, err := t.DB.Query(t.DB.Dialect().VersionSQL())
	t.NotError(err).NotNil(rows)
	defer func() {
		t.NotError(rows.Close())
	}()

	t.True(rows.Next())
	var ver string
	t.NotError(rows.Scan(&ver))
	t.NotEmpty(ver)
}

func testDialectDropConstraintStmtHook(t *test.Test) {
	db := t.DB

	// 不存在的约束，出错
	stmt := sqlbuilder.DropConstraint(db).
		Table("fk_table").
		Constraint("id_great_zero")

	t.Error(stmt.Exec())

	err := sqlbuilder.AddConstraint(db).
		Table("fk_table").
		Check("id_great_zero", "id>0").
		Exec()
	t.NotError(err)

	// 约束已经添加，可以正常删除
	// check
	stmt.Reset()
	err = stmt.Table("fk_table").Constraint("id_great_zero").Exec()
	t.NotError(err)

	// fk
	stmt.Reset()
	err = stmt.Table("usr").Constraint("xxx_fk").Exec()
	t.NotError(err)

	// unique
	stmt.Reset()
	err = stmt.Table("usr").Constraint("u_user_xx1").Exec()
	t.NotError(err)

	// pk
	stmt.Reset()
	err = stmt.Table("usr").Constraint(core.PKName("usr")).Exec()
	t.NotError(err)
}
