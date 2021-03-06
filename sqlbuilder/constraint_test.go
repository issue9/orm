// SPDX-License-Identifier: MIT

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v3/core"
	"github.com/issue9/orm/v3/internal/test"
	"github.com/issue9/orm/v3/sqlbuilder"
)

func TestConstraint(t *testing.T) {
	a := assert.New(t)

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
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
		a.ErrorType(err, sqlbuilder.ErrTableIsEmpty)

		err = dropStmt.Reset().Table("tbl").Exec()
		a.ErrorType(err, sqlbuilder.ErrConstraintIsEmpty)

		a.ErrorType(addStmt.Reset().Unique("", "name").Exec(), sqlbuilder.ErrTableIsEmpty)
		a.ErrorType(addStmt.Reset().Table("users").Unique("", "name").Exec(), sqlbuilder.ErrConstraintIsEmpty)
	})
}

func TestConstraint_Check(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		err := sqlbuilder.AddConstraint(t.DB).
			Table("info").
			Check("nick_not_null", "nickname IS NOT NULL").
			Exec()
		t.NotError(err)

		err = sqlbuilder.DropConstraint(t.DB).
			Table("info").
			Constraint("nick_not_null").
			Exec()
		a.NotError(err)
	})
}

func TestConstraint_PK(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		// 已经存在主键，出错
		addStmt := sqlbuilder.AddConstraint(t.DB)
		err := addStmt.Table("info").
			PK("tel").
			Exec()
		t.Error(err)

		err = sqlbuilder.DropConstraint(t.DB).
			Table("info").
			Constraint(core.PKName("info")).
			Exec()
		a.NotError(err)

		err = addStmt.Reset().Table("info").
			PK("tel", "nickname").
			Exec()
		t.NotError(err)
	})

	// 表名带 #
	suite.ForEach(func(t *test.Driver) {
		err := sqlbuilder.CreateTable(t.DB).
			Table("#info").
			Column("uid", core.Int64, false, false, false, nil).
			Column("tel", core.String, false, false, false, nil, 11).
			PK("tel", "uid").
			Exec()
		t.NotError(err)

		defer func() {
			err := sqlbuilder.DropTable(t.DB).Table("#info").Exec()
			t.NotError(err)
		}()

		// 已经存在主键，出错
		addStmt := sqlbuilder.AddConstraint(t.DB)
		err = addStmt.Table("#info").
			PK("tel").
			Exec()
		t.Error(err)

		err = sqlbuilder.DropConstraint(t.DB).
			Table("#info").
			Constraint(core.PKName("#info")).
			Exec()
		a.NotError(err)

		err = addStmt.Reset().Table("#info").
			PK("tel", "uid").
			Exec()
		t.NotError(err)
	})

	// 约束名不是根据 core.PKName() 生成的
	suite.ForEach(func(t *test.Driver) {
		query := "CREATE TABLE #info (uid BIGINT NOT NULL,CONSTRAINT test_pk PRIMARY KEY(uid))"
		_, err := t.DB.Exec(query)
		t.NotError(err)

		defer func() {
			err := sqlbuilder.DropTable(t.DB).Table("#info").Exec()
			t.NotError(err)
		}()

		// 已经存在主键，出错
		addStmt := sqlbuilder.AddConstraint(t.DB)
		err = addStmt.Table("#info").
			PK("uid").
			Exec()
		t.Error(err)

		err = sqlbuilder.DropConstraint(t.DB).
			Table("#info").
			Constraint("test_pk").
			PK().
			Exec()
		a.NotError(err)

		err = addStmt.Reset().Table("#info").
			PK("uid").
			Exec()
		t.NotError(err)
	})
}

func TestConstraint_FK(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
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
