// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/orm/v5/internal/test"
	"github.com/issue9/orm/v5/sqlbuilder"
)

var (
	_ sqlbuilder.DDLStmt = &sqlbuilder.DropConstraintStmt{}
	_ sqlbuilder.DDLStmt = &sqlbuilder.AddConstraintStmt{}
)

func TestConstraint(t *testing.T) {
	a := assert.New(t, false)

	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		addStmt := sqlbuilder.AddConstraint(t.DB)
		err := addStmt.Table("users").
			Unique("u_user_name", "name").
			Exec()
		t.NotError(err, "%s@%s", err, t.DriverName)

		// 删除约束
		dropStmt := sqlbuilder.DropConstraint(t.DB).
			Table("users").
			Constraint("u_user_name")
		err = dropStmt.Exec()
		a.NotError(err, "%s@%s", err, t.DriverName)

		// 删除不存在的约束名
		err = dropStmt.Reset().
			Table("users").
			Constraint("u_user_name_not_exists___").
			Exec()
		a.Error(err, "并未出错 @%s", t.DriverName)

		err = dropStmt.Reset().Exec()
		a.ErrorIs(err, sqlbuilder.ErrTableIsEmpty)

		err = dropStmt.Reset().Table("tbl").Exec()
		a.ErrorIs(err, sqlbuilder.ErrColumnsIsEmpty)

		a.ErrorIs(addStmt.Reset().Unique("", "name").Exec(), sqlbuilder.ErrTableIsEmpty)
		a.ErrorIs(addStmt.Reset().Table("users").Unique("", "name").Exec(), sqlbuilder.ErrConstraintIsEmpty)
	})
}

func TestConstraint_Check(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)
		sb := t.DB.SQLBuilder()

		err := sb.AddConstraint().
			Table("info").
			Check("nick_not_null", "nickname IS NOT NULL").
			Exec()
		t.NotError(err)

		err = sb.DropConstraint().
			Table("info").
			Constraint("nick_not_null").
			Exec()
		a.NotError(err)
	})
}

func TestConstraint_PK(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)
		sb := t.DB.SQLBuilder()

		// 已经存在主键，出错
		addStmt := sb.AddConstraint()
		err := addStmt.Table("info").
			PK("info_pk", "tel").
			Exec()
		t.Error(err)

		err = sb.DropConstraint().
			Table("info").
			PK("info_pk").
			Exec()
		a.NotError(err)

		err = addStmt.Reset().Table("info").
			PK("info_pk", "tel", "nickname").
			Exec()
		t.NotError(err)
	})

	// 约束名不是根据 core.pkName 生成的
	suite.Run(func(t *test.Driver) {
		query := "CREATE TABLE info (uid BIGINT NOT NULL,CONSTRAINT test_pk PRIMARY KEY(uid))"
		_, err := t.DB.Exec(query)
		t.NotError(err)

		defer func() {
			err := sqlbuilder.DropTable(t.DB).Table("info").Exec()
			t.NotError(err)
		}()

		// 已经存在主键，出错
		addStmt := sqlbuilder.AddConstraint(t.DB)
		err = addStmt.Table("info").
			PK("info_pk", "uid").
			Exec()
		t.Error(err)

		err = sqlbuilder.DropConstraint(t.DB).
			Table("info").
			PK("test_pk").
			Exec()
		a.NotError(err)

		err = addStmt.Reset().Table("info").
			PK("info_pk", "uid").
			Exec()
		t.NotError(err)
	})
}

func TestConstraint_FK(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		// 已经存在主键，出错
		addStmt := sqlbuilder.AddConstraint(t.DB)
		err := addStmt.Table("info").
			FK("info_fk", "uid", "users", "id", "CASCADE", "CASCADE").
			Exec()
		t.Error(err)

		err = sqlbuilder.DropConstraint(t.DB).
			Table("info").
			Constraint("info_fk").
			Exec()
		a.NotError(err)

		err = addStmt.Reset().Table("info").
			FK("info_fk", "uid", "users", "id", "CASCADE", "CASCADE").
			Exec()
		t.NotError(err)
	})
}
