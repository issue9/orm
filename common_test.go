// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

// 公用的测试文件

import (
	"fmt"
	"os"

	"github.com/issue9/assert"
	"github.com/issue9/orm/core"
	"github.com/issue9/orm/dialect"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

const dbFile = "./test.db"

// 关闭sqlite3的数据库文件
func closeDBFile(a *assert.Assertion) {
	_, err := os.Stat(dbFile)
	if err == nil || os.IsExist(err) {
		a.NotError(os.Remove(dbFile))
		return
	}

	if os.IsNotExist(err) {
		return
	}

	a.NotError(err)
}

// 注册其它测试用例需要用到的dialect，
// dialect测试函数需要注意，在测试完成之后，请调用些函数还原必要的测试用dialect
func init() {
	if !core.IsRegisted("sqlite3") {
		if err := core.Register("sqlite3", &dialect.Sqlite3{}); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if !core.IsRegisted("mysql") {
		if err := core.Register("mysql", &dialect.Mysql{}); err != nil {
			fmt.Println(err)
			os.Exit(1)
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
