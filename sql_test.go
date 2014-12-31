// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"testing"

	"github.com/issue9/assert"
)

func TestDelete(t *testing.T) {
	a := assert.New(t)
	db, e := initDB(a)
	defer closeDB(e, db, a)

	sql := e.SQL().
		Table("user").
		Where("id=?", 1)
	a.StringEqual(sql.deleteSQL(), "DELETE FROM user WHERE(id=?)", style)
	result, err := sql.Delete()
	a.NotError(err).NotNil(result)

	sql.Reset().
		Table("user").
		Where("id=?", 2)
	result, err = sql.Delete(3)
	a.Equal(8, getCount(db, a))
	a.Nil(getRecord(db, 3, a))    // 3被删除
	a.NotNil(getRecord(db, 2, a)) // 2还存在
}

func TestUpdate(t *testing.T) {
	a := assert.New(t)
	db, e := initDB(a)
	defer closeDB(e, db, a)

	sql := e.SQL().
		Table("user").
		Where("id=?", 1).
		Columns("Email", "{group}")
	a.StringEqual(sql.updateSQL(), "UPDATE user SET Email=?,{group}=? WHERE(id=?)", style)

	result, err := sql.Update("email@test.com", 2, 3)
	a.NotError(err).NotNil(result)

	record := getRecord(db, 3, a)
	a.Equal(record["Email"], "email@test.com")
	a.Equal(record["group"], "2")
}

func TestInsert(t *testing.T) {
	a := assert.New(t)
	db, e := initDB(a)
	defer closeDB(e, db, a)

	sql := e.SQL().
		Table("user").
		Add("Email", "email@test.com").
		Add("{group}", 1).
		Add("Username", "username")
	a.StringEqual(sql.insertSQL(), "INSERT INTO user(Email,{group},Username) VALUES(?,?,?)", style)

	// 插入一条数据
	result, err := sql.Insert()
	a.NotError(err).NotNil(result)
	a.Equal(11, getCount(db, a))

	id, err := result.LastInsertId()
	a.NotError(err)
	r := getRecord(db, int(id), a)
	a.Equal(r["Username"], "username").
		Equal(r["Email"], "email@test.com")

	// 再次插入一条数据
	result, err = sql.Insert("abc@test.com", 2, "username")
	a.NotError(err).NotNil(result)
	a.Equal(getCount(db, a), 12)

	id, err = result.LastInsertId()
	a.NotError(err)
	r = getRecord(db, int(id), a)
	a.Equal(r["Username"], "username").
		Equal(r["Email"], "abc@test.com").
		Equal(r["group"], "2")
}

func TestSelect(t *testing.T) {
	a := assert.New(t)
	db, e := initDB(a)
	defer closeDB(e, db, a)

	// 最基本的查询
	sql := e.SQL().
		Table("user").
		Columns("*").
		Where("id<?", 1)
	a.StringEqual(sql.selectSQL(), "SELECT * FROM user WHERE(id<?)", style)

	m, err := sql.Fetch2Map(2)
	a.NotError(err).Equal(m["id"], 0)

	// 带排序的查询
	sql.Reset().
		Table("user").
		Columns("*").
		Where("id<?").
		Desc("id").
		Limit(2)
	a.StringEqual(sql.selectSQL(), "SELECT * FROM user WHERE(id<?)ORDER BY id DESC LIMIT ?", style)

	arr, err := sql.Fetch2Maps(5, 2) // 小于5的ID，倒序两条记录
	a.NotError(err).
		Equal(2, len(arr)).
		Equal(4, arr[0]["id"]). // 4
		Equal(3, arr[1]["id"])  // 3
}
