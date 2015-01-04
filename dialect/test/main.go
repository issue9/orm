// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// github.com/issue9/orm/dialect的测试包.
// 通过运行go run *.go查看是否存在问题。
package main

import (
	"fmt"
	"os"

	"github.com/issue9/orm"
	"github.com/issue9/orm/dialect"
	"github.com/issue9/orm/fetch"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var sqlite3DBFile = "./test.db"

func main() {
	sqlite3CreateTable()

	mysqlCreateTable()

	postgresCreateTable()
}

func sqlite3CreateTable() {
	// 删除原有文件
	_, err := os.Stat(sqlite3DBFile)
	if err == nil || os.IsExist(err) {
		chkError(os.Remove(sqlite3DBFile))
	}

	chkError(dialect.Register("sqlite3", &dialect.Sqlite3{}))
	e, err := orm.New("sqlite3", sqlite3DBFile, "sqlite31", "sqlite31_")
	chkError(err)
	defer func() {
		e.Close()
		//os.Remove(sqlite3DBFile)
	}()

	runTest(e)
}

func mysqlCreateTable() {
	chkError(dialect.Register("mysql", &dialect.Mysql{}))
	e, err := orm.New("mysql", "root:@/test", "mysql1", "mysql1_")
	chkError(err)
	defer e.Close()

	runTest(e)
}

func postgresCreateTable() {
	chkError(dialect.Register("mysql", &dialect.Mysql{}))
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

// 获取数据库剩余的记录数量
func getCount(e *orm.Engine) int {
	rows, err := e.Query("SELECT count(*) AS c FROM user")
	chkError(err)
	defer rows.Close()

	ret, err := fetch.Column(true, "c", rows)
	chkError(err)

	if r, ok := ret[0].(int64); ok {
		return int(r)
	}

	return -1
}

func eq(v1, v2 int) {
	if v1 != v2 {
		chkError(fmt.Errorf("v1=[%v];v2=[%v]", v1, v2))
	}
}

// 运行测试内容
func runTest(e *orm.Engine) {
	chkError(e.Create(&User1{}))
	chkError(e.Insert(&User1{Username: "admin1", Group: "1"}))
	eq(1, getCount(e))

	chkError(e.Create(&User2{}))
	chkError(e.Insert(&User2{Username: "admin2", Group: 2}))
	eq(2, getCount(e))

	chkError(e.Create(&User3{}))
	chkError(e.Insert(&User3{Id: 3, Username: "admin3", Group: 3}))
	eq(3, getCount(e))

	chkError(e.Create(&User4{}))
	chkError(e.Insert(&User4{Id: 4, Account: "admin4", Group: 4}))
	eq(4, getCount(e))

	chkError(e.Create(&User5{}))
	chkError(e.Insert(&User5{Id: 5, Account: "admin5", Group: "5"}))
	eq(5, getCount(e))
}
