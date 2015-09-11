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

// mysql: BenchmarkDB_MultInsert	      50	  23436852 ns/op
func BenchmarkTx_MultInsert(b *testing.B) {
	a := assert.New(b)

	ms := make([]interface{}, 0, 100)
	for i := 0; i < cap(ms); i++ {
		ms = append(ms, &bench{
			Name: "name",
			Pass: "pass",
			Site: "http://www.github.com/issue9/orm",
		})
	}

	tx, err := newDB(a).Begin()
	a.NotError(err)
	defer func() {
		tx.Drop(&bench{})
		closeDB(a)
	}()

	a.NotError(tx.Create(&bench{}))

	for i := 0; i < b.N; i++ {
		a.NotError(tx.Insert(ms...))
	}
}

// mysql: BenchmarkDB_InsertMany	     500	   2349282 ns/op
func BenchmarkTx_InsertMany(b *testing.B) {
	a := assert.New(b)

	ms := make([]*bench, 0, 100)
	for i := 0; i < cap(ms); i++ {
		ms = append(ms, &bench{
			Name: "name",
			Pass: "pass",
			Site: "http://www.github.com/issue9/orm",
		})
	}

	tx, err := newDB(a).Begin()
	a.NotError(err)
	defer func() {
		tx.Drop(&bench{})
		closeDB(a)
	}()

	a.NotError(tx.Create(&bench{}))

	for i := 0; i < b.N; i++ {
		a.NotError(tx.InsertMany(ms))
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

// mysql: BenchmarkDB_MultUpdate	      50	  28662471 ns/op
func BenchmarkTx_MultUpdate(b *testing.B) {
	a := assert.New(b)

	ms := make([]interface{}, 0, 100)
	for i := 0; i < cap(ms); i++ {
		ms = append(ms, &bench{
			Name: "name",
			Pass: "pass",
			Site: "http://www.github.com/issue9/orm",
		})
	}

	tx, err := newDB(a).Begin()
	a.NotError(err)
	defer func() {
		tx.Drop(&bench{})
		closeDB(a)
	}()

	// 构造数据
	a.NotError(tx.Create(&bench{}))
	a.NotError(tx.Insert(ms...))

	i := 0
	for _, m := range ms {
		i++
		m.(*bench).ID = i
	}

	for i := 0; i < b.N; i++ {
		a.NotError(tx.Update(ms...))
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

// mysql: BenchmarkDB_MultSelect	     100	  18494929 ns/op
func BenchmarkTx_MultSelect(b *testing.B) {
	a := assert.New(b)

	ms := make([]interface{}, 0, 100)
	for i := 0; i < cap(ms); i++ {
		ms = append(ms, &bench{
			Name: "name",
			Pass: "pass",
			Site: "http://www.github.com/issue9/orm",
		})
	}

	tx, err := newDB(a).Begin()
	a.NotError(err)
	defer func() {
		tx.Drop(&bench{})
		closeDB(a)
	}()

	// 构造数据
	a.NotError(tx.Create(&bench{}))
	a.NotError(tx.Insert(ms...))

	i := 0
	for _, m := range ms {
		i++
		m.(*bench).ID = i
	}

	for i := 0; i < b.N; i++ {
		a.NotError(tx.Select(ms...))
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

// mysql: BenchmarkDB_WhereSelect	   10000	    182159 ns/op
func BenchmarkTx_WhereSelect(b *testing.B) {
	a := assert.New(b)

	ms := make([]interface{}, 0, 100)
	for i := 0; i < cap(ms); i++ {
		ms = append(ms, &bench{
			Name: "name",
			Pass: "pass",
			Site: "http://www.github.com/issue9/orm",
		})
	}

	tx, err := newDB(a).Begin()
	a.NotError(err)
	defer func() {
		tx.Drop(&bench{})
		closeDB(a)
	}()

	// 构造数据
	a.NotError(tx.Create(&bench{}))
	a.NotError(tx.Insert(ms...))

	models := make([]*bench, 0, 100)
	for i := 0; i < cap(models); i++ {
		models = append(models, &bench{
			Name: "name",
			Pass: "pass",
			Site: "http://www.github.com/issue9/orm",
		})
	}
	for i := 0; i < b.N; i++ {
		a.NotError(tx.Where("id>?", i).Table("#bench").Select(true, models))
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
