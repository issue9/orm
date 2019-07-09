// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm_test

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/conv"

	"github.com/issue9/orm/v2/fetch"
	"github.com/issue9/orm/v2/internal/test"
	"github.com/issue9/orm/v2/sqlbuilder"
)

// table 表中是否存在 size 条记录，若不是，则触发 error
func hasCount(e sqlbuilder.Engine, a *assert.Assertion, table string, size int) {
	rows, err := e.Query("SELECT COUNT(*) as cnt FROM #" + table)
	a.NotError(err).
		NotNil(rows)
	defer func() {
		a.NotError(rows.Close())
	}()

	data, err := fetch.Map(true, rows)
	a.NotError(err).
		NotNil(data)
	a.Equal(conv.MustInt(data[0]["cnt"], -1), size)
}

// 初始化测试数据，同时可当作 DB.Insert 的测试
// 清空其它数据，初始化成原始的测试数据
func initData(t *test.Test) {
	db := t.DB

	err := db.Create(&Group{})
	t.NotError(err, "%s@%s", err, t.DriverName)

	err = db.MultCreate(&Admin{}, &UserInfo{})
	t.NotError(err, "%s@%s", err, t.DriverName)

	insert := func(obj interface{}) {
		r, err := db.Insert(obj)
		t.NotError(err, "%s@%s", err, t.DriverName)
		cnt, err := r.RowsAffected()
		t.NotError(err, "%s@%s", err, t.DriverName).
			Equal(cnt, 1)
	}

	insert(&Group{
		Name: "group1",
		ID:   1,
	})

	insert(&Admin{
		Admin1: Admin1{Username: "username1", Password: "password1"},
		Email:  "email1",
		Group:  1,
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

	t.NotError(db.Select(u1))
	t.Equal(u1, &UserInfo{UID: 1, FirstName: "f1", LastName: "l1", Sex: "female"})

	t.NotError(db.Select(u2))
	t.Equal(u2, &UserInfo{UID: 2, FirstName: "f2", LastName: "l2", Sex: "male"})

	t.NotError(db.Select(a1))
	t.Equal(a1.Username, "username1")
}

func clearData(t *test.Test) {
	t.NotError(t.DB.MultDrop(&Admin{}, &User{}))
	t.NotError(t.DB.Drop(&Group{})) // admin 外键依赖 group
	t.NotError(t.DB.Drop(&UserInfo{}))
}

func TestDB_Update_error(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		initData(t)
		defer clearData(t)

		// 非结构体传入
		r, err := t.DB.Update(123)
		t.Error(err, "%s@%s", err, t.DriverName).Nil(r)
	})
}

func TestDB_LastInsertID(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		t.NotError(t.DB.Create(&User{}))
		defer func() {
			t.NotError(t.DB.Drop(&User{}))
		}()

		id, err := t.DB.LastInsertID(&User{Username: "1"})
		t.NotError(err).
			Equal(id, 1)

		id, err = t.DB.LastInsertID(&User{Username: "2"})
		t.NotError(err).
			Equal(id, 2)
	})
}

func TestDB_Update(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		initData(t)
		defer clearData(t)

		// update
		r, err := t.DB.Update(&UserInfo{
			UID:       1,
			FirstName: "firstName1",
			LastName:  "lastName1",
			Sex:       "sex1",
		})
		a.NotError(err)
		cnt, err := r.RowsAffected()
		t.NotError(err).
			Equal(1, cnt)

		r, err = t.DB.Update(&UserInfo{
			UID:       2,
			FirstName: "firstName2",
			LastName:  "lastName2",
			Sex:       "sex2",
		})
		t.NotError(err)
		cnt, err = r.RowsAffected()
		a.NotError(err).
			Equal(1, cnt)

		u1 := &UserInfo{UID: 1}
		t.NotError(t.DB.Select(u1))
		t.Equal(u1, &UserInfo{UID: 1, FirstName: "firstName1", LastName: "lastName1", Sex: "sex1"})

		u2 := &UserInfo{LastName: "lastName2", FirstName: "firstName2"}
		t.NotError(t.DB.Select(u2))
		t.Equal(u2, &UserInfo{UID: 2, FirstName: "firstName2", LastName: "lastName2", Sex: "sex2"})
	})
}

func TestDB_Delete(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		initData(t)
		defer clearData(t)

		// delete
		r, err := t.DB.Delete(&UserInfo{UID: 1})
		t.NotError(err)
		cnt, err := r.RowsAffected()
		t.NotError(err).
			Equal(cnt, 1)

		r, err = t.DB.Delete(
			&UserInfo{
				LastName:  "l2",
				FirstName: "f2",
			})
		t.NotError(err)
		cnt, err = r.RowsAffected()
		t.NotError(err).
			Equal(cnt, 1)

		r, err = t.DB.Delete(&Admin{Email: "email1"})
		t.NotError(err)
		cnt, err = r.RowsAffected()
		t.NotError(err).
			Equal(cnt, 1)

		hasCount(t.DB, t.Assertion, "user_info", 0)
		hasCount(t.DB, t.Assertion, "administrators", 0)

		// delete 并不会重置 ai 计数
		_, err = t.DB.Insert(&Admin{Group: 1, Email: "email1"})
		t.NotError(err)
		a1 := &Admin{Email: "email1"}
		t.NotError(t.DB.Select(a1))
		t.Equal(a1.ID, 2) // a1.ID为一个自增列,不会在delete中被重置
	})
}

func TestDB_Count(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		initData(t)
		defer clearData(t)

		// 单条件
		count, err := t.DB.Count(
			&UserInfo{
				UID: 1,
			},
		)
		t.NotError(err).
			Equal(1, count)

		// 无条件
		count, err = t.DB.Count(&UserInfo{})
		t.NotError(err).
			Equal(2, count)

		// 条件不存在
		count, err = t.DB.Count(
			&Admin{Email: "email1-1000"}, // 该条件不存在
		)
		t.NotError(err).
			Equal(0, count)
	})
}

func TestDB_Truncate(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		initData(t)
		defer clearData(t)

		hasCount(t.DB, t.Assertion, "administrators", 1)

		// truncate 之后，会重置 AI
		t.NotError(t.DB.Truncate(&Admin{}))
		hasCount(t.DB, t.Assertion, "administrators", 0)

		_, err := t.DB.Insert(&Admin{Group: 1, Email: "email1", Admin1: Admin1{Username: "u1"}})
		t.NotError(err)
		_, err = t.DB.Insert(&Admin{Group: 1, Email: "email2", Admin1: Admin1{Username: "u2"}})
		t.NotError(err)

		a1 := &Admin{Email: "email1"}
		t.NotError(t.DB.Select(a1))
		t.Equal(1, a1.ID)

		a2 := &Admin{Email: "email2"}
		t.NotError(t.DB.Select(a2))
		t.Equal(2, a2.ID)
	})
}

func TestDB_Drop(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		initData(t)
		defer clearData(t)

		t.NotError(t.DB.Drop(&UserInfo{}))
		t.NotError(t.DB.Drop(&Admin{}))
		r, err := t.DB.Insert(&Admin{})
		t.Error(err).Nil(r)
	})
}

func TestDB_Version(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		v, err := t.DB.Version()
		t.NotError(err).
			NotEmpty(v)
	})
}
