// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"database/sql"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/fetch"

	_ "github.com/mattn/go-sqlite3"
)

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

func TestDeleteOne(t *testing.T) {
	a := assert.New(t)
	db, e := initDB(a)
	a.NotNil(db).NotNil(e)
	defer closeDB(e, db, a)

	// 默认10条记录
	a.Equal(10, getCount(db, a))

	// 删除一条信息
	s := e.SQL()
	obj := &FetchUser{Id: 2}
	a.NotError(deleteOne(s.Reset(), obj))
	a.Equal(9, getCount(db, a))
}

func TestDeleteMult(t *testing.T) {
	a := assert.New(t)
	db, e := initDB(a)
	a.NotNil(db).NotNil(e)
	defer closeDB(e, db, a)

	// 默认10条记录
	a.Equal(10, getCount(db, a))

	// 删除多条信息
	s := e.SQL()
	obj := []*FetchUser{&FetchUser{Id: 2}, &FetchUser{Id: 3}}
	a.NotError(deleteMult(s.Reset(), obj))
	a.Equal(8, getCount(db, a))
}

func TestUpdateOne(t *testing.T) {
	a := assert.New(t)
	db, e := initDB(a)
	a.NotNil(db).NotNil(e)
	defer closeDB(e, db, a)

	s := e.SQL()
	obj := &FetchUser{Id: 2, FetchEmail: FetchEmail{Email: "12@test.com"}}
	a.NotError(updateOne(s.Reset(), obj))
	record := getRecord(db, 2, a)
	a.NotNil(record).Equal("12@test.com", record["Email"])
}

func TestUpdateMult(t *testing.T) {
	a := assert.New(t)
	db, e := initDB(a)
	a.NotNil(db).NotNil(e)
	defer closeDB(e, db, a)

	s := e.SQL()
	obj := []*FetchUser{
		&FetchUser{Id: 2, FetchEmail: FetchEmail{Email: "12@test.com"}},
		&FetchUser{Id: 3, FetchEmail: FetchEmail{Email: "13@test.com"}},
	}
	a.NotError(updateMult(s.Reset(), obj))
	record := getRecord(db, 2, a)
	a.NotNil(record).Equal("12@test.com", record["Email"])
	record = getRecord(db, 3, a)
	a.NotNil(record).Equal("13@test.com", record["Email"])
}

func TestInsertOne(t *testing.T) {
	a := assert.New(t)
	db, e := initDB(a)
	a.NotNil(db).NotNil(e)
	defer closeDB(e, db, a)

	// 默认10条记录
	a.Equal(10, getCount(db, a))

	s := e.SQL()
	obj := &FetchUser{Username: "abc"}
	a.NotError(insertOne(s.Reset(), obj))
	a.Equal(11, getCount(db, a))
}

func TestInsertMult(t *testing.T) {
	a := assert.New(t)
	db, e := initDB(a)
	a.NotNil(db).NotNil(e)
	defer closeDB(e, db, a)

	// 默认10条记录
	a.Equal(10, getCount(db, a))

	s := e.SQL()
	obj := []*FetchUser{
		&FetchUser{Username: "abc"},
		&FetchUser{Username: "def"},
	}
	a.NotError(insertMult(s.Reset(), obj))
	a.Equal(12, getCount(db, a))
}
