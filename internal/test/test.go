// SPDX-License-Identifier: MIT

// Package test 提供了整个包的基本测试数据
package test

import (
	"os"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v3"
	"github.com/issue9/orm/v3/core"
	"github.com/issue9/orm/v3/dialect"

	// 测试入口，数据库也在此初始化
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

const sqlite3DBFile = "orm_test.db"

var (
	// Sqlite3 Dialect 实例
	Sqlite3 = dialect.Sqlite3("sqlite3")

	// Mysql Dialect 实例
	Mysql = dialect.Mysql("mysql")

	// Postgres Dialect 实例
	Postgres = dialect.Postgres("postgres")
)

// 需要测试的数据用例
var cases = []struct {
	prefix  string
	dsn     string
	dialect orm.Dialect
}{
	{
		prefix:  "prefix_",
		dsn:     "./" + sqlite3DBFile + "?_fk=true",
		dialect: Sqlite3,
	},
	{
		prefix:  "prefix_",
		dsn:     "user=postgres password=postgres dbname=orm_test sslmode=disable",
		dialect: Postgres,
	},
	{
		prefix:  "prefix_",
		dsn:     "root:root@/orm_test?charset=utf8&parseTime=true",
		dialect: Mysql,
	},
}

// Driver 单个测试用例
type Driver struct {
	*assert.Assertion
	DB         *orm.DB
	DriverName string
	dsn        string
}

// Suite 测试用例管理
type Suite struct {
	a     *assert.Assertion
	tests []*Driver
}

// NewSuite 初始化测试内容
func NewSuite(a *assert.Assertion) *Suite {
	s := &Suite{a: a}

	for _, c := range cases {
		db, err := orm.NewDB(c.dsn, c.prefix, c.dialect)
		a.NotError(err).NotNil(db)

		s.tests = append(s.tests, &Driver{
			Assertion:  a,
			DB:         db,
			DriverName: c.dialect.DriverName(),
			dsn:        c.dsn,
		})
	}

	return s
}

// Close 销毁测试用例并关闭数据库
//
// 如果是 sqlite3，还会删除数据库文件。
func (s Suite) Close() {
	for _, t := range s.tests {
		t.NotError(t.DB.Close())

		if t.DB.Dialect().DriverName() != Sqlite3.DriverName() {
			return
		}

		if _, err := os.Stat(sqlite3DBFile); err == nil || os.IsExist(err) {
			t.NotError(os.Remove(sqlite3DBFile))
		}
	}
}

// ForEach 为每个数据库测试用例调用 f 进行测试
//
// dialects 为需要测试的驱动，如果为空表示测试全部
func (s Suite) ForEach(f func(t *Driver), dialects ...core.Dialect) {
	if len(dialects) == 0 {
		for _, test := range s.tests {
			f(test)
		}
		return
	}

LOOP:
	for _, d := range dialects {
		for _, test := range s.tests {
			if test.DriverName == d.DriverName() {
				f(test)
				continue LOOP
			}
		}

		panic("不存在的 driverName:" + d.DriverName())
	}
}
