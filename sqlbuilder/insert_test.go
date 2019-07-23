// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"database/sql"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/core"
	"github.com/issue9/orm/v2/internal/test"
	"github.com/issue9/orm/v2/sqlbuilder"
)

var _ sqlbuilder.SQLer = &sqlbuilder.InsertStmt{}

func TestInsert(t *testing.T) {
	a := assert.New(t)
	s := test.NewSuite(a)
	defer s.Close()
	tableName := "users"

	s.ForEach(func(t *test.Driver) {
		err := sqlbuilder.CreateTable(t.DB).
			Table(tableName).
			AutoIncrement("id", core.Int64Type).
			Column("name", core.StringType, false, false, nil, 20).
			Exec()
		a.NotError(err)
		defer func() {
			err := sqlbuilder.DropTable(t.DB).
				Table(tableName).
				Exec()
			a.NotError(err)
		}()

		i := sqlbuilder.Insert(t.DB).Table(tableName)
		a.NotNil(i)

		i.Columns("id", "name").Values(10, "name10").Values(11, "name11")
		_, err = i.Exec()
		a.NotError(err)

		i.Reset().Table("tb1").
			Table(tableName).
			KeyValue("id", 20).
			KeyValue("name", "name20")
		_, err = i.Exec()
		a.NotError(err)

		i.Reset().Columns("id", "name")
		_, err = i.Exec()
		a.ErrorType(err, sqlbuilder.ErrTableIsEmpty)

		i.Reset().Table(tableName).Columns("id", "name")
		_, err = i.Exec()
		a.ErrorType(err, sqlbuilder.ErrValueIsEmpty)

		i.Reset().Table(tableName).Columns("id", "name").Values("100")
		_, err = i.Exec()
		a.ErrorType(err, sqlbuilder.ErrArgsNotMatch)
	})
}

func TestInsert_NamedArgs(t *testing.T) {
	a := assert.New(t)
	s := test.NewSuite(a)
	defer s.Close()
	tableName := "users"

	s.ForEach(func(t *test.Driver) {
		err := sqlbuilder.CreateTable(t.DB).
			Table(tableName).
			AutoIncrement("id", core.Int64Type).
			Column("name", core.StringType, false, false, nil, 20).
			Exec()
		a.NotError(err)
		defer func() {
			err := sqlbuilder.DropTable(t.DB).
				Table(tableName).
				Exec()
			a.NotError(err)
		}()

		i := sqlbuilder.Insert(t.DB).Table(tableName)
		i.Reset().Table(tableName).
			Columns("id", "name").
			Values(sql.Named("id", 1), sql.Named("name", "name1"))
		_, err = i.Exec()
		t.NotError(err)

		// 预编译
		stmt, err := i.Prepare()
		a.NotError(err).NotNil(stmt)
		_, err = stmt.Exec(sql.Named("id", 2), sql.Named("name", "name2"))
		a.NotError(err)
		_, err = stmt.Exec(sql.Named("id", 3), sql.Named("name", "name3"))
		a.NotError(err)

		// 部分参数类型不正确
		_, err = stmt.Exec(sql.Named("id", 4), "name4")
		a.Error(err)

		// 参数类型不正确
		_, err = stmt.Exec(5, "name5")
		a.Error(err)
	})
}
