// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"testing"

	"github.com/issue9/assert"

	_ "github.com/mattn/go-sqlite3"
)

func TestDelete(t *testing.T) {
	a := assert.New(t)
	db, e := initDB(a)
	a.NotNil(db).NotNil(e)
	defer closeDB(e, db, a)

	sql := e.SQL().
		Table("#user").
		Where("id<>?", 1).
		And("{group}<>?", 1)

	a.StringEqual(sql.deleteSQL(), "DELETE FROM prefix_user WHERE(id<>?) AND([group]<>?)", style)
}

func TestUpdate(t *testing.T) {
	a := assert.New(t)
	db, e := initDB(a)
	a.NotNil(db).NotNil(e)
	defer closeDB(e, db, a)

	sql := e.SQL().
		Table("user").
		Where("id=?", 1).
		Add("email", "admin@example.com").
		Add("{group}", 1).
		Add("password", "password")

	a.StringEqual(sql.updateSQL(), "UPDATE user SET email=?,[group]=?,password=? WHERE(id=?)", style)
}

func TestInsert(t *testing.T) {
	a := assert.New(t)
	db, e := initDB(a)
	a.NotNil(db).NotNil(e)
	defer closeDB(e, db, a)

	sql := e.SQL().
		Table("#user").
		Add("email", "admin@example.com").
		Add("{group}", 1).
		Add("password", "password")

	a.StringEqual(sql.insertSQL(), "INSERT INTO prefix_user(email,[group],password) VALUES(?,?,?)", style)
}

func TestSelect(t *testing.T) {
	a := assert.New(t)
	db, e := initDB(a)
	a.NotNil(db).NotNil(e)
	defer closeDB(e, db, a)

	sql := e.SQL()
	sql.Table("#user")
}
