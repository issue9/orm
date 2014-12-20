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
)

const testDBFile = "./test.db"

type FetchEmail struct {
	Email string `orm:"unique(unique_index);nullable;pk(pk_name)"`
}

type FetchUser struct {
	FetchEmail
	Id       int    `orm:"name(id);ai(1,2);"`
	Username string `orm:"index(index)"`
	Group    int    `orm:"name(group);fk(fk_group,group,id)"`

	Regdate int `orm:"-"`
}

// 初始化一个sql.DB(sqlite3)，方便后面的测试用例使用。
func initDB(a *assert.Assertion) *sql.DB {
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

	return db
}

// 关闭sql.DB(sqlite3)的数据库连结。
func closeDB(db *sql.DB, a *assert.Assertion) {
	a.NotError(db.Close()).
		NotError(os.Remove(testDBFile)).
		FileNotExists(testDBFile)
}

func TestDeleteOne(t *testing.T) {
	a := assert.New(t)
}
