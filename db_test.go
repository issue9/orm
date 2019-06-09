// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm_test

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/conv"
	"github.com/issue9/orm/v2"
	"github.com/issue9/orm/v2/fetch"
	"github.com/issue9/orm/v2/internal/testconfig"
)

// table 表中是否存在 size 条记录，若不是，则触发 error
func hasCount(db orm.Engine, a *assert.Assertion, table string, size int) {
	rows, err := db.Query("SELECT COUNT(*) as cnt FROM #" + table)
	a.NotError(err).NotNil(rows)
	defer func() {
		a.NotError(rows.Close())
	}()

	data, err := fetch.Map(true, rows)
	a.NotError(err).NotNil(data)
	a.Equal(conv.MustInt(data[0]["cnt"], -1), size)
}

// 初始化测试数据，同时可当作 DB.Insert 的测试
// 清空其它数据，初始化成原始的测试数据
func initData(a *assert.Assertion) *orm.DB {
	db := testconfig.NewDB(a)

	a.NotError(db.Create(&Group{}))
	a.NotError(db.MultCreate(&Admin{}, &UserInfo{}))

	insert := func(obj interface{}) {
		r, err := db.Insert(obj)
		a.NotError(err)
		cnt, err := r.RowsAffected()
		a.NotError(err).Equal(cnt, 1)
	}

	insert(&Group{
		Name: "group1",
		ID:   1,
	})

	insert(&Admin{
		User:  User{Username: "username1", Password: "password1"},
		Email: "email1",
		Group: 1,
	})

	insert(&UserInfo{
		UID:       1,
		FirstName: "f1",
		LastName:  "l1",
		Sex:       "female",
	})
	insert(&UserInfo{ // sex 使用默认值
		UID:       2,
		FirstName: "f2",
		LastName:  "l2",
	})

	// select
	u1 := &UserInfo{UID: 1}
	u2 := &UserInfo{LastName: "l2", FirstName: "f2"}
	a1 := &Admin{Email: "email1"}

	a.NotError(db.Select(u1))
	a.Equal(u1, &UserInfo{UID: 1, FirstName: "f1", LastName: "l1", Sex: "female"})

	a.NotError(db.Select(u2))
	a.Equal(u2, &UserInfo{UID: 2, FirstName: "f2", LastName: "l2", Sex: "male"})

	a.NotError(db.Select(a1))
	a.Equal(a1.Username, "username1")

	return db
}

func clearData(db *orm.DB, a *assert.Assertion) {
	a.NotError(db.MultDrop(&Admin{}, &User{}))
	a.NotError(db.Drop(&Group{})) // admin 外键依赖 group
	a.NotError(db.Drop(&UserInfo{}))
	testconfig.CloseDB(db, a)
}

func TestDB_Update_error(t *testing.T) {
	a := assert.New(t)

	db := initData(a)
	defer clearData(db, a)

	// 非结构体传入
	r, err := db.Update(123)
	a.Error(err).Nil(r)
}

func TestDB_LastInsertID(t *testing.T) {
	a := assert.New(t)
	db := testconfig.NewDB(a)
	defer clearData(db, a)

	a.NotError(db.Drop(&User{}))
	a.NotError(db.Create(&User{}))

	id, err := db.LastInsertID(&User{Username: "1"})
	a.NotError(err).Equal(id, 1)

	id, err = db.LastInsertID(&User{Username: "2"})
	a.NotError(err).Equal(id, 2)
}

func TestDB_Update(t *testing.T) {
	a := assert.New(t)

	db := initData(a)
	defer clearData(db, a)

	// update
	r, err := db.Update(&UserInfo{
		UID:       1,
		FirstName: "firstName1",
		LastName:  "lastName1",
		Sex:       "sex1",
	})
	a.NotError(err)
	cnt, err := r.RowsAffected()
	a.NotError(err).Equal(1, cnt)

	r, err = db.Update(&UserInfo{
		UID:       2,
		FirstName: "firstName2",
		LastName:  "lastName2",
		Sex:       "sex2",
	})
	a.NotError(err)
	cnt, err = r.RowsAffected()
	a.NotError(err).Equal(1, cnt)

	u1 := &UserInfo{UID: 1}
	a.NotError(db.Select(u1))
	a.Equal(u1, &UserInfo{UID: 1, FirstName: "firstName1", LastName: "lastName1", Sex: "sex1"})

	u2 := &UserInfo{LastName: "lastName2", FirstName: "firstName2"}
	a.NotError(db.Select(u2))
	a.Equal(u2, &UserInfo{UID: 2, FirstName: "firstName2", LastName: "lastName2", Sex: "sex2"})
}

func TestDB_Delete(t *testing.T) {
	a := assert.New(t)

	db := initData(a)
	defer clearData(db, a)

	// delete
	r, err := db.Delete(&UserInfo{UID: 1})
	a.NotError(err)
	cnt, err := r.RowsAffected()
	a.NotError(err).Equal(cnt, 1)

	r, err = db.Delete(
		&UserInfo{
			LastName:  "l2",
			FirstName: "f2",
		})
	a.NotError(err)
	cnt, err = r.RowsAffected()
	a.NotError(err).Equal(cnt, 1)

	r, err = db.Delete(&Admin{Email: "email1"})
	a.NotError(err)
	cnt, err = r.RowsAffected()
	a.NotError(err).Equal(cnt, 1)

	hasCount(db, a, "user_info", 0)
	hasCount(db, a, "administrators", 0)

	// delete并不会重置ai计数
	_, err = db.Insert(&Admin{Group: 1, Email: "email1"})
	a.NotError(err)
	a1 := &Admin{Email: "email1"}
	a.NotError(db.Select(a1))
	a.Equal(a1.ID, 2) // a1.ID为一个自增列,不会在delete中被重置
}

func TestDB_Count(t *testing.T) {
	a := assert.New(t)

	db := initData(a)
	defer clearData(db, a)

	// 单条件
	count, err := db.Count(
		&UserInfo{
			UID: 1,
		},
	)
	a.NotError(err).Equal(1, count)

	// 无条件
	count, err = db.Count(&UserInfo{})
	a.NotError(err).Equal(2, count)

	// 条件不存在
	count, err = db.Count(
		&Admin{Email: "email1-1000"}, // 该条件不存在
	)
	a.NotError(err).Equal(0, count)
}

func TestDB_Truncate(t *testing.T) {
	a := assert.New(t)

	db := initData(a)
	defer clearData(db, a)

	hasCount(db, a, "administrators", 1)
	hasCount(db, a, "user_info", 2)

	// truncate之后，会重置AI
	a.NotError(db.Truncate(&Admin{}))
	a.NotError(db.Truncate(&UserInfo{}))
	hasCount(db, a, "administrators", 0)
	hasCount(db, a, "user_info", 0)

	_, err := db.Insert(&Admin{Group: 1, Email: "email1"})
	a.NotError(err)

	a1 := &Admin{Email: "email1"}
	a.NotError(db.Select(a1))
	a.Equal(1, a1.ID)
}

func TestDB_Drop(t *testing.T) {
	a := assert.New(t)

	db := initData(a)
	defer clearData(db, a)

	a.NotError(db.Drop(&UserInfo{}))
	a.NotError(db.Drop(&Admin{}))
	r, err := db.Insert(&Admin{})
	a.Error(err).Nil(r)
}
