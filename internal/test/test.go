// SPDX-License-Identifier: MIT

// Package test 提供了整个包的基本测试数据
package test

import (
	"os"

	"github.com/issue9/assert/v2"
	"github.com/issue9/sliceutil"

	"github.com/issue9/orm/v5"
	"github.com/issue9/orm/v5/core"
	"github.com/issue9/orm/v5/dialect"
	"github.com/issue9/orm/v5/internal/flagtest"

	// 测试入口，数据库也在此初始化
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

const sqlite3DBFile = "orm_test.db"

const Prefix = "prefix_"

var (
	// Sqlite3 Dialect 实例
	Sqlite3 = dialect.Sqlite3("sqlite3")

	// Mysql Dialect 实例
	Mysql = dialect.Mysql("mysql")

	// Mariadb Dialect 实例
	Mariadb = dialect.Mariadb("mysql")

	// Postgres Dialect 实例
	Postgres = dialect.Postgres("postgres")
)

// 以驱动为单的测试用例
//
// 部分设置项需要与 action 中的设置相同才能正常启动，比如端口号等。
var cases = []struct {
	dsn     string
	dialect orm.Dialect
}{
	{
		dsn:     "./" + sqlite3DBFile + "?_fk=true&_loc=UTC",
		dialect: Sqlite3,
	},
	{
		dsn:     "user=postgres password=postgres dbname=orm_test sslmode=disable",
		dialect: Postgres,
	},
	{
		dsn:     "root:root@/orm_test?charset=utf8&parseTime=true",
		dialect: Mysql,
	},
	{
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
//
// dialect 指定了当前需要测试的驱动，如果未指定表示需要测全部，否则只测试指定内容。
// 同时 dialect 还受到 flagsets.Flags 变量的影响，如果未在其中指定的，也不会执行测试。
func NewSuite(a *assert.Assertion, dialect ...core.Dialect) *Suite {
	s := &Suite{a: a}

	for _, c := range cases {
		name := c.dialect.DBName()
		driver := c.dialect.DriverName()

		if len(dialect) > 0 && sliceutil.Count(dialect, func(i core.Dialect) bool { return i.DBName() == name && i.DriverName() == driver }) <= 0 {
			continue
		}

		fs := flagtest.Flags
		if len(fs) > 0 && sliceutil.Count(fs, func(i *flagtest.Flag) bool { return i.DBName == name && i.DriverName == driver }) <= 0 {
			continue
		}

		db, err := orm.NewDB(c.dsn, Prefix, c.dialect)
		a.NotError(err).NotNil(db)

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
func (s Suite) ForEach(f func(*Driver)) {
	for _, test := range s.drivers {
		f(test)
	}
}
