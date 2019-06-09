// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package fetch_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/fetch"
	"github.com/issue9/orm/v2/internal/testconfig"
)

func BenchmarkObject(b *testing.B) {
	a := assert.New(b)
	db := initDB(a)
	defer testconfig.CloseDB(db, a)

	sql := `SELECT id,Email FROM user WHERE id<2 ORDER BY id`
	objs := []*FetchUser{
		&FetchUser{},
		&FetchUser{},
	}

	for i := 0; i < b.N; i++ {
		rows, err := db.Query(sql)
		a.NotError(err)

		cnt, err := fetch.Object(true, rows, &objs)
		a.NotError(err).NotEmpty(cnt)
		rows.Close()
	}
}

func BenchmarkMap(b *testing.B) {
	a := assert.New(b)
	db := initDB(a)
	defer testconfig.CloseDB(db, a)

	// 正常匹配数据，读取多行
	sql := `SELECT id,Email FROM user WHERE id<2 ORDER BY id`

	for i := 0; i < b.N; i++ {
		rows, err := db.Query(sql)
		a.NotError(err)

		mapped, err := fetch.Map(false, rows)
		a.NotError(err).NotNil(mapped)
		rows.Close()
	}
}
