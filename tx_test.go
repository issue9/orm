// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/conv"
	"github.com/issue9/orm/fetch"
)

func TestTx(t *testing.T) {
	a := assert.New(t)
	e, err := New("sqlite3", dbFile, "main", "main_")
	a.NotError(err).NotNil(e)
	defer func() {
		a.NotError(e.Close())
		closeDBFile(a)
	}()

	txUpdate(a, e)
	txDelete(a, e)
}

func txUpdate(a *assert.Assertion, e *Engine) {
	txCreateInsert(a, e)
	defer txTruncateDrop(a, e)

	// 检测user表中某个id值的account字段是否与wont相等，
	chkAccount := func(a *assert.Assertion, engine *Engine, id int, wont string) {
		rows, err := engine.Query("SELECT * FROM #user WHERE id=@id LIMIT 1", map[string]interface{}{"id": id})
		a.NotError(err).NotNil(rows)
		defer rows.Close()

		ret, err := fetch.Column(true, "account", rows)
		a.NotError(err).Equal(ret[0], wont)
	}

	// 一条数据，回滚
	tx, err := e.Begin()
	a.NotError(err).NotNil(tx)
	a.NotError(tx.Update(&User{Id: 1, Account: "account1"}))
	a.NotError(tx.Rollback())
	chkAccount(a, e, 1, "admin1")

	// 一条数据，提交
	tx, err = e.Begin()
	a.NotError(err).NotNil(tx)
	a.NotError(tx.Update(&User{Id: 1, Account: "account1"}))
	a.NotError(tx.Commit())
	chkAccount(a, e, 1, "account1")

	// 多条数据，回滚
	tx, err = e.Begin()
	a.NotError(err).NotNil(tx)
	a.NotError(tx.Update([]*User{
		&User{Id: 2, Account: "account2"},
		&User{Id: 3, Account: "account3"},
	}))
	a.NotError(tx.Rollback())
	chkAccount(a, e, 2, "admin2")
	chkAccount(a, e, 3, "admin3")

	// 多条数据，提交
	tx, err = e.Begin()
	a.NotError(err).NotNil(tx)
	a.NotError(tx.Update([]*User{
		&User{Id: 2, Account: "account2"},
		&User{Id: 3, Account: "account3"},
	}))
	a.NotError(tx.Commit())
	chkAccount(a, e, 2, "account2")
	chkAccount(a, e, 3, "account3")
}

func txDelete(a *assert.Assertion, e *Engine) {
	txCreateInsert(a, e)
	defer txTruncateDrop(a, e)

	// 删除一条记录,回滚
	tx, err := e.Begin()
	a.NotError(err).NotNil(tx)
	a.NotError(tx.Delete(&User{Account: "admin1"}))
	a.NotError(tx.Rollback())
	txChkCount(a, e, 4, "#user", "删除一条记录,回滚")

	// 删除一条记录，提交
	tx, err = e.Begin()
	a.NotError(err).NotNil(tx)
	a.NotError(tx.Delete(&User{Account: "admin1"}))
	a.NotError(tx.Commit())
	txChkCount(a, e, 3, "#user", "删除一条记录，提交")

	// 删除两条记录，回滚
	tx, err = e.Begin()
	a.NotError(err).NotNil(tx)
	a.NotError(tx.Delete([]*User{
		&User{Id: 2},
		&User{Address: Address{City: 3, Detail: "#3"}},
	}))
	a.NotError(tx.Rollback())
	txChkCount(a, e, 3, "#user", "删除两条记录，回滚")

	// 删除两条记录，提交
	tx, err = e.Begin()
	a.NotError(err).NotNil(tx)
	a.NotError(e.Delete([]*User{
		&User{Id: 2},
		&User{Address: Address{City: 3, Detail: "#3"}},
	}))
	a.NotError(tx.Commit())
	txChkCount(a, e, 1, "#user", "删除两条记录，提交")
}

// 创建和插入数据，顺便用作初始化数据库用。
func txCreateInsert(a *assert.Assertion, e *Engine) {
	// 创建表
	tx, err := e.Begin()
	a.NotError(err).NotNil(tx)
	a.NotError(tx.Create(&User{}))
	a.NotError(tx.Commit())
	txChkCount(a, e, 0, "#user", "创建表结构")

	// 插入数据
	users := []*User{
		&User{Address: Address{City: 1, Detail: "#1"}, Account: "admin1"},
		&User{Address: Address{City: 2, Detail: "#2"}, Account: "admin2"},
		&User{Address: Address{City: 3, Detail: "#3"}, Account: "admin3"},
		&User{Address: Address{City: 4, Detail: "#4"}, Account: "admin4"},
	}

	// 回滚
	tx, err = e.Begin()
	a.NotError(err).NotNil(tx)
	a.NotError(tx.Insert(users))
	a.NotError(tx.Rollback())
	txChkCount(a, e, 0, "#user", "插入数据，回滚")

	// 事务插入数据
	tx, err = e.Begin()
	a.NotError(err).NotNil(tx)
	a.NotError(tx.Insert(users))
	a.NotError(tx.Commit())
	txChkCount(a, e, 4, "#user", "插入数据，提交")
}

// 测试Truncate和Drop函数，顺便用作清除数据库内容的工具。
func txTruncateDrop(a *assert.Assertion, e *Engine) {
	tx, err := e.Begin()
	a.NotError(err).NotNil(tx)
	a.NotError(tx.Truncate("#user"))
	a.NotError(tx.Commit())
	txChkCount(a, e, 0, "#user", "清除所有数据，提交")

	// 最后清除所有的数据，方便其它测试
	tx, err = e.Begin()
	a.NotError(err).NotNil(tx)
	a.NotError(tx.Drop("#user"))
	a.NotError(tx.Commit())
}

// 检测#user表中的记录数据是否和size相同。
func txChkCount(a *assert.Assertion, e *Engine, size int, tableName, name string) {
	t, err := e.Begin()
	a.NotError(err).NotNil(t)
	a.True(e.DB() == t.DB(), "e.DB() != t.DB()")

	rows, err := t.Query("SELECT count(*) AS c FROM "+tableName, nil)
	a.NotError(err).NotNil(rows).NotError(rows.Err())
	a.NotError(t.Commit())
	defer rows.Close()

	// 导出数据
	ret, err := fetch.Column(true, "c", rows)
	a.NotError(err).NotNil(ret)

	count, err := conv.Int(ret[0])
	a.NotError(err).Equal(count, size, "在[%v]处发生错误：实际值=[%v];预想值=[%v]", name, count, size)
}
