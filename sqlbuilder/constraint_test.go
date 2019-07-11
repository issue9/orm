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

func TestPKName(t *testing.T) {
	a := assert.New(t)
	a.Equal("xx_pk", sqlbuilder.PKName("xx"))
}

func TestAIName(t *testing.T) {
	a := assert.New(t)
	a.Equal("xx_ai", sqlbuilder.AIName("xx"))
}

func TestConstraint(t *testing.T) {
	a := assert.New(t)

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		initDB(t)
		defer clearDB(t)

		db := t.DB.DB
		dialect := t.DB.Dialect()

		addStmt := sqlbuilder.AddConstraint(db, dialect)
		err := addStmt.Table("users").
			Unique("u_user_name", "name").
			Exec()
		t.NotError(err, "%s@%s", err, t.DriverName)

		// 删除约束
		dropStmt := sqlbuilder.DropConstraint(db, dialect).
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

func TestAddConstraintStmt_Check(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		initDB(t)
		defer clearDB(t)

		db := t.DB.DB
		dialect := t.DB.Dialect()

		err := sqlbuilder.AddConstraint(db, dialect).
			Table("info").
			Check("nick_not_null", "nickname IS NOT NULL").
			Exec()
		t.NotError(err)

		err = sqlbuilder.DropConstraint(db, dialect).
			Table("info").
			Constraint("nick_not_null").
			Exec()
		a.NotError(err)
	})
}

func TestAddConstraintStmt_PK(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		initDB(t)
		defer clearDB(t)

		db := t.DB.DB
		dialect := t.DB.Dialect()

		// 已经存在主键，出错
		addStmt := sqlbuilder.AddConstraint(db, dialect)
		err := addStmt.Table("info").
			PK("tel").
			Exec()
		t.Error(err)

		err = sqlbuilder.DropConstraint(db, dialect).
			Table("info").
			Constraint(sqlbuilder.PKName("info")).
			Exec()
		a.NotError(err)

		err = addStmt.Reset().Table("info").
			PK("tel", "nickname").
			Exec()
		t.NotError(err)
	})
}

func TestAddConstraintStmt_FK(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		initDB(t)
		defer clearDB(t)

		db := t.DB.DB
		dialect := t.DB.Dialect()

		// 已经存在主键，出错
		addStmt := sqlbuilder.AddConstraint(db, dialect)
		err := addStmt.Table("info").
			FK("info_fk", "uid", "users", "id", "CASCADE", "CASCADE").
			Exec()
		t.Error(err)

		err = sqlbuilder.DropConstraint(db, dialect).
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

func TestConstraint_String(t *testing.T) {
	a := assert.New(t)

	a.NotEqual(sqlbuilder.ConstraintAI.String(), "<unknown>")
	a.NotEqual(sqlbuilder.ConstraintFK.String(), "<unknown>")
	a.NotEqual(sqlbuilder.ConstraintPK.String(), "<unknown>")
	a.NotEqual(sqlbuilder.ConstraintUnique.String(), "<unknown>")
	a.NotEqual(sqlbuilder.ConstraintCheck.String(), "<unknown>")
	a.Equal(sqlbuilder.Constraint(-1).String(), "<unknown>")
	a.Equal(sqlbuilder.Constraint(100).String(), "<unknown>")
}
