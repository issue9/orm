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

// mysql: BenchmarkDB_Insert-4     	    5000	    280546 ns/op
func BenchmarkDB_Insert(b *testing.B) {
	a := assert.New(b)

	m := &Group{
		Name:    "name",
		Created: time.Now().Unix(),
	}

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		t.NotError(t.DB.Create(&Group{}))
		defer t.NotError(t.DB.Drop(&Group{}))

		for i := 0; i < b.N; i++ {
			t.NotError(t.DB.Insert(m))
		}
	}, benchDBDriverName)
}

// mysql: BenchmarkDB_Update-4     	    5000	    369461 ns/op
func BenchmarkDB_Update(b *testing.B) {
	a := assert.New(b)

	m := &Group{
		Name:    "name",
		Created: time.Now().Unix(),
	}

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
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

// mysql: BenchmarkDB_Select-4     	   10000	    218232 ns/op
func BenchmarkDB_Select(b *testing.B) {
	a := assert.New(b)

	m := &Group{
		Name:    "name",
		Created: time.Now().Unix(),
	}

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		t.NotError(t.DB.Create(&Group{}))
		defer t.NotError(t.DB.Drop(&Group{}))

		t.NotError(t.DB.Insert(m))

		m.ID = 1
		for i := 0; i < b.N; i++ {
			t.NotError(t.DB.Select(m))
		}
	}, benchDBDriverName)
}

// mysql: BenchmarkDB_WhereUpdate-4	   10000	    163209 ns/op
func BenchmarkDB_WhereUpdate(b *testing.B) {
	a := assert.New(b)

	m := &Group{
		Name:    "name",
		Created: time.Now().Unix(),
	}

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
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

// mysql: BenchmarkDB_Count-4      	   10000	    186920 ns/op
func BenchmarkDB_Count(b *testing.B) {
	a := assert.New(b)

	m := &Group{
		Name:    "name",
		Created: time.Now().Unix(),
	}

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		t.NotError(t.DB.Create(&Group{}))
		defer t.NotError(t.DB.Drop(&Group{}))

		t.NotError(t.DB.Insert(m))

		be := &Group{Name: "name"}
		for i := 0; i < b.N; i++ {
			count, _ := t.DB.Count(be)
			if count < 1 {
				t.Error("count:", count)
			}
		}
	}, benchDBDriverName)
}
