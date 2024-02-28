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

type u2 struct {
	ID       int64  `orm:"name(id);unique(u_id)"`
	Name     string `orm:"name(name);len(50);index(index_name)"`
	UserName string `orm:"name(username);len(50);pk"`
	Modified int64  `orm:"name(modified);default(0)"`
	Created  string `orm:"name(created);nullable"`
}

func (u *u2) TableName() string { return "upgrades" }

func (u *u2) ApplyModel(m *core.Model) error {
	return m.NewCheck("chk_username", "{username} IS NOT NULL")
}

func TestUpgrader(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a)

	suite.Run(func(t *test.Driver) {
		sql := t.DB.SQLBuilder().CreateTable().
			Column("id", core.Int64, false, false, false, nil).
			Column("name", core.String, false, false, false, nil, 50).
			Column("username", core.String, false, false, false, nil, 50).
			Column("created", core.Int64, false, false, false, nil).
			Index(core.IndexDefault, "i_name", "name").
			Unique("u_username", "username").
			Check("chk_id", "id>0").
			Table((&u2{}).TableName())
		t.NotError(sql.Exec())

		defer func(n string) {
			t.NotError(t.DB.Drop(&u2{}))
		}(t.DriverName)

		u, err := t.DB.Upgrade(&u2{})
		t.NotError(err, "%s@%s", err, t.DriverName).
			NotNil(u)

		err = u.DropConstraint("u_username", "chk_id").
			DropIndex("i_name").
			DropColumn("created").
			AddColumn("modified").
			AddConstraint("u_id", "chk_username", "u2_pk").
			AddIndex("index_name").
			AddColumn("created"). // 同名不同类型
			Do()
		t.NotError(err, "%s@%s", err, t.DriverName)
	})
}
