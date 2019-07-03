// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"reflect"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/sqltest"
	"github.com/issue9/orm/v2/internal/testconfig"
	"github.com/issue9/orm/v2/sqlbuilder"
)

var (
	_ sqlbuilder.DDLSQLer = &sqlbuilder.CreateTableStmt{}
	_ sqlbuilder.DDLSQLer = &sqlbuilder.DropTableStmt{}
)

func TestCreateTableStmt(t *testing.T) {
	a := assert.New(t)
	table := "create_table_test"

	db := testconfig.NewDB(a)
	defer testconfig.CloseDB(db, a)

	stmt := sqlbuilder.CreateTable(db, db.Dialect()).
		Table(table).
		AutoIncrement("create_table_test_ai", "id", reflect.TypeOf(int(1))).
		Column("age", reflect.TypeOf(int(1)), false, false, nil).
		Column("name", reflect.TypeOf(""), true, true, "", 100).
		Column("address", reflect.TypeOf(""), false, false, nil, 100).
		Index(sqlbuilder.IndexDefault, "index_index", "name", "address").
		Unique("u_age", "name", "address").
		Check("age_gt_0", "age>0")
	err := stmt.Exec()
	a.NotError(err)

	insert := sqlbuilder.Insert(db, db.Dialect()).
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

	cnt, err := sqlbuilder.Select(db, db.Dialect()).
		Select("count(*) AS cnt").
		From(table).
		QueryInt("cnt")
	a.NotError(err).Equal(cnt, 3)

	err = sqlbuilder.DropTable(db, db.Dialect()).
		Table(table).
		Exec()
	a.NotError(err)
}

func TestDropTable(t *testing.T) {
	a := assert.New(t)

	drop := sqlbuilder.DropTable(nil, nil).
		Table("table").
		Table("tbl2")
	sql, err := drop.DDLSQL()
	a.NotError(err).
		Equal(2, len(sql))
	sqltest.Equal(a, sql[0], "drop table if exists table")
	sqltest.Equal(a, sql[1], "drop table if exists tbl2")

	drop.Reset()
	sql, err = drop.DDLSQL()
	a.Equal(err, sqlbuilder.ErrTableIsEmpty).Empty(sql)
}
