// SPDX-License-Identifier: MIT

package orm_test

import (
	"github.com/issue9/assert"
	"github.com/issue9/conv"

	"github.com/issue9/orm/v3/core"
	"github.com/issue9/orm/v3/fetch"
	"github.com/issue9/orm/v3/internal/test"
)

// Group 带有自增 ID 的普通表结构
type Group struct {
	ID      int64  `orm:"name(id);ai;unique(unique_groups_id)"`
	Name    string `orm:"name(name);len(500)"`
	Created int64  `orm:"name(created)"`
}

// Meta 指定表属性
func (g *Group) Meta() string {
	return "name(groups)"
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
}

// Meta 指定表属性
func (m *User) Meta() string {
	return "name(users)"
}

// Meta 指定表属性
func (m *UserInfo) Meta() string {
	return "check(user_info_chk_name,uid>0);mysql_engine(innodb);mysql_charset(utf8);name(user_info)"
}

// Admin 带自增和两个唯一约束
type Admin struct {
	User
	Email string `orm:"name(email);len(20);unique(unique_email)"`
	Group int64  `orm:"name(group);fk(fk_name,#groups,id,NO ACTION)"`
}

// Meta 指定表属性
func (m *Admin) Meta() string {
	return "check(admin_chk_name,{group}>0);mysql_engine(innodb);mysql_charset(utf8);name(administrators)"
}

// Account 带一个 OCC 字段
type Account struct {
	UID     int64 `orm:"name(uid);pk"`
	Account int64 `orm:"name(account)"`
	Version int64 `orm:"name(version);occ;default(0)"`
}

// Meta 指定表属性
func (m *Account) Meta() string {
	return "name(account)"
}

// table 表中是否存在 size 条记录，若不是，则触发 error
func hasCount(e core.Engine, a *assert.Assertion, table string, size int) {
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
func initData(t *test.Driver) {
	db := t.DB

	err := db.Create(&Group{})
	t.NotError(err, "%s@%s", err, t.DriverName)

	err = db.MultCreate(&Admin{}, &UserInfo{}, &Account{})
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
		User:  User{Username: "username1", Password: "password1"},
		Email: "email1",
		Group: 1,
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

func clearData(t *test.Driver) {
	t.NotError(t.DB.MultDrop(&Admin{}, &Account{}))
	t.NotError(t.DB.Drop(&Group{})) // admin 外键依赖 group
	t.NotError(t.DB.Drop(&UserInfo{}))
}
