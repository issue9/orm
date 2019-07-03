// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/sqlbuilder"
)

func TestConstraint(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer clearDB(a, db)

	if db.Dialect().Name() != "sqlite3" {
		err := sqlbuilder.AddConstraint(db, db.Dialect()).
			Table("#user").
			Unique("u_user_name", "name").
			Exec()
		a.NotError(err)

		// 删除约束
		err = sqlbuilder.DropConstraint(db, db.Dialect()).
			Table("#user").
			Constraint("u_user_name").
			Exec()
		a.NotError(err)

		// 不存在的约束名
		err = sqlbuilder.DropConstraint(db, db.Dialect()).
			Table("#user").
			Constraint("u_user_name_not_exists___").
			Exec()
		a.Error(err)
	}
}
