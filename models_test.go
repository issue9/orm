// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

type user struct {
	ID       int    `orm:"name(id);ai;"`
	Username string `orm:"unique(unique_username);index(index_name);len(50)"`
	Password string `orm:"name(password)"`
	Regdate  int    `orm:"-"`
}

func (m *user) Meta() string {
	return "check(chk_name,id>0);engine(innodb);charset(utf-8);name(users)"
}

type userInfo struct {
	UID       int    `orm:"name(uid);pk"`
	FirstName string `orm:"name(firstName);unique(unique_name)"`
	LastName  string `orm:"name(lastName);unique(unique_name)"`
	Sex       string `orm:"name(sex);default(male)"`
}

func (m *userInfo) Meta() string {
	return "check(chk_name,id>0);engine(innodb);charset(utf-8);name(user_info)"
}

type admin struct {
	user

	Email string `orm:"name(email);unique(unique_email)"`
	Group int    `orm:"name(group);fk(fk_name,table.group,id,NO ACTION,)"`
}

func (m *admin) Meta() string {
	return "check(chk_name,id>0);engine(innodb);charset(utf-8);name(administrators)"
}
