// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package test

import (
	"testing"

	"github.com/issue9/assert"
)

// BenchmarkDB_Insert	    5000	    234033 ns/op
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

// BenchmarkDB_MultInsert	      50	  23436852 ns/op
func BenchmarkDB_MultInsert(b *testing.B) {
	a := assert.New(b)

	ms := make([]interface{}, 0, 100)
	for i := 0; i < cap(ms); i++ {
		ms = append(ms, &bench{
			Name: "name",
			Pass: "pass",
			Site: "http://www.github.com/issue9/orm",
		})
	}

	db := newDB(a)
	defer func() {
		db.Drop(&bench{})
		closeDB(a)
	}()

	a.NotError(db.Create(&bench{}))

	for i := 0; i < b.N; i++ {
		a.NotError(db.Insert(ms...))
	}
}

// BenchmarkDB_Update	    5000	    290209 ns/op
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

// BenchmarkDB_MultUpdate	      50	  28662471 ns/op
func BenchmarkDB_MultUpdate(b *testing.B) {
	a := assert.New(b)

	ms := make([]interface{}, 0, 100)
	for i := 0; i < cap(ms); i++ {
		ms = append(ms, &bench{
			Name: "name",
			Pass: "pass",
			Site: "http://www.github.com/issue9/orm",
		})
	}

	db := newDB(a)
	defer func() {
		db.Drop(&bench{})
		closeDB(a)
	}()

	// 构造数据
	a.NotError(db.Create(&bench{}))
	a.NotError(db.Insert(ms...))

	i := 0
	for _, m := range ms {
		i++
		m.(*bench).ID = i
	}

	for i := 0; i < b.N; i++ {
		a.NotError(db.Update(ms...))
	}
}

// BenchmarkDB_Select	   10000	    181897 ns/op
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

// BenchmarkDB_MultSelect	     100	  18494929 ns/op
func BenchmarkDB_MultSelect(b *testing.B) {
	a := assert.New(b)

	ms := make([]interface{}, 0, 100)
	for i := 0; i < cap(ms); i++ {
		ms = append(ms, &bench{
			Name: "name",
			Pass: "pass",
			Site: "http://www.github.com/issue9/orm",
		})
	}

	db := newDB(a)
	defer func() {
		db.Drop(&bench{})
		closeDB(a)
	}()

	// 构造数据
	a.NotError(db.Create(&bench{}))
	a.NotError(db.Insert(ms...))

	i := 0
	for _, m := range ms {
		i++
		m.(*bench).ID = i
	}

	for i := 0; i < b.N; i++ {
		a.NotError(db.Select(ms...))
	}
}
