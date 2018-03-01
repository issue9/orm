// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/sqlbuilder"
)

// mysql: BenchmarkDB_Insert-4     	    5000	    280546 ns/op
func BenchmarkDB_Insert(b *testing.B) {
	a := assert.New(b)

	m := &bench{
		Name: "name",
		Pass: "pass",
		Site: "http://www.github.com/issue9/orm",
	}

	db := newDB(a)
	defer func() {
		db.Drop(&bench{})
		closeDB(a)
	}()

	a.NotError(db.Create(&bench{}))

	for i := 0; i < b.N; i++ {
		a.NotError(db.Insert(m))
	}
}

// mysql: BenchmarkDB_Update-4     	    5000	    369461 ns/op
func BenchmarkDB_Update(b *testing.B) {
	a := assert.New(b)

	m := &bench{
		Name: "name",
		Pass: "pass",
		Site: "http://www.github.com/issue9/orm",
	}

	db := newDB(a)
	defer func() {
		db.Drop(&bench{})
		closeDB(a)
	}()

	// 构造数据
	a.NotError(db.Create(&bench{}))
	a.NotError(db.Insert(m))

	m.ID = 1 // 自增，从1开始
	for i := 0; i < b.N; i++ {
		a.NotError(db.Update(m))
	}
}

// mysql: BenchmarkDB_Select-4     	   10000	    218232 ns/op
func BenchmarkDB_Select(b *testing.B) {
	a := assert.New(b)

	m := &bench{
		Name: "name",
		Pass: "pass",
		Site: "http://www.github.com/issue9/orm",
	}

	db := newDB(a)
	defer func() {
		db.Drop(&bench{})
		closeDB(a)
	}()

	a.NotError(db.Create(&bench{}))
	a.NotError(db.Insert(m))

	m.ID = 1
	for i := 0; i < b.N; i++ {
		a.NotError(db.Select(m))
	}
}

// mysql: BenchmarkDB_WhereUpdate-4	   10000	    163209 ns/op
func BenchmarkDB_WhereUpdate(b *testing.B) {
	a := assert.New(b)

	m := &bench{
		Name: "name",
		Pass: "pass",
		Site: "http://www.github.com/issue9/orm",
	}

	db := newDB(a)
	defer func() {
		db.Drop(&bench{})
		closeDB(a)
	}()

	// 构造数据
	a.NotError(db.Create(&bench{}))
	a.NotError(db.Insert(m))

	for i := 0; i < b.N; i++ {
		_, err := sqlbuilder.
			Update(db, "{#bench}").
			Set("name", "n1").
			Set("pass", "p1").
			Set("site", "s1").
			Where("{id}=?", i+1).
			Exec()
		a.NotError(err)
	}
}

// mysql: BenchmarkDB_Count-4      	   10000	    186920 ns/op
func BenchmarkDB_Count(b *testing.B) {
	a := assert.New(b)

	m := &bench{
		Name: "name",
		Pass: "pass",
		Site: "http://www.github.com/issue9/orm",
	}

	db := newDB(a)
	defer func() {
		db.Drop(&bench{})
		closeDB(a)
	}()

	// 构造数据
	a.NotError(db.Create(&bench{}))
	a.NotError(db.Insert(m))

	be := &bench{Name: "name"}
	for i := 0; i < b.N; i++ {
		count, _ := db.Count(be)
		if count < 1 {
			b.Error("count:", count)
		}
	}
}
