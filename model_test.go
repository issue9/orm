// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm_test

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

// User 带自增和一个唯一约束
type User struct {
	ID       int    `orm:"name(id);ai;"`
	Username string `orm:"unique(unique_username);index(index_name);len(50)"`
	Password string `orm:"name(password);len(20)"`
	Regdate  int    `orm:"-"`
}

// Meta 指定表属性
func (m *User) Meta() string {
	return "mysql_engine(innodb);mysql_charset(utf8);name(users)"
}

// UserInfo 带主键和唯一约束(两个字段组成)
type UserInfo struct {
	UID       int    `orm:"name(uid);pk"`
	FirstName string `orm:"name(firstName);unique(unique_name);len(20)"`
	LastName  string `orm:"name(lastName);unique(unique_name);len(20)"`
	Sex       string `orm:"name(sex);default(male);len(6)"`
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
	Version int64 `orm:"name(version);occ(true);default(1)"`
}

// Meta 指定表属性
func (m *Account) Meta() string {
	return "name(account)"
}
