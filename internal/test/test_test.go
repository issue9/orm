// SPDX-License-Identifier: MIT

package test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v3/internal/flagtest"
)

func TestMain(m *testing.M) {
	flagtest.Main(m)
}

func TestSuite_ForEach(t *testing.T) {
	a := assert.New(t)

	s := NewSuite(a)
	defer s.Close()

	var size int
	s.ForEach(func(t *Driver) {
		a.NotNil(t).
			NotNil(t.DB).
			NotNil(t.DB.Dialect()).
			NotNil(t.DB.DB).
			Equal(t.Assertion, a)
		size++
	})
	a.Equal(size, len(flagtest.Flags)).
		Equal(size, len(cases))
}

func TestSuite_ForEach_withDialect(t *testing.T) {
	a := assert.New(t)

	// 不再限定 Flags
	flagtest.Flags = []*flagtest.Flag{
		{DBName: "mysql", DriverName: "mysql"},
		{DBName: "sqlite3", DriverName: "sqlite3"},
		{DBName: "mariadb", DriverName: "mariadb"},
		{DBName: "postgres", DriverName: "postgres"},
	}

	// 通过参数 限定了 dialect

	s := NewSuite(a, Mysql, Sqlite3)
	defer s.Close()

	size := 0
	s.ForEach(func(t *Driver) {
		a.NotNil(t).
			NotNil(t.DB).
			NotNil(t.DB.Dialect()).
			NotNil(t.DB.DB).
			Equal(t.Assertion, a)
		size++
	})
	a.Equal(size, 2)
}
