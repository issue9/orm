// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"github.com/issue9/orm/core"
)

///////////////////////////////////////////////////////////
///////////////////////// User1
///////////////////////////////////////////////////////////

type User1 struct {
	Id       int    `orm:"name(id);ai"`
	Username string `orm:"name(username);unique(unique_name);len(20)"`
	Group    string `orm:"name({group});len(20)"`
}

func (u *User1) Meta() string {
	return "name(user)"
}

var _ core.Metaer = &User1{}

///////////////////////////////////////////////////////////
///////////////////////// User2
///////////////////////////////////////////////////////////

// 相对于User1，将Group的类型从string改为int
type User2 struct {
	Id       int    `orm:"name(id);ai"`
	Username string `orm:"name(username);unique(unique_name);len(20)"`
	Group    int    `orm:"name({group})"`
}

func (u *User2) Meta() string {
	return "name(user)"
}

var _ core.Metaer = &User2{}

///////////////////////////////////////////////////////////
///////////////////////// User3
///////////////////////////////////////////////////////////

// 相对于User2，id的属性从ai变成了普通的PK
type User3 struct {
	Id       int    `orm:"name(id);pk"`
	Username string `orm:"name(username);unique(unique_name);len(20)"`
	Group    int    `orm:"name({group})"`
}

func (u *User3) Meta() string {
	return "name(user)"
}

var _ core.Metaer = &User3{}

///////////////////////////////////////////////////////////
///////////////////////// User4
///////////////////////////////////////////////////////////

// 相对于User3，字段名做了些增减
type User4 struct {
	Id      int    `orm:"name(id);pk"`
	Account string `orm:"name(account);unique(unique_name);len(20)"`
	Group   int    `orm:"name({group})"`
}

func (u *User4) Meta() string {
	return "name(user)"
}

var _ core.Metaer = &User4{}

///////////////////////////////////////////////////////////
///////////////////////// User5
///////////////////////////////////////////////////////////

// 相对于User4，使用了联合唯一索引
type User5 struct {
	Id      int    `orm:"name(id);pk"`
	Account string `orm:"name(account);unique(unique_name);len(20)"`
	Group   string `orm:"name({group});unique(unique_name);len(10)"`
}

func (u *User5) Meta() string {
	return "name(user)"
}

var _ core.Metaer = &User5{}
