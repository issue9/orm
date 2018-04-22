// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm_test

import (
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/orm"
	"github.com/issue9/orm/internal/modeltest"
	"github.com/issue9/orm/sqlbuilder"
)

// go1.10 BenchmarkNewModelNoCached-4   	  200000	      8161 ns/op
func BenchmarkNewModelNoCached(b *testing.B) {
	orm.Clear()
	a := assert.New(b)

	for i := 0; i < b.N; i++ {
		m, err := orm.NewModel(&modeltest.User{})
		orm.Clear()
		a.NotError(err).NotNil(m)
	}
}

// go1.10 BenchmarkNewModelCached-4     	10000000	       187 ns/op
func BenchmarkNewModelCached(b *testing.B) {
	orm.Clear()
	a := assert.New(b)

	for i := 0; i < b.N; i++ {
		m, err := orm.NewModel(&modeltest.User{})
		a.NotError(err).NotNil(m)
	}
}

// mysql: BenchmarkDB_Insert-4     	    5000	    280546 ns/op
func BenchmarkDB_Insert(b *testing.B) {
	a := assert.New(b)

	m := &modeltest.Group{
		Name:    "name",
		Created: time.Now().Unix(),
	}

	db := newDB(a)
	defer func() {
		db.Drop(&modeltest.Group{})
		closeDB(a)
	}()

	a.NotError(db.Create(&modeltest.Group{}))

	for i := 0; i < b.N; i++ {
		a.NotError(db.Insert(m))
	}
}

// mysql: BenchmarkDB_Update-4     	    5000	    369461 ns/op
func BenchmarkDB_Update(b *testing.B) {
	a := assert.New(b)

	m := &modeltest.Group{
		Name:    "name",
		Created: time.Now().Unix(),
	}

	db := newDB(a)
	defer func() {
		db.Drop(&modeltest.Group{})
		closeDB(a)
	}()

	// 构造数据
	a.NotError(db.Create(&modeltest.Group{}))
	a.NotError(db.Insert(m))

	m.ID = 1 // 自增，从 1 开始
	for i := 0; i < b.N; i++ {
		a.NotError(db.Update(m))
	}
}

// mysql: BenchmarkDB_Select-4     	   10000	    218232 ns/op
func BenchmarkDB_Select(b *testing.B) {
	a := assert.New(b)

	m := &modeltest.Group{
		Name:    "name",
		Created: time.Now().Unix(),
	}

	db := newDB(a)
	defer func() {
		db.Drop(&modeltest.Group{})
		closeDB(a)
	}()

	a.NotError(db.Create(&modeltest.Group{}))
	a.NotError(db.Insert(m))

	m.ID = 1
	for i := 0; i < b.N; i++ {
		a.NotError(db.Select(m))
	}
}

// mysql: BenchmarkDB_WhereUpdate-4	   10000	    163209 ns/op
func BenchmarkDB_WhereUpdate(b *testing.B) {
	a := assert.New(b)

	m := &modeltest.Group{
		Name:    "name",
		Created: time.Now().Unix(),
	}

	db := newDB(a)
	defer func() {
		db.Drop(&modeltest.Group{})
		closeDB(a)
	}()

	// 构造数据
	a.NotError(db.Create(&modeltest.Group{}))
	a.NotError(db.Insert(m))

	for i := 0; i < b.N; i++ {
		_, err := sqlbuilder.
			Update(db).Table("{#groups}").
			Set("name", "n1").
			Increase("created", 1).
			Where("{id}=?", i+1).
			Exec()
		a.NotError(err)
	}
}

// mysql: BenchmarkDB_Count-4      	   10000	    186920 ns/op
func BenchmarkDB_Count(b *testing.B) {
	a := assert.New(b)

	m := &modeltest.Group{
		Name:    "name",
		Created: time.Now().Unix(),
	}

	db := newDB(a)
	defer func() {
		db.Drop(&modeltest.Group{})
		closeDB(a)
	}()

	// 构造数据
	a.NotError(db.Create(&modeltest.Group{}))
	a.NotError(db.Insert(m))

	be := &modeltest.Group{Name: "name"}
	for i := 0; i < b.N; i++ {
		count, _ := db.Count(be)
		if count < 1 {
			b.Error("count:", count)
		}
	}
}
