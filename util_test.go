// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"os"

	"github.com/issue9/assert"
	"github.com/issue9/conv"
	"github.com/issue9/orm/dialect"
	"github.com/issue9/orm/fetch"
	"github.com/issue9/orm/forward"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	// 通过修改此值来确定使用哪个数据库驱动来测试
	driver = "mysql"

	prefix = "prefix_"
	dsn    string
	d      forward.Dialect
)

// 销毁数据库。默认仅对Sqlite3启作用，删除该数据库文件。
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

// table表中是否存在size条记录，若不是，则触发error
func hasCount(db *DB, a *assert.Assertion, table string, size int) {
	rows, err := db.Query(true, "SELECT COUNT(*) as cnt FROM #"+table)
	a.NotError(err).NotNil(rows)
	defer func() {
		a.NotError(rows.Close())
	}()

	data, err := fetch.Map(true, rows)
	a.NotError(err).NotNil(data)
	a.Equal(conv.MustInt(data[0]["cnt"], -1), size)
}
