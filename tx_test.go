// SPDX-License-Identifier: MIT

package orm_test

import (
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/orm/v5/core"
	"github.com/issue9/orm/v5/internal/test"
)

func TestTx_InsertMany(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		tx, err := t.DB.Begin()
		a.NotError(err)
		a.NotError(tx.Create(&UserInfo{}))

		a.NotError(tx.InsertMany(10, &UserInfo{
			UID:       1,
			FirstName: "f1",
			LastName:  "l1",
		}))

		// 分批插入
		a.NotError(tx.InsertMany(2, []core.TableNamer{
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
		}...))

		// 单个元素插入
		t.NotError(tx.InsertMany(10,
			&UserInfo{
				UID:       7,
				FirstName: "f7",
				LastName:  "l7",
			}))

		t.NotError(tx.Commit())

		u1 := &UserInfo{UID: 1}
		t.NotError(t.DB.Select(u1))
		t.Equal(u1, &UserInfo{UID: 1, FirstName: "f1", LastName: "l1", Sex: "male"})

		u2 := &UserInfo{LastName: "l2", FirstName: "f2"}
		t.NotError(t.DB.Select(u2))
		t.Equal(u2, &UserInfo{UID: 2, FirstName: "f2", LastName: "l2", Sex: "male"})

		u3 := &UserInfo{UID: 3}
		t.NotError(t.DB.Select(u3))
		t.Equal(u3, &UserInfo{UID: 3, FirstName: "f3", LastName: "l3", Sex: "male"})

		u4 := &UserInfo{UID: 4}
		t.NotError(t.DB.Select(u4))
		t.Equal(u4, &UserInfo{UID: 4, FirstName: "f4", LastName: "l4", Sex: "male"})

		u5 := &UserInfo{UID: 5}
		t.NotError(t.DB.Select(u5))
		t.Equal(u5, &UserInfo{UID: 5, FirstName: "f5", LastName: "l5", Sex: "male"})

		u6 := &UserInfo{UID: 6}
		t.NotError(t.DB.Select(u6))
		t.Equal(u6, &UserInfo{UID: 6, FirstName: "f6", LastName: "l6", Sex: "male"})

		u7 := &UserInfo{UID: 7}
		t.NotError(t.DB.Select(u7))
		t.Equal(u7, &UserInfo{UID: 7, FirstName: "f7", LastName: "l7", Sex: "male"})

		t.NotError(t.DB.Drop(&UserInfo{}))
	})
}

func TestTx_LastInsertID(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		a.NotError(t.DB.Create(&User{}))
		defer func() {
			a.NotError(t.DB.Drop(&User{}))
		}()

		tx, err := t.DB.Begin()
		t.NotError(err)

		id, err := tx.LastInsertID(&User{Username: "1"})
		t.NotError(err).Equal(id, 1)

		id, err = tx.LastInsertID(&User{Username: "2"})
		t.NotError(err).Equal(id, 2)

		t.NotError(tx.Commit())
	})
}

func TestTx_Insert(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		t.NotError(t.DB.Create(&User{}))
		defer func() {
			t.NotError(t.DB.Drop(&User{}))
		}()

		tx, err := t.DB.Begin()
		t.NotError(err)

		_, err = tx.Insert(&User{
			ID:       1,
			Username: "u1",
		})
		t.NotError(err)

		_, err = tx.Insert(&User{
			ID:       2,
			Username: "u2",
		})
		a.NotError(err)

		_, err = tx.Insert(&User{
			ID:       3,
			Username: "u3",
		})
		t.NotError(err)

		t.NotError(tx.Commit())

		u1 := &User{ID: 1}
		t.NotError(t.DB.Select(u1))
		t.Equal(u1, &User{ID: 1, Username: "u1"})

		u3 := &User{ID: 3}
		t.NotError(t.DB.Select(u3))
		t.Equal(u3, &User{ID: 3, Username: "u3"})
	})
}

func TestTx_Update(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initData(t)
		defer clearData(t)

		tx, err := t.DB.Begin()
		t.NotError(err).NotNil(tx)

		// update
		_, err = tx.Update(&UserInfo{
			UID:       1,
			FirstName: "firstName1",
			LastName:  "lastName1",
			Sex:       "sex1",
		})
		t.NotError(err)
		_, err = tx.Update(&UserInfo{
			UID:       2,
			FirstName: "firstName2",
			LastName:  "lastName2",
			Sex:       "sex2",
		})
		t.NotError(err)
		t.NotError(tx.Commit())

		u1 := &UserInfo{UID: 1}
		t.NotError(t.DB.Select(u1))
		t.Equal(u1, &UserInfo{UID: 1, FirstName: "firstName1", LastName: "lastName1", Sex: "sex1"})

		u2 := &UserInfo{LastName: "lastName2", FirstName: "firstName2"}
		t.NotError(t.DB.Select(u2))
		t.Equal(u2, &UserInfo{UID: 2, FirstName: "firstName2", LastName: "lastName2", Sex: "sex2"})
	})
}

func TestTX(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		t.NotError(t.DB.Create(&User{}))
		t.NotError(t.DB.Create(&UserInfo{}))
		defer func() {
			t.NotError(t.DB.Drop(&User{}))
			t.NotError(t.DB.Drop(&UserInfo{}))
		}()

		// 回滚事务
		tx, err := t.DB.Begin()
		t.NotError(err).NotNil(tx)
		_, err = tx.Insert(&User{Username: "u1"})
		t.NotError(err)
		_, err = tx.Insert(&User{Username: "u2"})
		t.NotError(err)
		_, err = tx.Insert(&User{Username: "u3"})
		t.NotError(err)
		t.NotError(tx.Rollback())
		hasCount(t.DB, a, "users", 0) // tx 已经结束

		// 正常提交
		tx, err = t.DB.Begin()
		t.NotError(err).NotNil(tx)
		_, err = tx.Insert(&User{Username: "u1"})
		t.NotError(err)
		_, err = tx.Insert(&User{Username: "u2"})
		t.NotError(err)
		_, err = tx.Insert(&User{Username: "u3"})
		t.NotError(err)
		t.NotError(tx.Commit())
		hasCount(t.DB, a, "users", 3)
	})
}
