// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package orm_test

import (
	"database/sql"
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/orm/v6"
	"github.com/issue9/orm/v6/internal/test"
)

func TestTx_InsertMany(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		tx, err := t.DB.Begin()
		a.NotError(err).NotNil(tx)

		a.NotError(tx.Create(&UserInfo{}))

		p1 := tx.NewEngine("p1_")
		a.NotError(p1.Create(&UserInfo{}))

		defer func() {
			t.TB().Helper()
			t.NotError(t.DB.Drop(&UserInfo{})).
				NotError(t.DB.New("p1_").Drop(&UserInfo{}))
		}()

		a.NotError(tx.InsertMany(10, &UserInfo{
			UID:       1,
			FirstName: "f1",
			LastName:  "l1",
		})).NotError(p1.InsertMany(10, &UserInfo{
			UID:       1,
			FirstName: "f1",
			LastName:  "l1",
		}))

		// 分批插入
		a.NotError(tx.InsertMany(2, []orm.TableNamer{
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
		found, err := t.DB.Select(u1)
		t.NotError(err).True(found)
		t.Equal(u1, &UserInfo{UID: 1, FirstName: "f1", LastName: "l1", Sex: "male"})

		u2 := &UserInfo{LastName: "l2", FirstName: "f2"}
		found, err = t.DB.Select(u2)
		t.NotError(err).True(found).
			Equal(u2, &UserInfo{UID: 2, FirstName: "f2", LastName: "l2", Sex: "male"})

		u3 := &UserInfo{UID: 3}
		found, err = t.DB.Select(u3)
		t.NotError(err).True(found).
			Equal(u3, &UserInfo{UID: 3, FirstName: "f3", LastName: "l3", Sex: "male"})

		u4 := &UserInfo{UID: 4}
		found, err = t.DB.Select(u4)
		t.NotError(err).True(found).
			Equal(u4, &UserInfo{UID: 4, FirstName: "f4", LastName: "l4", Sex: "male"})

		u5 := &UserInfo{UID: 5}
		found, err = t.DB.Select(u5)
		t.NotError(err).True(found).
			Equal(u5, &UserInfo{UID: 5, FirstName: "f5", LastName: "l5", Sex: "male"})

		u6 := &UserInfo{UID: 6}
		found, err = t.DB.Select(u6)
		t.NotError(err).True(found).
			Equal(u6, &UserInfo{UID: 6, FirstName: "f6", LastName: "l6", Sex: "male"})

		u7 := &UserInfo{UID: 7}
		found, err = t.DB.Select(u7)
		t.NotError(err).True(found).
			Equal(u7, &UserInfo{UID: 7, FirstName: "f7", LastName: "l7", Sex: "male"})

		p1db := t.DB.New("p1_")

		pu1 := &UserInfo{UID: 1}
		found, err = p1db.Select(pu1)
		t.NotError(err).True(found).
			Equal(pu1, &UserInfo{UID: 1, FirstName: "f1", LastName: "l1", Sex: "male"})

		pu7 := &UserInfo{UID: 7}
		found, err = p1db.Select(pu7)
		t.NotError(err).False(found)
	})
}

func TestTx_LastInsertID(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
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
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		t.NotError(t.DB.Create(&User{}))
		defer func() {
			t.NotError(t.DB.Drop(&User{}))
		}()

		tx, err := t.DB.Begin()
		t.NotError(err)

		_, err = tx.Insert(&User{
			Username: "u1",
		})
		t.NotError(err)

		_, err = tx.Insert(&User{
			Username: "u2",
		})
		a.NotError(err)

		_, err = tx.Insert(&User{
			Username: "u3",
		})
		t.NotError(err)

		t.NotError(tx.Commit())

		u1 := &User{ID: 1}
		found, err := t.DB.Select(u1)
		t.NotError(err).True(found)
		t.Equal(u1, &User{ID: 1, Username: "u1"})

		u3 := &User{ID: 3}
		found, err = t.DB.Select(u3)
		t.NotError(err).True(found)
		t.Equal(u3, &User{ID: 3, Username: "u3"})
	})
}

func TestTx_Update(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
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
		found, err := t.DB.Select(u1)
		t.NotError(err).True(found)
		t.Equal(1, u1.UID).
			Equal("firstName1", u1.FirstName).
			Equal("lastName1", u1.LastName).
			Equal("sex1", u1.Sex)
		if t.DriverName == "mysql" {
			t.Equal(u1.Any, []byte("55"))
		} else {
			t.Equal(u1.Any, "55")
		}

		u2 := &UserInfo{LastName: "lastName2", FirstName: "firstName2"}
		found, err = t.DB.Select(u2)
		t.NotError(err).True(found)
		t.Equal(u2, &UserInfo{UID: 2, FirstName: "firstName2", LastName: "lastName2", Sex: "sex2"})
	})
}

func TestTx_Save(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		initData(t)
		defer clearData(t)

		tx, err := t.DB.Begin()
		t.NotError(err).NotNil(tx)

		// 指定的行存在，采用 update
		cnt, err := t.DB.SQLBuilder().Select().Column("count(*) AS cnt").From(orm.TableName(&Group{})).QueryInt("cnt")
		a.NotError(err).Equal(cnt, 1)
		lastid, isnew, err := tx.Save(&Group{ID: sql.NullInt64{Valid: true, Int64: 1}, Name: "save"})
		a.NotError(err).False(isnew).Zero(lastid).
			NotError(tx.Commit())
		cnt, err = t.DB.SQLBuilder().Select().Column("count(*) AS cnt").From(orm.TableName(&Group{})).QueryInt("cnt")
		a.NotError(err).Equal(cnt, 1) // 没有增加行数

		// insert
		tx, err = t.DB.Begin()
		t.NotError(err).NotNil(tx)
		lastid, isnew, err = t.DB.Save(&Group{ID: sql.NullInt64{Valid: true, Int64: 2}, Name: "save"})
		a.Error(err)

		// insert
		tx, err = t.DB.Begin()
		t.NotError(err).NotNil(tx)
		lastid, isnew, err = t.DB.Save(&Group{Name: "save"})
		a.NotError(err).True(isnew).Equal(lastid, 2)
		cnt, err = t.DB.SQLBuilder().Select().Column("count(*) AS cnt").From(orm.TableName(&Group{})).QueryInt("cnt")
		a.NotError(err).Equal(cnt, 2).
			NotError(tx.Commit())
	})
}

func TestTX(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		t.NotError(t.DB.Create(&User{})).
			NotError(t.DB.Create(&UserInfo{}))
		defer func() {
			t.NotError(t.DB.Drop(&User{})).
				NotError(t.DB.Drop(&UserInfo{}))
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
