// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"testing"

	"github.com/issue9/assert"
)

// mysql: BenchmarkDB_Insert	    5000	    234033 ns/op
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

// mysql: BenchmarkDB_Update	    5000	    290209 ns/op
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

// mysql: BenchmarkDB_Select	   10000	    181897 ns/op
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

// mysql: BenchmarkDB_WhereUpdate	   10000	    169674 ns/op
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
		a.NotError(db.Where("id=?", i+1).Table("#bench").Update(true, map[string]interface{}{
			"name": "n1",
			"pass": "p1",
			"site": "s1",
		}))
	}
}

// mysql: BenchmarkDB_Count	   10000	    168311 ns/op
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
