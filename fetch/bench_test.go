// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package fetch_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/fetch"
	"github.com/issue9/orm/v2/internal/test"
)

func BenchmarkObject(b *testing.B) {
	a := assert.New(b)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		sql := `SELECT id,Email FROM user WHERE id<2 ORDER BY id`
		objs := []*FetchUser{
			{},
			{},
		}

		for i := 0; i < b.N; i++ {
			rows, err := t.DB.Query(sql)
			t.NotError(err)

			cnt, err := fetch.Object(true, rows, &objs)
			t.NotError(err).NotEmpty(cnt)
			t.NotError(rows.Close())
		}
	})
}

func BenchmarkMap(b *testing.B) {
	a := assert.New(b)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		// 正常匹配数据，读取多行
		sql := `SELECT id,Email FROM user WHERE id<2 ORDER BY id`

		for i := 0; i < b.N; i++ {
			rows, err := t.DB.Query(sql)
			t.NotError(err)

			mapped, err := fetch.Map(false, rows)
			t.NotError(err).NotNil(mapped)
			t.NotError(rows.Close())
		}
	})
}
