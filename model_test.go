// SPDX-License-Identifier: MIT

package orm_test

import (
	"database/sql"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/conv"

	"github.com/issue9/orm/v5/core"
	"github.com/issue9/orm/v5/fetch"
	"github.com/issue9/orm/v5/internal/test"
)

// Group 带有自增 ID 的普通表结构
type Group struct {
	ID      sql.NullInt64 `orm:"name(id);ai;unique(unique_groups_id)"`
	Name    string        `orm:"name(name);len(500)"`
	Created time.Time     `orm:"name(created)"`
	Any     any           `orm:"name(any)"` // any
}

func (g *Group) TableName() string { return "groups" }

func (g *Group) BeforeInsert() error {
	g.Created = time.Now()
	return nil
}

type User struct {
	ID       int    `orm:"name(id);ai;"`
	Username string `orm:"unique(unique_admin_username);index(index_admin_name);len(50)"`
	Password string `orm:"name(password);len(20)"`
	Regdate  int    `orm:"-"`
}

// UserInfo 带主键和唯一约束(两个字段组成)
type UserInfo struct {
	UID       int    `orm:"name(uid);pk"`
	FirstName string `orm:"name(first_name);unique(unique_name);len(20)"`
	LastName  string `orm:"name(last_name);unique(unique_name);len(20)"`
	Sex       string `orm:"name(sex);default(male);len(6)"`
	Any       any    `orm:"name(any);nullable"`
}

func (u *User) TableName() string { return "users" }

func (u *UserInfo) TableName() string { return "user_info" }

func (u *UserInfo) ApplyModel(m *core.Model) error {
	m.Options["mysql_engine"] = []string{"innodb"}
	m.Options["mysql_charset"] = []string{"utf8"}
	return m.NewCheck("user_info_chk_name", "uid>0")
}

// Admin 带自增和两个唯一约束
type Admin struct {
	User
	Email string `orm:"name(email);len(20);unique(unique_email)"`
	Group int64  `orm:"name(group);fk(fk_name,groups,id,NO ACTION)"`
}

func (a *Admin) TableName() string { return "administrators" }

func (a *Admin) ApplyModel(m *core.Model) error {
	m.Options["mysql_engine"] = []string{"innodb"}
	m.Options["mysql_charset"] = []string{"utf8"}
	return m.NewCheck("admin_chk_name", "{group}>0")
}

// Account 带一个 OCC 字段
type Account struct {
	UID     int64 `orm:"name(uid);pk"`
	Account int64 `orm:"name(account)"`
	Version int64 `orm:"name(version);occ;default(0)"`
}

func (a *Account) TableName() string { return "account" }

// table 表中是否存在 size 条记录，若不是，则触发 error
func hasCount(e core.Engine, a *assert.Assertion, table string, size int) {
	rows, err := e.Query("SELECT COUNT(*) as cnt FROM " + table)
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
func initData(t *test.Driver) {
	db := t.DB

	err := db.Create(&Group{})
	t.NotError(err, "%s@%s", err, t.DriverName)

	err = db.Create(&Admin{})
	t.NotError(err, "%s@%s", err, t.DriverName)
	err = db.Create(&UserInfo{})
	t.NotError(err, "%s@%s", err, t.DriverName)
	err = db.Create(&Account{})
	t.NotError(err, "%s@%s", err, t.DriverName)

	insert := func(obj core.TableNamer) {
		t.TB().Helper()

		r, err := db.Insert(obj)
		t.NotError(err, "%s@%s", err, t.DriverName)
		cnt, err := r.RowsAffected()
		t.NotError(err, "%s@%s", err, t.DriverName).
			Equal(cnt, 1)
	}

	insert(&Group{
		Name: "group1",
		ID:   sql.NullInt64{Int64: 1, Valid: true},
		Any:  "attr",
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
		Any:       55,
	})
	insert(&UserInfo{ // sex 使用默认值，any 采用 nullable
		UID:       2,
		FirstName: "f2",
		LastName:  "l2",
	})

	// select
	g1 := &Group{ID: sql.NullInt64{Int64: 1, Valid: true}}
	u1 := &UserInfo{UID: 1}
	u2 := &UserInfo{LastName: "l2", FirstName: "f2"}
	a1 := &Admin{Email: "email1"}

	found, err := t.DB.Select(g1)
	t.NotError(err).True(found).
		True(g1.Created.After(time.Now().Add(-24*time.Hour)), "g1.Created 应该有大于昨天的时间，页实际值为 %s", g1.Created).
		Equal(g1.Name, "group1").
		Equal(g1.Any, "attr")

	found, err = t.DB.Select(u1)
	t.NotError(err).True(found).
		Equal(1, u1.UID).
		Equal("f1", u1.FirstName).
		Equal("l1", u1.LastName).
		Equal("female", u1.Sex)
	if t.DriverName == "mysql" {
		t.Equal(u1.Any, []byte("55"))
	} else {
		t.Equal(u1.Any, "55")
	}

	found, err = t.DB.Select(u2)
	t.NotError(err).True(found).
		Equal(u2, &UserInfo{UID: 2, FirstName: "f2", LastName: "l2", Sex: "male", Any: nil})

	found, err = t.DB.Select(a1)
	t.NotError(err).True(found).
		Equal(a1.Username, "username1")
}

func clearData(t *test.Driver) {
	t.NotError(t.DB.Drop(&Admin{}))
	t.NotError(t.DB.Drop(&Account{}))
	t.NotError(t.DB.Drop(&Group{})) // admin 外键依赖 group
	t.NotError(t.DB.Drop(&UserInfo{}))
}
