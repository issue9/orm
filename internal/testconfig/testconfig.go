// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package testconfig 为测试内容提供一个统一的数据库环境
package testconfig

import (
	"os"

	"github.com/issue9/assert"
	"github.com/issue9/orm/v2"
	"github.com/issue9/orm/v2/dialect"

	// 供其它包测试用，直接在此引用数据库包会更文件
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	// 通过修改此值来确定使用哪个数据库驱动来测试
	// 若需要其它两种数据库测试，需要先在创建相应的数据库
	driver = "sqlite3"

	prefix = "prefix_"

	dsn string
)

// CloseDB 销毁数据库。
//
// 如果数据库类型为 sqlite3，则还会删除数据库文件。
func CloseDB(db *orm.DB, a *assert.Assertion, dropTable ...interface{}) {
	a.NotError(db.MultDrop(dropTable...))

	a.NotError(db.Close())

	if driver == "sqlite3" {
		if _, err := os.Stat(dsn); err == nil || os.IsExist(err) {
			a.NotError(os.Remove(dsn))
		}
	}
}

// NewDB 声明 orm.DB 实例
func NewDB(a *assert.Assertion) *orm.DB {
	var d orm.Dialect

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
		panic("仅支持 mysql,sqlite3,postgres 三种数据库测试")
	}

	db, err := orm.NewDB(driver, dsn, prefix, d)
	a.NotError(err).NotNil(db)
	return db
}
