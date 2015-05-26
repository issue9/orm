// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package test

import (
	"testing"

	"github.com/issue9/assert"
)

func BenchmarkDB_Insert(b *testing.B) {
	a := assert.New(b)

	m := &bench{
		Name: "name",
		Pass: "pass",
	}

	db := newDB(a)
	defer func() {
		closeDB(a)
	}()

	a.NotError(db.Create(&bench{}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.NotError(db.Insert(m))
	}
}

func BenchmarkDB_MultInsert(b *testing.B) {
	a := assert.New(b)

	m := &bench{
		Name: "name",
		Pass: "pass",
	}

	ms := make([]interface{}, 0, 100)
	for i := 0; i < len(ms); i++ {
		ms = append(ms, m)
	}

	db := newDB(a)
	defer func() {
		closeDB(a)
	}()

	b.ResetTimer()
	a.NotError(db.Create(&bench{}))
	for i := 0; i < b.N; i++ {
		a.NotError(db.Insert(ms...))
	}
}
