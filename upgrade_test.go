// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v3/internal/test"
)

type u1 struct {
	ID       int64  `orm:"name(id)"`
	Name     string `orm:"name(name);len(50);index(i_name)"`
	UserName string `orm:"name(username);len(50);unique(u_username)"`
	Created  int64  `orm:"name(created)"`
}

func (u *u1) Meta() string {
	return `name(upgrades);check(chk_id,id>0)`
}

type u2 struct {
	ID       int64  `orm:"name(id);unique(u_id)"`
	Name     string `orm:"name(name);len(50);index(index_name)"`
	UserName string `orm:"name(username);len(50);pk"`
	Modified int64  `orm:"name(modified);default(0)"`
}

func (u *u2) Meta() string {
	return `name(upgrades);check(chk_username,{username} IS NOT NULL)`
}

func TestUpgrader(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		t.NotError(t.DB.Create(&u1{}))

		defer func(n string) {
			t.NotError(t.DB.Drop(&u1{}))
		}(t.DriverName)

		u, err := t.DB.Upgrade(&u2{})
		t.NotError(err, "%s@%s", err, t.DriverName).
			NotNil(u)

		err = u.DropConstraint("u_username", "chk_id").
			DropColumns("created").
			DropIndex("i_name").
			AddConstraint("u_id", "chk_username").
			AddColumn("modified").
			AddIndex("index_name").
			Do()
		t.NotError(err, "%s@%s", err, t.DriverName)
	})
}
