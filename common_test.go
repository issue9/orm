// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

// 公用的测试文件

import (
	"github.com/issue9/orm/core"
	"github.com/issue9/orm/dialect"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

const dbFile = "./test.db"

// 注册其它测试用例需要用到的dialect，
// dialect测试函数需要注意，在测试完成之后，请调用些函数还原必要的测试用dialect
func init() {
	if !core.IsRegisted("sqlite3") {
		if err := core.Register("sqlite3", &dialect.Sqlite3{}); err != nil {
			panic(err)
		}
	}

	if !core.IsRegisted("mysql") {
		if err := core.Register("mysql", &dialect.Mysql{}); err != nil {
			panic(err)
		}
	}
}

type Address struct {
	City   int    `orm:"name(city);unique(unique_address)"`
	Detail string `orm:"name(detail);unique(unique_address)"`
}

type User struct {
	Id      int    `orm:"name(id);ai"`
	Account string `orm:"name(account);unique(unique_account);len(20);"`
	Group   int    `orm:"name({group});len(11);default(1)"`

	Address
}

func (u *User) Meta() string {
	return "name(#user)"
}
