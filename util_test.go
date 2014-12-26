// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/core"
	"github.com/issue9/orm/fetch"

	_ "github.com/mattn/go-sqlite3"
)

const testDBFile = "./test.db"

type FetchEmail struct {
	Email string `orm:"unique(unique_index);nullable;pk"`
}

type FetchUser struct {
	FetchEmail
	Id       int    `orm:"name(id);ai(1,2);"`
	Username string `orm:"index(index)"`
	Group    int    `orm:"name(group);fk(fk_group,group,id)"`

	Regdate int `orm:"-"`
}

// core.Metaer.Meta()
func (fu *FetchUser) Meta() string {
	return "check(chk_name,id>5);engine(innodb);charset(utf-8);name(user)"
}

var _ core.Metaer = &FetchUser{}

// 初始化一个sql.DB(sqlite3)，方便后面的测试用例使用。
func initDB(a *assert.Assertion) (*sql.DB, *Engine) {
	db, err := sql.Open("sqlite3", testDBFile)
	a.NotError(err).NotNil(db)

	/* 创建表 */
	sql := `create table user (
        id integer not null primary key, 
        Email text,
        Username text,
        [group] interger)`
	_, err = db.Exec(sql)
	a.NotError(err)

	/* 插入数据 */
	tx, err := db.Begin()
	a.NotError(err).NotNil(tx)

	stmt, err := tx.Prepare("insert into user(id, Email,Username,[group]) values(?, ?, ?, ?)")
	a.NotError(err).NotNil(stmt)

	for i := 0; i < 100; i++ {
		_, err = stmt.Exec(i, fmt.Sprintf("email-%d", i), fmt.Sprintf("username-%d", i), 1)
		a.NotError(err)
	}
	tx.Commit()
	stmt.Close()

	e, err := newEngine("sqlite3", testDBFile, "prefix_")
	a.NotError(err).NotNil(e)

	return db, e
}

// 获取数据库剩余的记录数量
func getCount(db *sql.DB, a *assert.Assertion) int {
	rows, err := db.Query("SELECT count(*) AS c FROM user")
	a.NotError(err).NotNil(rows)
	defer rows.Close()

	ret, err := fetch.Column(true, "c", rows)
	a.NotError(err).NotNil(ret)

	if r, ok := ret[0].(int64); ok {
		return int(r)
	}

	return -1
}

// 获取指定ID的记录
func getRecord(db *sql.DB, id int, a *assert.Assertion) map[string]string {
	rows, err := db.Query("SELECT * FROM user WHERE id=? LIMIT 1", id)
	a.NotError(err).NotNil(rows)
	defer rows.Close()

	ret, err := fetch.MapString(true, rows)
	a.NotError(err).NotNil(ret)

	return ret[0]
}

// 关闭sql.DB(sqlite3)的数据库连结。
func closeDB(db *sql.DB, a *assert.Assertion) {
	a.NotError(db.Close()).
		NotError(os.Remove(testDBFile)).
		FileNotExists(testDBFile)
}

func TestDeleteOne(t *testing.T) {
	a := assert.New(t)
	db, e := initDB(a)
	a.NotNil(db).NotNil(e)
	defer closeDB(db, a)
	defer e.close()

	// 默认100条记录
	a.Equal(100, getCount(db, a))

	// 删除一条信息
	s := e.SQL()
	obj := &FetchUser{Id: 12}
	a.NotError(deleteOne(s.Reset(), obj))
	a.Equal(99, getCount(db, a))
}

func TestDeleteMult(t *testing.T) {
	a := assert.New(t)
	db, e := initDB(a)
	a.NotNil(db).NotNil(e)
	defer closeDB(db, a)
	defer e.close()

	// 默认100条记录
	a.Equal(100, getCount(db, a))

	// 删除多条信息
	s := e.SQL()
	obj := []*FetchUser{&FetchUser{Id: 12}, &FetchUser{Id: 13}}
	a.NotError(deleteMult(s.Reset(), obj))
	a.Equal(98, getCount(db, a))
}

func TestUpdateOne(t *testing.T) {
	a := assert.New(t)
	db, e := initDB(a)
	a.NotNil(db).NotNil(e)
	defer closeDB(db, a)
	defer e.close()

	s := e.SQL()
	obj := &FetchUser{Id: 12, FetchEmail: FetchEmail{Email: "12@test.com"}}
	a.NotError(updateOne(s.Reset(), obj))
	record := getRecord(db, 12, a)
	a.NotNil(record).Equal("12@test.com", record["Email"])
}

func TestUpdateMult(t *testing.T) {
	a := assert.New(t)
	db, e := initDB(a)
	a.NotNil(db).NotNil(e)
	defer closeDB(db, a)
	defer e.close()

	s := e.SQL()
	obj := []*FetchUser{
		&FetchUser{Id: 12, FetchEmail: FetchEmail{Email: "12@test.com"}},
		&FetchUser{Id: 13, FetchEmail: FetchEmail{Email: "13@test.com"}},
	}
	a.NotError(updateMult(s.Reset(), obj))
	record := getRecord(db, 12, a)
	a.NotNil(record).Equal("12@test.com", record["Email"])
	record = getRecord(db, 13, a)
	a.NotNil(record).Equal("13@test.com", record["Email"])
}

func TestInsertOne(t *testing.T) {
	a := assert.New(t)
	db, e := initDB(a)
	a.NotNil(db).NotNil(e)
	defer closeDB(db, a)
	defer e.close()

	// 默认100条记录
	a.Equal(100, getCount(db, a))

	s := e.SQL()
	obj := &FetchUser{Username: "abc"}
	a.NotError(insertOne(s.Reset(), obj))
	a.Equal(101, getCount(db, a))
}

func TestInsertMult(t *testing.T) {
	a := assert.New(t)
	db, e := initDB(a)
	a.NotNil(db).NotNil(e)
	defer closeDB(db, a)
	defer e.close()

	// 默认100条记录
	a.Equal(100, getCount(db, a))

	s := e.SQL()
	obj := []*FetchUser{
		&FetchUser{Username: "abc"},
		&FetchUser{Username: "def"},
	}
	a.NotError(insertMult(s.Reset(), obj))
	a.Equal(102, getCount(db, a))
}
