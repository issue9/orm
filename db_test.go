// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/core"
)

var _ core.Engine = &DB{}

func TestNewDB(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer func() {
		a.NotError(db.Close())
	}()

	a.Equal(db.tablePrefix, prefix)
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

	insert := func(obj interface{}) {
		r, err := db.Insert(obj)
		a.NotError(err)
		cnt, err := r.RowsAffected()
		a.NotError(err).Equal(cnt, 1)
	}

	insert(&admin{
		user:  user{Username: "username1", Password: "password1"},
		Email: "email1",
		Group: 1,
	})

	insert(&userInfo{
		UID:       1,
		FirstName: "f1",
		LastName:  "l1",
		Sex:       "female",
	})
	insert(&userInfo{ // sex使用默认值
		UID:       2,
		FirstName: "f2",
		LastName:  "l2",
	})

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
	r, err := db.Update(&userInfo{
		UID:       1,
		FirstName: "firstName1",
		LastName:  "lastName1",
		Sex:       "sex1",
	})
	a.NotError(err)
	cnt, err := r.RowsAffected()
	a.NotError(err).Equal(1, cnt)

	r, err = db.Update(&userInfo{
		UID:       2,
		FirstName: "firstName2",
		LastName:  "lastName2",
		Sex:       "sex2",
	})
	a.NotError(err)
	cnt, err = r.RowsAffected()
	a.NotError(err).Equal(1, cnt)

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
	r, err := db.Delete(&userInfo{UID: 1})
	a.NotError(err)
	cnt, err := r.RowsAffected()
	a.NotError(err).Equal(cnt, 1)

	r, err = db.Delete(
		&userInfo{
			LastName:  "l2",
			FirstName: "f2",
		})
	a.NotError(err)
	cnt, err = r.RowsAffected()
	a.NotError(err).Equal(cnt, 1)

	r, err = db.Delete(&admin{Email: "email1"})
	a.NotError(err)
	cnt, err = r.RowsAffected()
	a.NotError(err).Equal(cnt, 1)

	hasCount(db, a, "user_info", 0)
	hasCount(db, a, "administrators", 0)

	// delete并不会重置ai计数
	_, err = db.Insert(&admin{Group: 1, Email: "email1"})
	a.NotError(err)
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

	_, err := db.Insert(&admin{Group: 1, Email: "email1"})
	a.NotError(err)

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
	r, err := db.Insert(&admin{})
	a.Error(err).Nil(r)
}
