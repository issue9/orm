// SPDX-License-Identifier: MIT

// Package test 提供了整个包的基本测试数据
package test

import (
	"os"

	"github.com/issue9/assert/v3"
	"github.com/issue9/sliceutil"

	"github.com/issue9/orm/v5"
	"github.com/issue9/orm/v5/core"
	"github.com/issue9/orm/v5/dialect"

	// 测试入口，数据库也在此初始化
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

const sqlite3DBFile = "orm_test.db"

var (
	// Sqlite3 Dialect 实例
	Sqlite3 = dialect.Sqlite3("sqlite3")

	// Sqlite Dialect 实例
	Sqlite = dialect.Sqlite3("sqlite")

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
		dsn:     "./" + sqlite3DBFile + "?_fk=true&_loc=UTC",
		dialect: Sqlite,
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
// dialect 指定了当前需要测试的驱动，若未指定表示测试 flags 中的所有内容。
func NewSuite(a *assert.Assertion, dialect ...core.Dialect) *Suite {
	s := &Suite{a: a}
	a.TB().Cleanup(func() {
		s.close()
	})

	for _, c := range cases {
		name := c.dialect.DBName()
		driver := c.dialect.DriverName()

		if len(dialect) > 0 && sliceutil.Count(dialect, func(i core.Dialect, _ int) bool { return i.DBName() == name && i.DriverName() == driver }) <= 0 {
			continue
		}

		fs := flags
		if len(fs) > 0 && sliceutil.Count(fs, func(i *flagVar, _ int) bool { return i.DBName == name && i.DriverName == driver }) <= 0 {
			continue
		}

		db, err := orm.NewDB(c.dsn, c.dialect)
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

func (s Suite) close() {
	for _, t := range s.drivers {
		t.NotError(t.DB.Close())

		if t.DB.Dialect().DriverName() != Sqlite3.DriverName() {
			return
		}

		// sqlite3 删除数据库文件
		if _, err := os.Stat(sqlite3DBFile); err == nil || os.IsExist(err) {
			t.NotError(os.Remove(sqlite3DBFile))
		}
	}
}

// Run 为每个数据库测试用例调用 f 进行测试
func (s Suite) Run(f func(*Driver)) {
	for _, test := range s.drivers {
		f(test)
	}
}
