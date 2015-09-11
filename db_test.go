// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

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

// 初始化测试数据，同时可用于DB.Inert的测试
// 清空其它数据，初始化成原始的测试数据
func initData(db *DB, a *assert.Assertion) {
	a.NotError(db.Drop(&user{}))
	a.NotError(db.Drop(&admin{}))
	a.NotError(db.Drop(&userInfo{}))
	a.NotError(db.Create(&admin{}))
	a.NotError(db.Create(&userInfo{}))

	a.NotError(db.Insert(&admin{
		user:  user{Username: "username1", Password: "password1"},
		Email: "email1",
		Group: 1,
	}))
	a.NotError(db.Insert(&userInfo{
		UID:       1,
		FirstName: "f1",
		LastName:  "l1",
		Sex:       "female",
	}))
	a.NotError(db.Insert(&userInfo{ // sex使用默认值
		UID:       2,
		FirstName: "f2",
		LastName:  "l2",
	}))

	// select
	u1 := &userInfo{UID: 1}
	u2 := &userInfo{LastName: "l2", FirstName: "f2"}
	a1 := &admin{Email: "email1"}

	a.NotError(db.Select(u1))
	a.Equal(u1, &userInfo{UID: 1, FirstName: "f1", LastName: "l1", Sex: "female"})

	a.NotError(db.Select(u2))
	a.Equal(u2, &userInfo{UID: 2, FirstName: "f2", LastName: "l2", Sex: "male"})

	a.NotError(db.Select(a1))
	a.Equal(a1.Username, "username1")
}

func clearData(db *DB, a *assert.Assertion) {
	a.NotError(db.Drop(&admin{}))
	a.NotError(db.Drop(&user{}))
	a.NotError(db.Drop(&userInfo{}))
	a.NotError(db.Close())
	closeDB(a)
}

func TestDB_Update(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	initData(db, a)
	defer clearData(db, a)

	// update
	a.NotError(db.Update(&userInfo{
		UID:       1,
		FirstName: "firstName1",
		LastName:  "lastName1",
		Sex:       "sex1",
	}))
	a.NotError(db.Update(&userInfo{
		UID:       2,
		FirstName: "firstName2",
		LastName:  "lastName2",
		Sex:       "sex2",
	}))

	u1 := &userInfo{UID: 1}
	a.NotError(db.Select(u1))
	a.Equal(u1, &userInfo{UID: 1, FirstName: "firstName1", LastName: "lastName1", Sex: "sex1"})

	u2 := &userInfo{LastName: "lastName2", FirstName: "firstName2"}
	a.NotError(db.Select(u2))
	a.Equal(u2, &userInfo{UID: 2, FirstName: "firstName2", LastName: "lastName2", Sex: "sex2"})
}

func TestDB_Delete(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	initData(db, a)
	defer clearData(db, a)

	// delete
	a.NotError(db.Delete(&userInfo{UID: 1}))

	a.NotError(db.Delete(
		&userInfo{
			LastName:  "l2",
			FirstName: "f2",
		}))
	a.NotError(db.Delete(&admin{Email: "email1"}))

	hasCount(db, a, "user_info", 0)
	hasCount(db, a, "administrators", 0)

	// delete并不会重置ai计数
	a.NotError(db.Insert(&admin{Group: 1, Email: "email1"}))
	a1 := &admin{Email: "email1"}
	a.NotError(db.Select(a1))
	a.Equal(a1.ID, 2) // a1.ID为一个自增列,不会在delete中被重置
}

func TestDB_Count(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	initData(db, a)
	defer clearData(db, a)

	// 单条件
	count, err := db.Count(
		&userInfo{
			UID: 1,
		},
	)
	a.NotError(err).Equal(1, count)

	// 条件不存在
	count, err = db.Count(
		&admin{Email: "email1-1000"}, // 该条件不存在
	)
	a.NotError(err).Equal(0, count)
}

func TestDB_Truncate(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	initData(db, a)
	defer clearData(db, a)

	hasCount(db, a, "administrators", 1)
	hasCount(db, a, "user_info", 2)

	// truncate之后，会重置AI
	a.NotError(db.Truncate(&admin{}))
	a.NotError(db.Truncate(&userInfo{}))
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
	initData(db, a)
	defer clearData(db, a)

	a.NotError(db.Drop(&userInfo{}))
	a.NotError(db.Drop(&admin{}))
	a.Error(db.Insert(&admin{}))
}
