// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"os"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/fetch"
	_ "github.com/mattn/go-sqlite3"
)

func closeDB(a *assert.Assertion) {
	if _, err := os.Stat("./test.db"); err == nil || os.IsExist(err) {
		a.NotError(os.Remove("./test.db"))
	}
}

func newDB(a *assert.Assertion) *DB {
	db, err := NewDB("sqlite3", "./test.db", "sqlite3_", &sqlite3{})
	a.NotError(err).NotNil(db)
	return db
}

// table表中是否存在size记录，若不是，则触发error
func hasCount(db *DB, a *assert.Assertion, table string, size int) {
	rows, err := db.Query(true, "SELECT COUNT(*) as cnt FROM #"+table)
	a.NotError(err).NotNil(rows)
	defer func() {
		a.NotError(rows.Close())
	}()

	data, err := fetch.Map(true, rows)
	a.NotError(err).NotNil(data)
	a.Equal(size, data[0]["cnt"])

}

func TestNewDB(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer func() {
		a.NotError(db.Close())
	}()

	a.Equal(db.prefix(), "sqlite3_")
	a.NotNil(db.stdDB).NotNil(db.dialect).NotNil(db.replacer)
	a.Equal(db.stdDB, db.StdDB()).Equal(db.dialect, db.Dialect())

}

func TestDB_Create(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer func() {
		a.NotError(db.Close())
	}()

	a.NotError(db.Create(&admin{}, &userInfo{}))
}

func TestDB_Insert(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer func() {
		a.NotError(db.Close())
	}()

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

	// select
	u1 := &userInfo{UID: 1}
	u2 := &userInfo{LastName: "l2", FirstName: "f2"}
	a1 := &admin{Email: "email1"}

	a.NotError(db.Select(u1, u2, a1))
	a.Equal(u1, &userInfo{UID: 1, FirstName: "f1", LastName: "l1", Sex: "female"})
	a.Equal(u2, &userInfo{UID: 2, FirstName: "f2", LastName: "l2", Sex: "male"})
	a.Equal(a1.Username, "username1")
}

func TestDB_Update(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer func() {
		a.NotError(db.Close())
	}()

	// update
	a.NotError(db.Update(&userInfo{
		UID:       1,
		FirstName: "firstName1",
		LastName:  "lastName1",
		Sex:       "sex1",
	}, &userInfo{
		UID:       2,
		FirstName: "firstName2",
		LastName:  "lastName2",
		Sex:       "sex2",
	}))

	u1 := &userInfo{UID: 1}
	u2 := &userInfo{LastName: "lastName2", FirstName: "firstName2"}

	a.NotError(db.Select(u1, u2))
	a.Equal(u1, &userInfo{UID: 1, FirstName: "firstName1", LastName: "lastName1", Sex: "sex1"})
	a.Equal(u2, &userInfo{UID: 2, FirstName: "firstName2", LastName: "lastName2", Sex: "sex2"})
}

func TestDB_Delete(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer func() {
		a.NotError(db.Close())
	}()

	// delete
	a.NotError(db.Delete(
		&userInfo{
			UID: 1,
		},
		&userInfo{
			LastName:  "lastName2",
			FirstName: "firstName2",
		},
		&admin{Email: "email1"},
	))

	hasCount(db, a, "user_info", 0)
	hasCount(db, a, "administrators", 0)

	// delete并不会重置ai计数
	a.NotError(db.Insert(&admin{Group: 1, Email: "email1"}))
	a1 := &admin{Email: "email1"}
	a.NotError(db.Select(a1))
	a.Equal(a1.ID, 2) // a1.ID为一个自增列,不会在delete中被重置
}

func TestDB_Truncate(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer func() {
		a.NotError(db.Close())
	}()

	a.NotError(db.Truncate(&admin{}, "user_info"))
	hasCount(db, a, "administrators", 0)
	hasCount(db, a, "user_info", 0)

	// 插入数据
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

	hasCount(db, a, "administrators", 1)
	hasCount(db, a, "user_info", 2)

	// truncate之后，会重置AI
	a.NotError(db.Truncate(&admin{}, "user_info"))
	hasCount(db, a, "administrators", 0)
	hasCount(db, a, "user_info", 0)
}

func TestDB_Drop(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer func() {
		a.NotError(db.Close())
	}()

	a.NotError(db.Drop(&admin{}, "user_info"))
	a.Error(db.Insert(&admin{}))
}

// 放在最后，仅用于删除数据库文件
func TestDB_Close(t *testing.T) {
	a := assert.New(t)

	closeDB(a)
}
