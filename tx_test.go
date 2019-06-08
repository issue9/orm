// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/dialect"
	"github.com/issue9/orm/v2/internal/testconfig"
)

func TestTx_InsertMany(t *testing.T) {
	a := assert.New(t)

	db := testconfig.NewDB(a)
	defer clearData(db, a)

	tx, err := db.Begin()
	a.NotError(err)
	a.NotError(tx.Create(&UserInfo{}))

	a.NotError(tx.InsertMany([]*UserInfo{
		&UserInfo{
			UID:       1,
			FirstName: "f1",
			LastName:  "l1",
		},
	}, 10))

	// 分批插入
	a.NotError(tx.InsertMany([]*UserInfo{
		&UserInfo{
			UID:       2,
			FirstName: "f2",
			LastName:  "l2",
		},
		&UserInfo{
			UID:       3,
			FirstName: "f3",
			LastName:  "l3",
		},
		&UserInfo{
			UID:       4,
			FirstName: "f4",
			LastName:  "l4",
		},
		&UserInfo{
			UID:       5,
			FirstName: "f5",
			LastName:  "l5",
		},
		&UserInfo{
			UID:       6,
			FirstName: "f6",
			LastName:  "l6",
		},
	}, 2))

	// 单个元素插入
	a.NotError(tx.InsertMany(
		&UserInfo{
			UID:       7,
			FirstName: "f7",
			LastName:  "l7",
		}, 10))

	a.NotError(tx.Commit())

	u1 := &UserInfo{UID: 1}
	a.NotError(db.Select(u1))
	a.Equal(u1, &UserInfo{UID: 1, FirstName: "f1", LastName: "l1", Sex: "male"})

	u2 := &UserInfo{LastName: "l2", FirstName: "f2"}
	a.NotError(db.Select(u2))
	a.Equal(u2, &UserInfo{UID: 2, FirstName: "f2", LastName: "l2", Sex: "male"})

	u3 := &UserInfo{UID: 3}
	a.NotError(db.Select(u3))
	a.Equal(u3, &UserInfo{UID: 3, FirstName: "f3", LastName: "l3", Sex: "male"})

	u4 := &UserInfo{UID: 4}
	a.NotError(db.Select(u4))
	a.Equal(u4, &UserInfo{UID: 4, FirstName: "f4", LastName: "l4", Sex: "male"})

	u5 := &UserInfo{UID: 5}
	a.NotError(db.Select(u5))
	a.Equal(u5, &UserInfo{UID: 5, FirstName: "f5", LastName: "l5", Sex: "male"})

	u6 := &UserInfo{UID: 6}
	a.NotError(db.Select(u6))
	a.Equal(u6, &UserInfo{UID: 6, FirstName: "f6", LastName: "l6", Sex: "male"})

	u7 := &UserInfo{UID: 7}
	a.NotError(db.Select(u7))
	a.Equal(u7, &UserInfo{UID: 7, FirstName: "f7", LastName: "l7", Sex: "male"})

	// 类型错误
	tx, err = db.Begin()
	a.NotError(err)
	a.NotError(tx.Create(&UserInfo{}))
	a.Error(tx.InsertMany(5, 10))
	a.NotError(tx.Rollback())

	// 类型错误
	tx, err = db.Begin()
	a.NotError(err)
	a.NotError(tx.Create(&UserInfo{}))
	a.Error(tx.InsertMany([]int{1, 2, 3}, 10))
	a.NotError(tx.Rollback())
}

func TestTx_LastInsertID(t *testing.T) {
	a := assert.New(t)
	db := testconfig.NewDB(a)
	defer clearData(db, a)

	a.NotError(db.Drop(&User{}))
	a.NotError(db.Create(&User{}))

	tx, err := db.Begin()
	a.NotError(err)

	id, err := tx.LastInsertID(&User{Username: "1"})
	a.NotError(err).Equal(id, 1)

	id, err = tx.LastInsertID(&User{Username: "2"})
	a.NotError(err).Equal(id, 2)

	a.NotError(tx.Commit())
}

func TestTx_Insert(t *testing.T) {
	a := assert.New(t)

	db := testconfig.NewDB(a)
	defer clearData(db, a)

	tx, err := db.Begin()
	a.NotError(err)
	a.NotError(tx.Create(&User{}))

	r, err := tx.Insert(&User{
		ID:       1,
		Username: "u1",
	})
	a.NotError(err)
	if db.Dialect().Name() != "postgres" { // postgresql 默认情况下不支持 lastID
		id, err := r.LastInsertId()
		a.NotError(err).Equal(id, 1)
	}

	r, err = tx.Insert(&User{
		ID:       2,
		Username: "u2",
	})
	a.NotError(err)
	if db.Dialect() != dialect.Postgres() {
		id, err := r.LastInsertId()
		a.NotError(err).Equal(id, 2)
	}

	r, err = tx.Insert(&User{
		ID:       3,
		Username: "u3",
	})
	a.NotError(err)
	if db.Dialect().Name() != "postgres" {
		id, err := r.LastInsertId()
		a.NotError(err).Equal(id, 3)
	}

	a.NotError(tx.Commit())

	u1 := &User{ID: 1}
	a.NotError(db.Select(u1))
	a.Equal(u1, &User{ID: 1, Username: "u1"})

	u3 := &User{ID: 3}
	a.NotError(db.Select(u3))
	a.Equal(u3, &User{ID: 3, Username: "u3"})
}

func TestTx_Update(t *testing.T) {
	a := assert.New(t)

	db := testconfig.NewDB(a)
	initData(db, a)
	defer clearData(db, a)

	tx, err := db.Begin()
	a.NotError(err).NotNil(tx)

	// update
	a.NotError(tx.MultUpdate(&UserInfo{
		UID:       1,
		FirstName: "firstName1",
		LastName:  "lastName1",
		Sex:       "sex1",
	}, &UserInfo{
		UID:       2,
		FirstName: "firstName2",
		LastName:  "lastName2",
		Sex:       "sex2",
	}))
	a.NotError(tx.Commit())

	u1 := &UserInfo{UID: 1}
	a.NotError(db.Select(u1))
	a.Equal(u1, &UserInfo{UID: 1, FirstName: "firstName1", LastName: "lastName1", Sex: "sex1"})

	u2 := &UserInfo{LastName: "lastName2", FirstName: "firstName2"}
	a.NotError(db.Select(u2))
	a.Equal(u2, &UserInfo{UID: 2, FirstName: "firstName2", LastName: "lastName2", Sex: "sex2"})
}

func TestTx_Delete(t *testing.T) {
	a := assert.New(t)

	db := testconfig.NewDB(a)
	initData(db, a)
	defer clearData(db, a)

	tx, err := db.Begin()
	a.NotError(err)

	// delete
	a.NotError(tx.MultDelete(
		&UserInfo{
			UID: 1,
		},
		&UserInfo{
			LastName:  "l2",
			FirstName: "f2",
		},
		&Admin{Email: "email1"},
	))

	hasCount(tx, a, "user_info", 0)
	hasCount(tx, a, "administrators", 0)

	// delete并不会重置ai计数
	a.NotError(tx.Insert(&Admin{Group: 1, Email: "email1"}))

	a.NotError(tx.Commit())

	a1 := &Admin{Email: "email1"}
	a.NotError(db.Select(a1))
	a.Equal(a1.ID, 2) // a1.ID为一个自增列,不会在delete中被重置
}

func TestTx_Count(t *testing.T) {
	a := assert.New(t)

	db := testconfig.NewDB(a)
	initData(db, a)
	defer clearData(db, a)

	tx, err := db.Begin()
	a.NotError(err)

	// 单条件
	count, err := tx.Count(
		&UserInfo{
			UID: 1,
		},
	)
	a.NotError(tx.Commit())
	a.NotError(err).Equal(1, count)

	tx, err = db.Begin()
	a.NotError(err)
	count, err = tx.Count(
		&Admin{Email: "email1-1000"}, // 该条件不存在
	)
	a.NotError(tx.Commit())
	a.NotError(err).Equal(0, count)
}

func TestTx_Truncate(t *testing.T) {
	a := assert.New(t)

	db := testconfig.NewDB(a)
	initData(db, a)
	defer clearData(db, a)

	hasCount(db, a, "administrators", 1)
	hasCount(db, a, "user_info", 2)

	// truncate之后，会重置AI
	tx, err := db.Begin()
	a.NotError(err)
	a.NotError(tx.MultTruncate(&Admin{}, &UserInfo{}))
	a.NotError(tx.Commit())
	hasCount(db, a, "administrators", 0)
	hasCount(db, a, "user_info", 0)

	tx, err = db.Begin()
	a.NotError(err)
	a.NotError(tx.Insert(&Admin{Group: 1, Email: "email1"}))
	a.NotError(tx.Commit())
	a1 := &Admin{Email: "email1"}
	a.NotError(db.Select(a1))
	a.Equal(1, a1.ID)
}

func TestTX(t *testing.T) {
	a := assert.New(t)

	db := testconfig.NewDB(a)
	defer clearData(db, a)

	a.NotError(db.Create(&User{}))
	a.NotError(db.Create(&UserInfo{}))

	// 回滚事务
	tx, err := db.Begin()
	a.NotError(err).NotNil(tx)
	a.NotError(tx.Insert(&User{Username: "u1"}))
	a.NotError(tx.Insert(&User{Username: "u2"}))
	a.NotError(tx.Insert(&User{Username: "u3"}))
	a.NotError(tx.Rollback())
	hasCount(db, a, "users", 0)

	// 正常提交
	tx, err = db.Begin()
	a.NotError(err).NotNil(tx)
	a.NotError(tx.Insert(&User{Username: "u1"}))
	a.NotError(tx.Insert(&User{Username: "u2"}))
	a.NotError(tx.Insert(&User{Username: "u3"}))
	a.NotError(tx.Commit())
	hasCount(db, a, "users", 3)
}
