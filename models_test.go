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
	driver = "sqlite3"

	prefix = "prefix_"
	dsn    string
	d      forward.Dialect
)

type bench struct {
	ID   int    `orm:"name(id);ai"`
	Name string `orm:"name(name);len(20)"`
	Pass string `orm:"name(pass);len(32)"`
	Site string `orm:"name(site);len(32)"`
}

func (b *bench) Meta() string {
	return "name(bench)"
}

type user struct {
	ID       int    `orm:"name(id);ai;"`
	Username string `orm:"unique(unique_username);index(index_name);len(50)"`
	Password string `orm:"name(password);len(20)"`
	Regdate  int    `orm:"-"`
}

func (m *user) Meta() string {
	return "check(chk_name,id>0);engine(innodb);charset(utf-8);name(users)"
}

type userInfo struct {
	UID       int    `orm:"name(uid);pk"`
	FirstName string `orm:"name(firstName);unique(unique_name);len(20)"`
	LastName  string `orm:"name(lastName);unique(unique_name);len(20)"`
	Sex       string `orm:"name(sex);default(male);len(6)"`
}

func (m *userInfo) Meta() string {
	return "check(chk_name,uid>0);engine(innodb);charset(utf-8);name(user_info)"
}

type admin struct {
	user

	Email string `orm:"name(email);len(20);unique(unique_email)"`
	Group int    `orm:"name(group);"`
}

func (m *admin) Meta() string {
	return "check(chk_name,id>0);engine(innodb);charset(utf-8);name(administrators)"
}

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
func hasCount(db forward.Engine, a *assert.Assertion, table string, size int) {
	rows, err := db.Query(true, "SELECT COUNT(*) as cnt FROM #"+table)
	a.NotError(err).NotNil(rows)
	defer func() {
		a.NotError(rows.Close())
	}()

	data, err := fetch.Map(true, rows)
	a.NotError(err).NotNil(data)
	a.Equal(conv.MustInt(data[0]["cnt"], -1), size)
}
