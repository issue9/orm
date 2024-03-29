// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package test

import (
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/sliceutil"

	"github.com/issue9/orm/v5/core"
)

func TestMain(m *testing.M) {
	Main(m)
}

func TestSuite_Run(t *testing.T) {
	a := assert.New(t, false)

	s := NewSuite(a)

	var size int
	s.Run(func(t *Driver) {
		a.NotNil(t).
			NotNil(t.DB).
			NotNil(t.DB.Dialect()).
			NotNil(t.DB.DB).
			Equal(t.Assertion, a)
		size++
	})
	a.Equal(size, len(flags))
}

func TestSuite_Run_withDialect(t *testing.T) {
	a := assert.New(t, false)

	// 不再限定 flags
	flags = []*flagVar{
		{DBName: "mysql", DriverName: "mysql"},
		{DBName: "sqlite3", DriverName: "sqlite3"},
		{DBName: "sqlite3", DriverName: "sqlite"},
		{DBName: "mariadb", DriverName: "mariadb"},
		{DBName: "postgres", DriverName: "postgres"},
	}

	// 通过参数限定了 dialect

	dialects := []core.Dialect{Mysql, Sqlite3}
	s := NewSuite(a, dialects...)

	size := 0
	s.Run(func(t *Driver) {
		a.NotNil(t).
			NotNil(t.DB).
			NotNil(t.DB.Dialect()).
			NotNil(t.DB.DB).
			Equal(t.Assertion, a)

		d := t.DB.Dialect()
		a.Equal(sliceutil.Count(dialects, func(i core.Dialect, _ int) bool {
			return i.DBName() == d.DBName() && i.DriverName() == d.DriverName()
		}), 1)

		size++
	})
	a.Equal(size, len(dialects))
}
