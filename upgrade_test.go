// SPDX-License-Identifier: MIT

package orm_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v3/core"
	"github.com/issue9/orm/v3/internal/test"
)

type u2 struct {
	ID       int64  `orm:"name(id);unique(u_id)"`
	Name     string `orm:"name(name);len(50);index(index_name)"`
	UserName string `orm:"name(username);len(50);pk"`
	Modified int64  `orm:"name(modified);default(0)"`
}

func (u *u2) TableName() string { return "#upgrades" }

func (u *u2) Meta(m *core.Model) error {
	return m.NewCheck("chk_username", "{username} IS NOT NULL")
}

func TestUpgrader(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
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

		//t.DB.ccl

		u, err := t.DB.Upgrade(&u2{})
		t.NotError(err, "%s@%s", err, t.DriverName).
			NotNil(u)

		err = u.DropConstraint("u_username", "chk_id").
			DropColumn("created").
			DropIndex("i_name").
			AddConstraint("u_id", "chk_username").
			AddColumn("modified").
			AddIndex("index_name").
			Do()
		t.NotError(err, "%s@%s", err, t.DriverName)
	})
}
