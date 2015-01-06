// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// github.com/issue9/orm/dialect的测试包.
// 通过运行go run *.go查看是否存在问题。
package main

import (
	"fmt"
	"os"

	"github.com/issue9/conv"
	"github.com/issue9/orm"
	"github.com/issue9/orm/dialect"
	"github.com/issue9/orm/fetch"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var sqlite3DBFile = "./test.db"

func main() {
	testSqlite3()

	testMysql()

	testPostgres()
}

func testSqlite3() {
	// 删除原有文件
	_, err := os.Stat(sqlite3DBFile)
	if err == nil || os.IsExist(err) {
		chkError(os.Remove(sqlite3DBFile))
	}

	chkError(orm.Register("sqlite3", &dialect.Sqlite3{}))
	e, err := orm.New("sqlite3", sqlite3DBFile, "sqlite31", "sqlite31_")
	chkError(err)
	defer func() {
		e.Close()
		//os.Remove(sqlite3DBFile)
	}()

	runTest(e)
}

func testMysql() {
	chkError(orm.Register("mysql", &dialect.Mysql{}))
	e, err := orm.New("mysql", "root:@/test", "mysql1", "mysql1_")
	chkError(err)
	defer e.Close()

	runTest(e)
}

func testPostgres() {
	chkError(orm.Register("mysql", &dialect.Mysql{}))
	e, err := orm.New("pq", "dbname=test", "pq1", "pq1_")
	chkError(err)
	defer e.Close()

	runTest(e)
}

// 检测错误，若存在，则抛出异常
func chkError(err error) {
	if err != nil {
		// NOTE:panic在windows下中文乱码
		fmt.Println(err)
		os.Exit(1)
		//panic(err)
	}
}

// 检测数据库中记录数量是否和count相等。
func chkCount(e *orm.Engine, count int) {
	rows, err := e.Query("SELECT count(*) FROM user")
	chkError(err)
	defer rows.Close()

	// 在mysql中count(*)返回的是[]byte
	ret, err := fetch.Column(true, "count(*)", rows)
	chkError(err)

	r, err := conv.Int(ret[0])
	chkError(err)

	if r != count {
		chkError(fmt.Errorf("记录数量[%v]与预期值[%v]不相等", r, count))
	}
}

// 运行测试内容
func runTest(e *orm.Engine) {
	chkError(e.Create(&User1{}))
	chkError(e.Insert(&User1{Username: "admin1", Group: "1"}))
	chkCount(e, 1)

	chkError(e.Upgrade(&User2{}))
	chkError(e.Insert(&User2{Username: "admin2", Group: 2}))
	chkCount(e, 2)

	chkError(e.Upgrade(&User3{}))
	chkError(e.Insert(&User3{Id: 3, Username: "admin3", Group: 3}))
	chkCount(e, 3)

	chkError(e.Upgrade(&User4{}))
	chkError(e.Insert(&User4{Id: 4, Account: "admin4", Group: 4}))
	chkCount(e, 4)
}
