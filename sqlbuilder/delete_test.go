// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/v2/internal/sqltest"
	"github.com/issue9/orm/v2/internal/testconfig"
	"github.com/issue9/orm/v2/sqlbuilder"
)

var (
	_ sqlbuilder.SQLer       = &sqlbuilder.DeleteStmt{}
	_ sqlbuilder.WhereStmter = &sqlbuilder.DeleteStmt{}
)

func TestDelete(t *testing.T) {
	a := assert.New(t)

	d := sqlbuilder.Delete(nil).
		Table("#table").
		Where("id=?", 1).
		Or("id=?", 2).
		And("id=?", 3)
	query, args, err := d.SQL()
	a.NotError(err)
	a.Equal(args, []interface{}{1, 2, 3})
	sqltest.Equal(a, query, "delete from #table where id=? or id=? and id=?")

	d.Reset()
	query, args, err = d.Table("tb1").Where("id=?").Or("id=?", 1).SQL()
	a.Equal(err, sqlbuilder.ErrArgsNotMatch) // 由 where 抛出
	a.Empty(query).Nil(args)
}

func TestDelete_Exec(t *testing.T) {
	a := assert.New(t)
	db := createTable(a)
	defer testconfig.CloseDB(db, a)

	insertData(a, db)

	sql := sqlbuilder.Delete(db).Table("#user").Where("id=?", 1)
	_, err := sql.Exec()
	a.NotError(err)
}
