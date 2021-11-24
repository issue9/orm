// SPDX-License-Identifier: MIT

package test

import (
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/sliceutil"

	"github.com/issue9/orm/v4/core"
	"github.com/issue9/orm/v4/internal/flagtest"
)

func TestMain(m *testing.M) {
	flagtest.Main(m)
}

func TestSuite_ForEach(t *testing.T) {
	a := assert.New(t, false)

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
	a.Equal(size, len(flagtest.Flags))
}

func TestSuite_ForEach_withDialect(t *testing.T) {
	a := assert.New(t, false)

	// 不再限定 Flags
	flagtest.Flags = []*flagtest.Flag{
		{DBName: "mysql", DriverName: "mysql"},
		{DBName: "sqlite3", DriverName: "sqlite3"},
		{DBName: "mariadb", DriverName: "mariadb"},
		{DBName: "postgres", DriverName: "postgres"},
	}

	// 通过参数 限定了 dialect

	dialects := []core.Dialect{Mysql, Sqlite3}
	s := NewSuite(a, dialects...)
	defer s.Close()

	size := 0
	s.ForEach(func(t *Driver) {
		a.NotNil(t).
			NotNil(t.DB).
			NotNil(t.DB.Dialect()).
			NotNil(t.DB.DB).
			Equal(t.Assertion, a)

		d := t.DB.Dialect()
		a.Equal(sliceutil.Count(dialects, func(i int) bool {
			return dialects[i].DBName() == d.DBName() && dialects[i].DriverName() == d.DriverName()
		}), 1)

		size++
	})
	a.Equal(size, len(dialects))
}
