// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm_test

import (
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/test"
	"github.com/issue9/orm/v2/sqlbuilder"
)

// 测试性能的数据库驱动名称
const benchDBDriverName = "mysql"

func BenchmarkDB_Insert(b *testing.B) {
	a := assert.New(b)

	m := &Group{
		Name:    "name",
		Created: time.Now().Unix(),
	}

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		t.NotError(t.DB.Create(&Group{}))
		defer t.NotError(t.DB.Drop(&Group{}))

		for i := 0; i < b.N; i++ {
			t.NotError(t.DB.Insert(m))
		}
	}, benchDBDriverName)
}

func BenchmarkDB_Update(b *testing.B) {
	a := assert.New(b)

	m := &Group{
		Name:    "name",
		Created: time.Now().Unix(),
	}

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		t.NotError(t.DB.Create(&Group{}))
		defer t.NotError(t.DB.Drop(&Group{}))

		// 构造数据
		for i := 0; i < 10000; i++ {
			t.NotError(t.DB.Insert(m))
		}

		m.ID = 1 // 自增，从 1 开始
		for i := 0; i < b.N; i++ {
			t.NotError(t.DB.Update(m))
		}
	}, benchDBDriverName)
}

func BenchmarkDB_Select(b *testing.B) {
	a := assert.New(b)

	m := &Group{
		Name:    "name",
		Created: time.Now().Unix(),
	}

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		t.NotError(t.DB.Create(&Group{}))
		defer t.NotError(t.DB.Drop(&Group{}))

		t.NotError(t.DB.Insert(m))

		m.ID = 1
		for i := 0; i < b.N; i++ {
			t.NotError(t.DB.Select(m))
		}
	}, benchDBDriverName)
}

func BenchmarkDB_WhereUpdate(b *testing.B) {
	a := assert.New(b)

	m := &Group{
		Name:    "name",
		Created: time.Now().Unix(),
	}

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		t.NotError(t.DB.Create(&Group{}))
		defer t.NotError(t.DB.Drop(&Group{}))

		// 构造数据
		for i := 0; i < 10000; i++ {
			t.NotError(t.DB.Insert(m))
		}

		for i := 0; i < b.N; i++ {
			_, err := sqlbuilder.
				Update(t.DB).Table("{#groups}").
				Set("name", "n1").
				Increase("created", 1).
				Where("{id}=?", i+1).
				Exec()
			t.NotError(err)
		}
	}, benchDBDriverName)
}
