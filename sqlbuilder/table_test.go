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
	_ sqlbuilder.DDLSQLer = &sqlbuilder.CreateTableStmt{}
	_ sqlbuilder.DDLSQLer = &sqlbuilder.DropTableStmt{}
)

func TestCreateTableStmt(t *testing.T) {
	a := assert.New(t)
	table := "create_table_test"
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		db := t.DB.DB
		dialect := t.DB.Dialect()
		stmt := sqlbuilder.CreateTable(db, dialect).
			Table(table).
			AutoIncrement("id", reflect.TypeOf(int(1))).
			Column("age", reflect.TypeOf(int(1)), false, false, nil).
			Column("name", reflect.TypeOf(""), true, true, "", 100).
			Column("address", reflect.TypeOf(""), false, false, nil, 100).
			Index(sqlbuilder.IndexDefault, "index_index", "name", "address").
			Unique("u_age", "name", "address").
			Check("age_gt_0", "age>0")
		err := stmt.Exec()
		a.NotError(err)

		insert := sqlbuilder.Insert(db, dialect).
			Table(table).
			KeyValue("age", 1).
			KeyValue("name", "name1").
			KeyValue("address", "address1")
		rslt, err := insert.Exec()
		a.NotError(err).NotNil(rslt)

		prepare, err := insert.Prepare()
		a.NotError(err).NotNil(prepare)
		rslt, err = prepare.Exec(2, "name2", "address2")
		a.NotError(err).NotNil(rslt)
		rslt, err = prepare.Exec(3, "name3", "address3")
		a.NotError(err).NotNil(rslt)

		cnt, err := sqlbuilder.Select(db, dialect).
			Select("count(*) AS cnt").
			From(table).
			QueryInt("cnt")
		a.NotError(err).Equal(cnt, 3)

		err = sqlbuilder.DropTable(db, dialect).
			Table(table).
			Exec()
		a.NotError(err)
	})
}

func TestTruncateTable(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		initDB(t)
		defer clearDB(t)

		err := sqlbuilder.TruncateTable(t.DB.DB, t.DB.Dialect()).
			Table("users", "id").
			Exec()
		t.NotError(err)

		sel := sqlbuilder.Select(t.DB.DB, t.DB.Dialect()).
			Select("count(*) as cnt").
			From("users")
		rows, err := sel.Query()
		t.NotError(err).NotNil(rows)
		t.True(rows.Next())
		var val int
		t.NotError(rows.Scan(&val))
		t.NotError(rows.Close())
		t.Equal(val, 0)
	})
}
