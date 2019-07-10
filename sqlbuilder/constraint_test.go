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

func TestConstraint(t *testing.T) {
	a := assert.New(t)

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		initDB(t)
		defer clearDB(t)

		db := t.DB.DB
		dialect := t.DB.Dialect()

		err := sqlbuilder.AddConstraint(db, dialect).
			Table("users").
			Unique("u_user_name", "name").
			Exec()
		t.NotError(err, "%s@%s", err, t.DriverName)

		// 删除约束
		err = sqlbuilder.DropConstraint(db, dialect).
			Table("users").
			Constraint("u_user_name").
			Exec()
		a.NotError(err, "%s@%s", err, t.DriverName)

		// 删除不存在的约束名
		err = sqlbuilder.DropConstraint(db, dialect).
			Table("users").
			Constraint("u_user_name_not_exists___").
			Exec()
		a.Error(err, "并未出错 @%s", t.DriverName)
	})
}
