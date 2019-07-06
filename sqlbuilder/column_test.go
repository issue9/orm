// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"reflect"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/test"
	"github.com/issue9/orm/v2/sqlbuilder"
)

var (
	_ sqlbuilder.DDLSQLer = &sqlbuilder.AddColumnStmt{}
	_ sqlbuilder.DDLSQLer = &sqlbuilder.DropColumnStmt{}
)

func TestColumn(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		initDB(t)
		defer clearDB(t)

		db := t.DB.DB
		dialect := t.DB.Dialect()
		err := sqlbuilder.AddColumn(db, dialect).
			Table("users").
			Column("col1", reflect.TypeOf(1), true, false, nil).
			Exec()
		a.NotError(err)

		if dialect.Name() != "sqlite3" {
			err = sqlbuilder.DropColumn(db, dialect).
				Table("users").
				Column("col1").
				Exec()
			t.NotError(err)
		}
	}) // end suite.ForEach
}
