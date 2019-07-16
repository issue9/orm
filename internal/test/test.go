// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package test 提供了整个包的基本测试数据。
package test

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

// 需要测试的数据用例
var cases = []struct {
	prefix     string
	dsn        string
	dialect    orm.Dialect
	driverName string // 需要唯一
}{
	{
		prefix:     "prefix_",
		dsn:        "./orm_test.db",
		dialect:    dialect.Sqlite3(),
		driverName: "sqlite3",
	},
	{
		prefix:     "prefix_",
		dsn:        "user=postgres dbname=orm_test sslmode=disable",
		dialect:    dialect.Postgres(),
		driverName: "postgres",
	},
	{
		prefix:     "prefix_",
		dsn:        "root@/orm_test?charset=utf8&parseTime=true",
		dialect:    dialect.Mysql(),
		driverName: "mysql",
	},
}

// Test 单个测试用例
type Test struct {
	*assert.Assertion
	DB         *orm.DB
	DriverName string
	dsn        string
}

// Suite 测试用例管理
type Suite struct {
	a     *assert.Assertion
	tests []*Test
}

// NewSuite 初始化测试内容
func NewSuite(a *assert.Assertion) *Suite {
	s := &Suite{a: a}

	for _, c := range cases {
		db, err := orm.NewDB(c.driverName, c.dsn, c.prefix, c.dialect)
		a.NotError(err).NotNil(db)

		s.tests = append(s.tests, &Test{
			Assertion:  a,
			DB:         db,
			DriverName: c.driverName,
			dsn:        c.dsn,
		})
	}

	return s
}

func (s Suite) Close() {
	for _, t := range s.tests {
		t.NotError(t.DB.Close())

		if t.DB.Dialect().Name() != "sqlite3" {
			return
		}

		if _, err := os.Stat(t.dsn); err == nil || os.IsExist(err) {
			t.NotError(os.Remove(t.dsn))
		}
	}
}

// ForEach 为每个数据库测试用例调用 f 进行测试
//
// driverName 为需要测试的驱动，如果为空表示测试全部
func (s Suite) ForEach(f func(t *Test), driverName ...string) {
	if len(driverName) == 0 {
		for _, test := range s.tests {
			f(test)
		}
		return
	}

LOOP:
	for _, name := range driverName {
		for _, test := range s.tests {
			if test.DriverName == name {
				f(test)
				continue LOOP
			}
		}

		panic("不存在的 driverName:" + name)
	} // end for driverName
}
