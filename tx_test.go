// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/forward"
)

var _ forward.Engine = &Tx{}

func TestTx_InsertMany(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer clearData(db, a)

	tx, err := db.Begin()
	a.NotError(err)
	a.NotError(tx.Create(&userInfo{}))

	a.NotError(tx.InsertMany([]*userInfo{
		&userInfo{
			UID:       1,
			FirstName: "f1",
			LastName:  "l1",
		},
	}))
	a.NotError(tx.InsertMany([]*userInfo{
		&userInfo{
			UID:       2,
			FirstName: "f2",
			LastName:  "l2",
		}, &userInfo{
			UID:       3,
			FirstName: "f3",
			LastName:  "l3",
		}}))
	a.NotError(tx.Commit())

	u1 := &userInfo{UID: 1}
	a.NotError(db.Select(u1))
	a.Equal(u1, &userInfo{UID: 1, FirstName: "f1", LastName: "l1", Sex: "male"})

	u2 := &userInfo{LastName: "l2", FirstName: "f2"}
	a.NotError(db.Select(u2))
	a.Equal(u2, &userInfo{UID: 2, FirstName: "f2", LastName: "l2", Sex: "male"})

	u3 := &userInfo{UID: 3}
	a.NotError(db.Select(u3))
	a.Equal(u3, &userInfo{UID: 3, FirstName: "f3", LastName: "l3", Sex: "male"})
}

func TestTx_Insert(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer clearData(db, a)

	tx, err := db.Begin()
	a.NotError(err)
	a.NotError(tx.Create(&user{}))

	r, err := tx.Insert(&user{
		ID:       1,
		Username: "u1",
	})
	a.NotError(err)
	id, err := r.LastInsertId()
	a.NotError(err).Equal(id, 1)

	r, err = tx.Insert(&user{
		ID:       2,
		Username: "u2",
	})
	a.NotError(err)
	id, err = r.LastInsertId()
	a.NotError(err).Equal(id, 2)

	r, err = tx.Insert(&user{
		ID:       3,
		Username: "u3",
	})
	a.NotError(err)
	id, err = r.LastInsertId()
	a.NotError(err).Equal(id, 3)

	a.NotError(tx.Commit())

	u1 := &user{ID: 1}
	a.NotError(db.Select(u1))
	a.Equal(u1, &user{ID: 1, Username: "u1"})

	u3 := &user{ID: 3}
	a.NotError(db.Select(u3))
	a.Equal(u3, &user{ID: 3, Username: "u3"})
}

func TestTx_Update(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	initData(db, a)
	defer clearData(db, a)

	tx, err := db.Begin()
	a.NotError(err).NotNil(tx)

	// update
	a.NotError(tx.MultUpdate(&userInfo{
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
	a.NotError(tx.Commit())

	u1 := &userInfo{UID: 1}
	a.NotError(db.Select(u1))
	a.Equal(u1, &userInfo{UID: 1, FirstName: "firstName1", LastName: "lastName1", Sex: "sex1"})

	u2 := &userInfo{LastName: "lastName2", FirstName: "firstName2"}
	a.NotError(db.Select(u2))
	a.Equal(u2, &userInfo{UID: 2, FirstName: "firstName2", LastName: "lastName2", Sex: "sex2"})
}

func TestTx_Delete(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	initData(db, a)
	defer clearData(db, a)

	tx, err := db.Begin()
	a.NotError(err)

	// delete
	a.NotError(tx.MultDelete(
		&userInfo{
			UID: 1,
		},
		&userInfo{
			LastName:  "l2",
			FirstName: "f2",
		},
		&admin{Email: "email1"},
	))

	hasCount(tx, a, "user_info", 0)
	hasCount(tx, a, "administrators", 0)

	// delete并不会重置ai计数
	a.NotError(tx.Insert(&admin{Group: 1, Email: "email1"}))

	a.NotError(tx.Commit())

	a1 := &admin{Email: "email1"}
	a.NotError(db.Select(a1))
	a.Equal(a1.ID, 2) // a1.ID为一个自增列,不会在delete中被重置
}

func TestTx_Count(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	initData(db, a)
	defer clearData(db, a)

	tx, err := db.Begin()
	a.NotError(err)

	// 单条件
	count, err := tx.Count(
		&userInfo{
			UID: 1,
		},
	)
	a.NotError(tx.Commit())
	a.NotError(err).Equal(1, count)

	tx, err = db.Begin()
	a.NotError(err)
	count, err = tx.Count(
		&admin{Email: "email1-1000"}, // 该条件不存在
	)
	a.NotError(tx.Commit())
	a.NotError(err).Equal(0, count)
}

func TestTx_Truncate(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	initData(db, a)
	defer clearData(db, a)

	hasCount(db, a, "administrators", 1)
	hasCount(db, a, "user_info", 2)

	// truncate之后，会重置AI
	tx, err := db.Begin()
	a.NotError(err)
	a.NotError(tx.MultTruncate(&admin{}, &userInfo{}))
	a.NotError(tx.Commit())
	hasCount(db, a, "administrators", 0)
	hasCount(db, a, "user_info", 0)

	tx, err = db.Begin()
	a.NotError(err)
	a.NotError(tx.Insert(&admin{Group: 1, Email: "email1"}))
	a.NotError(tx.Commit())
	a1 := &admin{Email: "email1"}
	a.NotError(db.Select(a1))
	a.Equal(1, a1.ID)
}

func TestTX(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer clearData(db, a)

	a.NotError(db.Create(&user{}))
	a.NotError(db.Create(&userInfo{}))

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
