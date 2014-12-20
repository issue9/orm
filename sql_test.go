// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/dialect"

	_ "github.com/mattn/go-sqlite3"
)

var style = assert.StyleSpace | assert.StyleTrim

func newDB(a *assert.Assertion) *Engine {
	if !dialect.IsRegisted("sqlite3") {
		a.NotError(dialect.Register("sqlite3", &dialect.Sqlite3{}))
	}

	if db, found := Get("sqlite3"); found {
		return db
	}

	db, err := New("sqlite3", "./test.db", "sqlite3", "prefix_")
	a.NotError(err).NotNil(db)

	return db
}

func TestDelete(t *testing.T) {
	a := assert.New(t)
	db := newDB(a)
	defer db.close()

	sql := db.SQL().
		Table("#user").
		Where("id<>?", 1).
		And("{group}<>?", 1)

	a.StringEqual(sql.deleteSQL(), "DELETE FROM prefix_user WHERE(id<>?) AND([group]<>?)", style)
}

func TestUpdate(t *testing.T) {
	a := assert.New(t)
	db := newDB(a)
	defer db.close()

	sql := db.SQL().
		Table("user").
		Where("id=?", 1).
		Add("email", "admin@example.com").
		Add("{group}", 1).
		Add("password", "password")

	a.StringEqual(sql.updateSQL(), "UPDATE user SET email=?,[group]=?,password=? WHERE(id=?)", style)
}

func TestInsert(t *testing.T) {
	a := assert.New(t)
	db := newDB(a)
	defer db.close()

	sql := db.SQL().
		Table("#user").
		Add("email", "admin@example.com").
		Add("{group}", 1).
		Add("password", "password")

	a.StringEqual(sql.insertSQL(), "INSERT INTO prefix_user(email,[group],password) VALUES(?,?,?)", style)
}

func TestSelect(t *testing.T) {
	a := assert.New(t)
	db := newDB(a)
	defer db.close()

	sql := db.SQL()
	sql.Table("#user")
}
