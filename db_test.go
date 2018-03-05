// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"os"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/conv"
	"github.com/issue9/orm/core"
	"github.com/issue9/orm/dialect"
	"github.com/issue9/orm/fetch"
	"github.com/issue9/orm/internal/modeltest"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var _ core.Engine = &DB{}

var (
	// 通过修改此值来确定使用哪个数据库驱动来测试
	// 若需要其它两种数据库测试，需要先在创建相应的数据库
	driver = "sqlite3"

	prefix = "prefix_"
	dsn    string
	d      core.Dialect
)

// 销毁数据库。默认仅对 sqlite3 启作用，删除该数据库文件。
func closeDB(a *assert.Assertion) {
	if driver != "sqlite3" {
		return
	}

	if _, err := os.Stat(dsn); err == nil || os.IsExist(err) {
		a.NotError(os.Remove(dsn))
	}
}

func newDB(a *assert.Assertion) *DB {
	switch driver {
	case "mysql":
		dsn = "root@/orm_test?charset=utf8"
		d = dialect.Mysql()
	case "sqlite3":
		dsn = "./orm_test.db"
		d = dialect.Sqlite3()
	case "postgres":
		dsn = "user=caixw dbname=orm_test sslmode=disable"
		d = dialect.Postgres()
	default:
		panic("仅支持mysql,sqlite3,postgres三种数据库测试")
	}

	db, err := NewDB(driver, dsn, prefix, d)
	a.NotError(err).NotNil(db)
	return db
}

// table表中是否存在 size 条记录，若不是，则触发 error
func hasCount(db core.Engine, a *assert.Assertion, table string, size int) {
	rows, err := db.Query("SELECT COUNT(*) as cnt FROM #" + table)
	a.NotError(err).NotNil(rows)
	defer func() {
		a.NotError(rows.Close())
	}()

	data, err := fetch.Map(true, rows)
	a.NotError(err).NotNil(data)
	a.Equal(conv.MustInt(data[0]["cnt"], -1), size)
}

func TestNewDB(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer func() {
		a.NotError(db.Close())
	}()

	a.Equal(db.tablePrefix, prefix)
	a.NotNil(db.StdDB()).NotNil(db.Dialect())
}

// 初始化测试数据，同时可当作 DB.Inert 的测试
// 清空其它数据，初始化成原始的测试数据
func initData(db *DB, a *assert.Assertion) {
	a.NotError(db.Drop(&modeltest.User{}))
	a.NotError(db.Drop(&modeltest.Admin{}))
	a.NotError(db.Drop(&modeltest.Group{})) // admin 外键依赖 group
	a.NotError(db.Drop(&modeltest.UserInfo{}))
	a.NotError(db.Create(&modeltest.Group{}))
	a.NotError(db.Create(&modeltest.Admin{}))
	a.NotError(db.Create(&modeltest.UserInfo{}))

	insert := func(obj interface{}) {
		r, err := db.Insert(obj)
		a.NotError(err)
		cnt, err := r.RowsAffected()
		a.NotError(err).Equal(cnt, 1)
	}

	insert(&modeltest.Group{
		Name: "group1",
		ID:   1,
	})

	insert(&modeltest.Admin{
		User:  modeltest.User{Username: "username1", Password: "password1"},
		Email: "email1",
		Group: 1,
	})

	insert(&modeltest.UserInfo{
		UID:       1,
		FirstName: "f1",
		LastName:  "l1",
		Sex:       "female",
	})
	insert(&modeltest.UserInfo{ // sex使用默认值
		UID:       2,
		FirstName: "f2",
		LastName:  "l2",
	})

	// select
	u1 := &modeltest.UserInfo{UID: 1}
	u2 := &modeltest.UserInfo{LastName: "l2", FirstName: "f2"}
	a1 := &modeltest.Admin{Email: "email1"}

	a.NotError(db.Select(u1))
	a.Equal(u1, &modeltest.UserInfo{UID: 1, FirstName: "f1", LastName: "l1", Sex: "female"})

	a.NotError(db.Select(u2))
	a.Equal(u2, &modeltest.UserInfo{UID: 2, FirstName: "f2", LastName: "l2", Sex: "male"})

	a.NotError(db.Select(a1))
	a.Equal(a1.Username, "username1")
}

func clearData(db *DB, a *assert.Assertion) {
	a.NotError(db.Drop(&modeltest.Admin{}))
	a.NotError(db.Drop(&modeltest.User{}))
	a.NotError(db.Drop(&modeltest.UserInfo{}))
	a.NotError(db.Close())
	closeDB(a)
}

func TestDB_Update(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	initData(db, a)
	defer clearData(db, a)

	// update
	r, err := db.Update(&modeltest.UserInfo{
		UID:       1,
		FirstName: "firstName1",
		LastName:  "lastName1",
		Sex:       "sex1",
	})
	a.NotError(err)
	cnt, err := r.RowsAffected()
	a.NotError(err).Equal(1, cnt)

	r, err = db.Update(&modeltest.UserInfo{
		UID:       2,
		FirstName: "firstName2",
		LastName:  "lastName2",
		Sex:       "sex2",
	})
	a.NotError(err)
	cnt, err = r.RowsAffected()
	a.NotError(err).Equal(1, cnt)

	u1 := &modeltest.UserInfo{UID: 1}
	a.NotError(db.Select(u1))
	a.Equal(u1, &modeltest.UserInfo{UID: 1, FirstName: "firstName1", LastName: "lastName1", Sex: "sex1"})

	u2 := &modeltest.UserInfo{LastName: "lastName2", FirstName: "firstName2"}
	a.NotError(db.Select(u2))
	a.Equal(u2, &modeltest.UserInfo{UID: 2, FirstName: "firstName2", LastName: "lastName2", Sex: "sex2"})
}

func TestDB_Delete(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	initData(db, a)
	defer clearData(db, a)

	// delete
	r, err := db.Delete(&modeltest.UserInfo{UID: 1})
	a.NotError(err)
	cnt, err := r.RowsAffected()
	a.NotError(err).Equal(cnt, 1)

	r, err = db.Delete(
		&modeltest.UserInfo{
			LastName:  "l2",
			FirstName: "f2",
		})
	a.NotError(err)
	cnt, err = r.RowsAffected()
	a.NotError(err).Equal(cnt, 1)

	r, err = db.Delete(&modeltest.Admin{Email: "email1"})
	a.NotError(err)
	cnt, err = r.RowsAffected()
	a.NotError(err).Equal(cnt, 1)

	hasCount(db, a, "user_info", 0)
	hasCount(db, a, "administrators", 0)

	// delete并不会重置ai计数
	_, err = db.Insert(&modeltest.Admin{Group: 1, Email: "email1"})
	a.NotError(err)
	a1 := &modeltest.Admin{Email: "email1"}
	a.NotError(db.Select(a1))
	a.Equal(a1.ID, 2) // a1.ID为一个自增列,不会在delete中被重置
}

func TestDB_Count(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	initData(db, a)
	defer clearData(db, a)

	// 单条件
	count, err := db.Count(
		&modeltest.UserInfo{
			UID: 1,
		},
	)
	a.NotError(err).Equal(1, count)

	// 条件不存在
	count, err = db.Count(
		&modeltest.Admin{Email: "email1-1000"}, // 该条件不存在
	)
	a.NotError(err).Equal(0, count)
}

func TestDB_Truncate(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	initData(db, a)
	defer clearData(db, a)

	hasCount(db, a, "administrators", 1)
	hasCount(db, a, "user_info", 2)

	// truncate之后，会重置AI
	a.NotError(db.Truncate(&modeltest.Admin{}))
	a.NotError(db.Truncate(&modeltest.UserInfo{}))
	hasCount(db, a, "administrators", 0)
	hasCount(db, a, "user_info", 0)

	_, err := db.Insert(&modeltest.Admin{Group: 1, Email: "email1"})
	a.NotError(err)

	a1 := &modeltest.Admin{Email: "email1"}
	a.NotError(db.Select(a1))
	a.Equal(1, a1.ID)
}

func TestDB_Drop(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	initData(db, a)
	defer clearData(db, a)

	a.NotError(db.Drop(&modeltest.UserInfo{}))
	a.NotError(db.Drop(&modeltest.Admin{}))
	r, err := db.Insert(&modeltest.Admin{})
	a.Error(err).Nil(r)
}
