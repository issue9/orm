// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package orm_test

import (
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/orm/v5/core"
	"github.com/issue9/orm/v5/internal/test"
)

func TestPrefix_InsertMany(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		tx, err := t.DB.Begin()
		a.NotError(err)

		p := tx.Prefix("p1_")
		a.NotNil(p)

		a.NotError(p.Create(&UserInfo{}))

		defer func() {
			t.TB().Helper()
			t.NotError(p.Drop(&UserInfo{}))
		}()

		a.NotError(p.InsertMany(10, &UserInfo{
			UID:       1,
			FirstName: "f1",
			LastName:  "l1",
		}))

		// 分批插入
		a.NotError(p.InsertMany(2, []core.TableNamer{
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
		t.NotError(p.InsertMany(10,
			&UserInfo{
				UID:       7,
				FirstName: "f7",
				LastName:  "l7",
			}))

		t.NotError(tx.Commit())

		p = t.DB.Prefix("p1_")

		u1 := &UserInfo{UID: 1}
		found, err := p.Select(u1)
		t.NotError(err).True(found)
		t.Equal(u1, &UserInfo{UID: 1, FirstName: "f1", LastName: "l1", Sex: "male"})

		u2 := &UserInfo{LastName: "l2", FirstName: "f2"}
		found, err = p.Select(u2)
		t.NotError(err).True(found)
		t.Equal(u2, &UserInfo{UID: 2, FirstName: "f2", LastName: "l2", Sex: "male"})

		u3 := &UserInfo{UID: 3}
		found, err = p.Select(u3)
		t.NotError(err).True(found)
		t.Equal(u3, &UserInfo{UID: 3, FirstName: "f3", LastName: "l3", Sex: "male"})

		u4 := &UserInfo{UID: 4}
		found, err = p.Select(u4)
		t.NotError(err).True(found)
		t.Equal(u4, &UserInfo{UID: 4, FirstName: "f4", LastName: "l4", Sex: "male"})

		// 不带前经的 DB，找不到数据。
		u5 := &UserInfo{UID: 5}
		found, err = t.DB.Select(u5)
		t.Error(err).False(found)

		u6 := &UserInfo{UID: 6}
		found, err = p.Select(u6)
		t.NotError(err).True(found)
		t.Equal(u6, &UserInfo{UID: 6, FirstName: "f6", LastName: "l6", Sex: "male"})

		u7 := &UserInfo{UID: 7}
		found, err = p.Select(u7)
		t.NotError(err).True(found)
		t.Equal(u7, &UserInfo{UID: 7, FirstName: "f7", LastName: "l7", Sex: "male"})
	})
}

func TestPrefix_LastInsertID(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		p1 := t.DB.Prefix("p1_")
		p2 := t.DB.Prefix("p2_")

		a.NotError(p1.Create(&User{}))
		a.NotError(p2.Create(&User{}))
		defer func() {
			a.NotError(p1.Drop(&User{}))
			a.NotError(p2.Drop(&User{}))
		}()

		ids1, err := p1.LastInsertID(&User{Username: "1"})
		t.NotError(err).Equal(ids1, 1)

		ids2, err := p2.LastInsertID(&User{Username: "1"})
		t.NotError(err).Equal(ids2, 1)

		ids1, err = p1.LastInsertID(&User{Username: "2"})
		t.NotError(err).Equal(ids1, 2)

		ids2, err = p2.LastInsertID(&User{Username: "2"})
		t.NotError(err).Equal(ids2, 2)
	})
}
