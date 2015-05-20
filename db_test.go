// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"testing"

	"github.com/issue9/assert"
	_ "github.com/mattn/go-sqlite3"
)

func newDB(a *assert.Assertion) *DB {
	db, err := NewDB("sqlite3", "./test.db", "sqlite3_", &sqlite3{})
	a.NotError(err).NotNil(db)
	return db
}

func TestNewDB(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer func() {
		a.NotError(db.Close())
	}()

	a.Equal(db.prefix, "sqlite3_")
	a.NotNil(db.stdDB).NotNil(db.dialect).NotNil(db.replacer)
	a.Equal(db.stdDB, db.StdDB()).Equal(db.dialect, db.Dialect())
}

func TestDB_Create_Insert(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer func() {
		a.NotError(db.Close())
	}()

	a.NotError(db.Create(&admin{}, &userInfo{}))
	a.NotError(db.Insert(&admin{
		user:  user{Username: "username1", Password: "password1"},
		Email: "email1",
		Group: 1,
	}, &userInfo{
		UID:       1,
		FirstName: "f1",
		LastName:  "l1",
		Sex:       "female",
	}, &userInfo{ // sex使用默认值
		UID:       2,
		FirstName: "f2",
		LastName:  "l2",
	}))
}
