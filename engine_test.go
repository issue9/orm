// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"os"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/conv"
	"github.com/issue9/orm/fetch"
)

func TestEngine(t *testing.T) {
	a := assert.New(t)
	e, err := New("sqlite3", dbFile, "main", "main_")
	a.NotError(err).NotNil(e)

	testDelete(a, e)
	testUpdate(a, e)
}

// 检测#user表中的记录数据是否和size相同。
func engineChkCount(a *assert.Assertion, engine *Engine, size int, tableName string) {
	rows, err := engine.Query("SELECT count(*) AS c FROM "+tableName, nil)
	a.NotError(err).NotNil(rows)
	defer rows.Close()

	ret, err := fetch.Column(true, "c", rows)
	a.NotError(err).NotNil(ret)

	count, err := conv.Int(ret[0])
	a.NotError(err)
	a.Equal(count, size)
}

// 测试Create和Insert两个函数，顺便可以用作初始化数据库。
func testCreateInsert(a *assert.Assertion, e *Engine) {
	// Create
	a.NotError(e.Create(&User{}))
	engineChkCount(a, e, 0, "#user")

	// Drop
	a.NotError(e.Drop("#user"))

	// 创建多个表
	a.NotError(e.Create(&User{}, &Address{}))
	engineChkCount(a, e, 0, "#user")
	engineChkCount(a, e, 0, "Address")

	// 插入一条数据
	a.NotError(e.Insert(&User{Address: Address{City: 1, Detail: "#1"}, Group: 1, Account: "admin1"}))
	engineChkCount(a, e, 1, "#user")

	// 插入多条数据
	a.NotError(e.Insert(
		[]*User{
			&User{Address: Address{City: 2, Detail: "#2"}, Account: "admin2"},
			&User{Address: Address{City: 3, Detail: "#3"}, Account: "admin3"},
		},
	))
	engineChkCount(a, e, 3, "#user")
	engineChkCount(a, e, 0, "Address")

	// 插入一条非唯一数据，联合唯一约束(unique_address)不成立
	a.Error(e.Insert(&User{Address: Address{City: 1, Detail: "#1"}, Group: 5, Account: "admin4"}))

	// 插入一条非唯一数据，唯一约束(account)不成立
	a.Error(e.Insert(&User{Address: Address{City: 5, Detail: "#1"}, Group: 1, Account: "admin3"}))
}

// 测试Truncate和Drop函数，顺便用作清除数据库内容的工具。
func testTruncateDrop(a *assert.Assertion, e *Engine) {
	a.NotError(e.Truncate("#user"))
	engineChkCount(a, e, 0, "#user")

	// 最后清除所有的数据，方便其它测试
	a.NotError(e.Drop("#user"))
	a.NotError(e.Drop("Address"))
}

// 关闭数据库文件
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

func testUpdate(a *assert.Assertion, e *Engine) {
	testCreateInsert(a, e)
	defer testTruncateDrop(a, e)

	// 检测user表中某个id值的account字段是否与wont相等，
	chkAccount := func(a *assert.Assertion, engine *Engine, id int, wont string) {
		rows, err := engine.Query("SELECT * FROM #user WHERE id=@id LIMIT 1", map[string]interface{}{"id": id})
		a.NotError(err).NotNil(rows)
		defer rows.Close()

		ret, err := fetch.Column(true, "account", rows)
		a.NotError(err).Equal(ret[0], wont)
	}

	// 一条数据
	a.NotError(e.Update(&User{Id: 1, Account: "account1"}))
	chkAccount(a, e, 1, "account1")

	// 多条数据
	a.NotError(e.Update([]*User{
		&User{Id: 2, Account: "account2"},
		&User{Id: 3, Account: "account3"},
	}))
	chkAccount(a, e, 2, "account2")
	chkAccount(a, e, 3, "account3")
}

func testDelete(a *assert.Assertion, e *Engine) {
	testCreateInsert(a, e)
	defer testTruncateDrop(a, e)

	// 删除一条记录
	a.NotError(e.Delete(&User{Account: "admin1"}))
	engineChkCount(a, e, 2, "#user")

	// 删除两条记录
	a.NotError(e.Delete([]*User{
		&User{Id: 2},
		&User{Address: Address{City: 3, Detail: "#3"}},
	}))
	engineChkCount(a, e, 0, "#user")
}
