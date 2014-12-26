// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

// 公用的测试文件

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/issue9/assert"
	"github.com/issue9/orm/core"
)

var style = assert.StyleSpace | assert.StyleTrim

// 测试数据库(sqlite3)所使用的文件地址
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

// 初始化一个sql.DB(sqlite3)，以及相关的Engine实例。
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

	for i := 0; i < 10; i++ {
		_, err = stmt.Exec(i, fmt.Sprintf("email-%d", i), fmt.Sprintf("username-%d", i), 1)
		a.NotError(err)
	}
	tx.Commit()
	stmt.Close()

	e, err := newEngine("sqlite3", testDBFile, "prefix_")
	a.NotError(err).NotNil(e)

	return db, e
}

// 关闭由initDB初始化的一切实例。
func closeDB(e *Engine, db *sql.DB, a *assert.Assertion) {
	e.close()

	a.NotError(db.Close()).
		NotError(os.Remove(testDBFile)).
		FileNotExists(testDBFile)
}
