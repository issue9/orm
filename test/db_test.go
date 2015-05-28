// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package test

import (
	"testing"

	"github.com/issue9/assert"
)

func TestNewDB(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer func() {
		a.NotError(db.Close())
	}()

	a.Equal(db.Prefix(), prefix)
	a.NotNil(db.StdDB()).NotNil(db.Dialect())
}

func TestDB_Create(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer func() {
		a.NotError(db.Close())
	}()

	a.NotError(db.Drop(&admin{}, &userInfo{}))
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

func TestDB_InsertMany(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer func() {
		a.NotError(db.Close())
	}()

	a.NotError(db.Insert(
		&userInfo{
			UID:       1,
			FirstName: "f1",
			LastName:  "l1",
		}, &userInfo{
			UID:       2,
			FirstName: "f2",
			LastName:  "l2",
		}, &userInfo{
			UID:       3,
			FirstName: "f3",
			LastName:  "l3",
		}))

	// select
	u1 := &userInfo{UID: 1}
	u2 := &userInfo{LastName: "l2", FirstName: "f2"}
	u3 := &userInfo{UID: 3}

	a.NotError(db.Select(u1, u2, u3))
	a.Equal(u1, &userInfo{UID: 1, FirstName: "f1", LastName: "l1", Sex: "male"})
	a.Equal(u2, &userInfo{UID: 2, FirstName: "f2", LastName: "l2", Sex: "male"})
	a.Equal(u3, &userInfo{UID: 3, FirstName: "f3", LastName: "l3", Sex: "male"})
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

	a.NotError(db.Insert(&admin{Group: 1, Email: "email1"}))
	a1 := &admin{Email: "email1"}
	a.NotError(db.Select(a1))
	a.Equal(1, a1.ID)
}

func TestDB_Drop(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer func() {
		a.NotError(db.Close())
	}()

	a.NotError(db.Drop(&admin{}, []byte("user_info"))) // []byte应该能正常转换成string
	a.Error(db.Insert(&admin{}))
}

func TestTX(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer func() {
		a.NotError(db.Close())
	}()

	a.NotError(db.Create(&user{}, &userInfo{}))

	// 回滚事务
	tx, err := db.Begin()
	a.NotError(err).NotNil(tx)
	a.NotError(tx.Insert(&user{Username: "u1"}))
	a.NotError(tx.Insert(&user{Username: "u2"}))
	a.NotError(tx.Insert(&user{Username: "u3"}))
	a.NotError(tx.Rollback())
	hasCount(db, a, "users", 0)

	// 正常提交
	tx, err = db.Begin()
	a.NotError(err).NotNil(tx)
	a.NotError(tx.Insert(&user{Username: "u1"}))
	a.NotError(tx.Insert(&user{Username: "u2"}))
	a.NotError(tx.Insert(&user{Username: "u3"}))
	a.NotError(tx.Commit())
	hasCount(db, a, "users", 3)
}

// 放在最后，销毁数据库内容。
func TestDB_Close(t *testing.T) {
	a := assert.New(t)
	db := newDB(a)
	defer func() {
		a.NotError(db.Close())
	}()

	db.Drop(&user{}, &userInfo{}, &admin{})
	closeDB(a)
}
