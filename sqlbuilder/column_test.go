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

		addStmt := sqlbuilder.AddColumn(db, dialect)
		err := addStmt.Table("users").
			Column("col1", reflect.TypeOf(1), true, false, nil).
			Exec()
		a.NotError(err, "%s@%s", err, t.DriverName)

		dropStmt := sqlbuilder.DropColumn(db, dialect)
		err = dropStmt.Table("users").
			Column("col1").
			Exec()
		t.NotError(err, "%s@%s", err, t.DriverName)

		err = addStmt.Reset().Exec()
		a.ErrorType(err, sqlbuilder.ErrTableIsEmpty)

		err = addStmt.Reset().Table("users").Exec()
		a.ErrorType(err, sqlbuilder.ErrColumnsIsEmpty)

		err = dropStmt.Reset().Exec()
		a.ErrorType(err, sqlbuilder.ErrTableIsEmpty)
	}) // end suite.ForEach
}
