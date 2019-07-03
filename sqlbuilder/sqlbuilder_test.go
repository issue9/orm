// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2"
	"github.com/issue9/orm/v2/internal/testconfig"
	"github.com/issue9/orm/v2/sqlbuilder"
)

type user struct {
	ID   int64  `orm:"name(id);ai"`
	Name string `orm:"name(name);len(20)"`
}

func (u *user) Meta() string {
	return "name(user)"
}

func initDB(a *assert.Assertion) *orm.DB {
	db := testconfig.NewDB(a)

	a.NotError(db.Create(&user{}))

	sql := sqlbuilder.Insert(db, db.Dialect()).
		Columns("name").
		Table("#user").
		Values("1").
		Values("2")
	_, err := sql.Exec()
	a.NotError(err)

	stmt, err := sql.Prepare()
	a.NotError(err).NotNil(stmt)

	_, err = stmt.Exec("3", "4")
	a.NotError(err)
	_, err = stmt.Exec("5", "6")
	a.NotError(err)

	sql.Reset()

	sql.Table("#user").
		Columns("name").
		Values("7")
	id, err := sql.LastInsertID("user", "id")
	a.NotError(err).Equal(id, 7)

	// 多行插入，不能拿到 lastInsertID
	sql.Table("#user").
		Columns("name").
		Values("8").
		Values("9")
	id, err = sql.LastInsertID("user", "id")
	a.Error(err).Empty(id)

	return db
}

func clearDB(a *assert.Assertion, db *orm.DB) {
	err := sqlbuilder.DropTable(db, db.Dialect()).Table("#user").Exec()
	a.NotError(err)
	testconfig.CloseDB(db, a)
}

func TestSQLBuilder(t *testing.T) {
	a := assert.New(t)

	b := sqlbuilder.New("")
	b.WriteByte('1')
	b.WriteString("23")

	a.Equal("123", b.String())
	a.Equal(3, b.Len())

	b.Reset()
	a.Equal(b.String(), "")
	a.Equal(b.Len(), 0)

	b.WriteByte('3').WriteString("21")
	a.Equal(b.String(), "321")

	b.TruncateLast(1)
	a.Equal(b.String(), "32").Equal(2, b.Len())
}
