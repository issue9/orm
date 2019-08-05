// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/test"
)

func TestWhereStmt_Delete(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initData(t)
		defer clearData(t)

		// delete
		r, err := t.DB.Where("uid=?", 1).Delete(&UserInfo{})
		t.NotError(err)
		cnt, err := r.RowsAffected()
		t.NotError(err).
			Equal(cnt, 1)

		r, err = t.DB.Where("last_name=?", "l2").And("first_name=?", "f2").Delete(&UserInfo{})
		t.NotError(err)
		cnt, err = r.RowsAffected()
		t.NotError(err).
			Equal(cnt, 1)

		r, err = t.DB.Where("email=?", "email1").Delete(&Admin{Email: "e"})
		t.NotError(err)
		cnt, err = r.RowsAffected()
		t.NotError(err).
			Equal(cnt, 1)

		hasCount(t.DB, t.Assertion, "user_info", 0)
		hasCount(t.DB, t.Assertion, "administrators", 0)
	})
}

func TestWhereStmt_Update(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initData(t)
		defer clearData(t)

		r, err := t.DB.Where("last_name=?", "l2").
			And("first_name=?", "f2").
			Update(&UserInfo{
				FirstName: "firstName2",
				LastName:  "lastName2",
				Sex:       "sex2",
			})
		t.NotError(err)
		cnt, err := r.RowsAffected()
		t.NotError(err).
			Equal(cnt, 1)

		r, err = t.DB.Where("email=?", "email1").Update(&Admin{Email: "email1111"})
		t.NotError(err)
		cnt, err = r.RowsAffected()
		t.NotError(err).
			Equal(cnt, 1)

		u2 := &UserInfo{LastName: "lastName2", FirstName: "firstName2"}
		t.NotError(t.DB.Select(u2))
		t.Equal(u2, &UserInfo{UID: 2, FirstName: "firstName2", LastName: "lastName2", Sex: "sex2"})

		admin := &Admin{Email: "email1111"}
		t.NotError(t.DB.Select(admin))
		t.Equal(admin, &Admin{User: User{ID: 1, Username: "username1", Password: "password1"}, Email: "email1111", Group: 1})
	})
}

func TestWhereStmt_Select(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initData(t)
		defer clearData(t)

		u := &UserInfo{}
		cnt, err := t.DB.Where("uid>=?", 1).Select(true, u)
		a.NotError(err).
			Equal(cnt, 1).
			Equal(u.UID, 1).
			Equal(u.FirstName, "f1")

		us := make([]*UserInfo, 0, 0)
		cnt, err = t.DB.Where("uid>=?", 1).Select(true, &us)
		a.NotError(err).
			Equal(cnt, 2).
			Equal(us[0].UID, 1).
			Equal(us[0].FirstName, "f1").
			Equal(us[1].UID, 2).
			Equal(us[1].FirstName, "f2")

		type ui struct {
			UserInfo
		}

		items := []interface{}{
			&UserInfo{},
			&ui{},
		}
		cnt, err = t.DB.Where("uid>=?", 1).Select(true, &items)
		a.Error(err).Empty(cnt)
	})
}
