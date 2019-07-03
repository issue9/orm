// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"reflect"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/sqlbuilder"
)

var (
	_ sqlbuilder.DDLSQLer = &sqlbuilder.AddColumnStmt{}
	_ sqlbuilder.DDLSQLer = &sqlbuilder.DropColumnStmt{}
)

func TestColumn(t *testing.T) {
	a := assert.New(t)

	db := initDB(a)
	defer clearDB(a, db)

	err := sqlbuilder.AddColumn(db, db.Dialect()).
		Table("#user").
		Column("col1", reflect.TypeOf(1), true, false, nil).
		Exec()
	a.NotError(err)

	if db.Dialect().Name() != "sqlite3" {
		err = sqlbuilder.DropColumn(db, db.Dialect()).
			Table("#user").
			Column("col1").
			Exec()
		a.NotError(err)
	}
}
