// SPDX-License-Identifier: MIT

// Package test 提供了整个包的基本测试数据
package test

import (
	"os"

	"github.com/issue9/assert"
	"github.com/issue9/sliceutil"

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
	Sqlite3 = dialect.Sqlite3("sqlite3", "sqlite3")

	// Mysql Dialect 实例
	Mysql = dialect.Mysql("mysql", "mysql")

	// Mariadb Dialect 实例
	Mariadb = dialect.Mysql("mariadb", "mysql")

	// Postgres Dialect 实例
	Postgres = dialect.Postgres("postgresql", "postgres")
)

// 以驱动为单的测试用例
var cases = []struct {
	prefix  string
	dsn     string
	dialect orm.Dialect
}{
	{
		prefix:  "prefix_",
		dsn:     "./" + sqlite3DBFile + "?_fk=true&_loc=UTC",
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
	{
		prefix:  "prefix_",
		dsn:     "root:root@/orm_test?charset=utf8&parseTime=true",
		dialect: Mariadb,
	},
}

// Driver 单个测试用例
type Driver struct {
	*assert.Assertion
	DB         *orm.DB
	DriverName string
	DBName     string
	dsn        string
}

// Suite 测试用例管理
type Suite struct {
	a       *assert.Assertion
	drivers []*Driver
}

// NewSuite 初始化测试内容
func NewSuite(a *assert.Assertion) *Suite {
	parseFlag()

	s := &Suite{a: a}

	for _, c := range cases {
		db, err := orm.NewDB(c.dsn, c.prefix, c.dialect)
		a.NotError(err).NotNil(db)

		name := c.dialect.DBName()
		driver := c.dialect.DriverName()

		if len(dbs) != 0 && sliceutil.Count(dbs, func(i int) bool { return dbs[i].dbName == name && dbs[i].driverNme == driver }) <= 0 {
			continue
		}

		s.drivers = append(s.drivers, &Driver{
			Assertion:  a,
			DB:         db,
			DriverName: driver,
			DBName:     name,
			dsn:        c.dsn,
		})
	}

	return s
}

// Close 销毁测试用例并关闭数据库
//
// 如果是 sqlite3，还会删除数据库文件。
func (s Suite) Close() {
	for _, t := range s.drivers {
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
func (s Suite) ForEach(f func(*Driver), dialects ...core.Dialect) {
	if len(dialects) == 0 {
		for _, test := range s.drivers {
			f(test)
		}
		return
	}

LOOP:
	for _, d := range dialects {
		for _, test := range s.drivers {
			if test.DriverName == d.DriverName() && test.DBName == d.DBName() {
				f(test)
				continue LOOP
			}
		}

		s.a.TB().Logf("忽略的驱动 :%s:%s\n", d.DBName(), d.DriverName())
	}
}
