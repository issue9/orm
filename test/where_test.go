// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package test

import (
	"testing"

	"github.com/issue9/assert"
)

func TestWhere_Update_Delete_Select(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer func() {
		a.NotError(db.Drop(&admin{}, &user{}, &userInfo{}))
		a.NotError(db.Close())
		closeDB(a)
	}()

	a.NotError(db.Create(&userInfo{}))
	a.NotError(db.Insert(
		&userInfo{UID: 1, FirstName: "f1", LastName: "l1"},
		&userInfo{UID: 2, FirstName: "f2", LastName: "l2"},
		&userInfo{UID: 3, FirstName: "f3", LastName: "l3"},
	))

	// Where.Update
	err := db.Where("uid=?", 1).
		Table("#user_info").
		Update(true, map[string]interface{}{
		"firstName": "firstName1",
		"lastName":  "lastName1",
	})
	a.NotError(err)

	// Where.Update
	err = db.Where("{lastName}=?", "l2").
		And("{firstName}=?", "f2").
		Table("#user_info").
		Update(true, map[string]interface{}{
		"firstName": "firstName2",
		"lastName":  "lastName2",
	})
	a.NotError(err)

	// 验证修改

	// Where.SelectMap
	m, err := db.Where("uid<?", 4).
		Table("#user_info").
		Asc("uid").
		SelectMap(true, "uid", "firstName", "lastName")
	a.NotError(err).NotError(m)
	for _, item := range m {
		if v, ok := item["firstName"].([]byte); ok {
			item["firstName"] = string(v)
		}

		if v, ok := item["lastName"].([]byte); ok {
			item["lastName"] = string(v)
		}

	}
	a.Equal(m, []map[string]interface{}{
		map[string]interface{}{"uid": 1, "firstName": "firstName1", "lastName": "lastName1"},
		map[string]interface{}{"uid": 2, "firstName": "firstName2", "lastName": "lastName2"},
		map[string]interface{}{"uid": 3, "firstName": "f3", "lastName": "l3"},
	})

	// Where.Select
	objs := []*userInfo{&userInfo{}, &userInfo{}, &userInfo{}}
	err = db.Where("uid<?", 4).
		Table("#user_info").
		Asc("uid").
		Select(true, objs)
	a.NotError(err)
	a.Equal(objs, []*userInfo{
		&userInfo{UID: 1, FirstName: "firstName1", LastName: "lastName1", Sex: "male"},
		&userInfo{UID: 2, FirstName: "firstName2", LastName: "lastName2", Sex: "male"},
		&userInfo{UID: 3, FirstName: "f3", LastName: "l3", Sex: "male"},
	})

	// Where.Delete
	a.NotError(db.Where("uid=?", 3).Table("#user_info").Delete(true))

	// 确认Where.Delete()起作用
	hasCount(db, a, "user_info", 2)
}
