// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/test"
	"github.com/issue9/orm/v2/sqlbuilder"
)

var _ sqlbuilder.SQLer = &sqlbuilder.InsertStmt{}

func TestInsert(t *testing.T) {
	a := assert.New(t)
	s := test.NewSuite(a)
	defer s.Close()

	s.ForEach(func(t *test.Test) {
		initDB(t)
		defer clearDB(t)

		db := t.DB.DB
		dialect := t.DB.Dialect()

		i := sqlbuilder.Insert(db, dialect).Table("users")
		a.NotNil(i)

		i.Columns("id", "name").Values(10, "name10").Values(11, "name11")
		_, err := i.Exec()
		a.NotError(err)

		i.Reset()
		i.Table("tb1").
			Table("users").
			KeyValue("id", 20).
			KeyValue("name", "name20")
		_, err = i.Exec()
		a.NotError(err)

		i.Reset()
		i.Columns("id", "name")
		_, err = i.Exec()
		a.ErrorType(err, sqlbuilder.ErrTableIsEmpty)

		i.Reset()
		i.Table("users").Columns("id", "name")
		_, err = i.Exec()
		a.ErrorType(err, sqlbuilder.ErrValueIsEmpty)

		i.Reset()
		i.Table("users").Columns("id", "name").Values("100")
		_, err = i.Exec()
		a.ErrorType(err, sqlbuilder.ErrArgsNotMatch)
	})
}
