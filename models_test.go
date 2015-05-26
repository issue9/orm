// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

var createdSQL = map[string]string{
	"bench": ` CREATE TABLE sqlite3_bench(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		pass TEXT
	)`,
	"users": `CREATE TABLE sqlite3_users(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		Username TEXT,
		password TEXT,
		CONSTRAINT unique_username UNIQUE(Username)
	)`,
	"user_info": `CREATE TABLE sqlite3_user_info(
		uid INTEGER PRIMARY KEY,
		firstName TEXT,
		lastName TEXT,
		sex TEXT DEFAULT male,
		CONSTRAINT unique_name UNIQUE(firstName,lastName)

	)`,
	"administrators": `CREATE TABLE sqlite3_administrators(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		Username TEXT,
		password TEXT,
		email TEXT,
		"group" INTEGER,
		CONSTRAINT unique_username UNIQUE(Username),
		CONSTRAINT unique_email UNIQUE(email)
	)`,
}

type bench struct {
	ID   int    `orm:"name(id);ai"`
	Name string `orm:"name(name)"`
	Pass string `orm:"name(pass)"`
}

func (b *bench) Meta() string {
	return "name(bench)"
}

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
	Group int    `orm:"name(group);fk(fk_name,table_group,id,NO ACTION,)"`
}

func (m *admin) Meta() string {
	return "check(chk_name,id>0);engine(innodb);charset(utf-8);name(administrators)"
}
