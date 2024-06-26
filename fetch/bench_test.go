// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package fetch_test

import (
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/orm/v6/fetch"
	"github.com/issue9/orm/v6/internal/test"
)

func BenchmarkObject(b *testing.B) {
	a := assert.New(b, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		sql := `SELECT id,Email FROM fetch_users WHERE id<2 ORDER BY id`
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
	a := assert.New(b, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		// 正常匹配数据，读取多行
		sql := `SELECT id,Email FROM fetch_users WHERE id<2 ORDER BY id`

		for i := 0; i < b.N; i++ {
			rows, err := t.DB.Query(sql)
			t.NotError(err)

			mapped, err := fetch.Map(false, rows)
			t.NotError(err).NotNil(mapped)
			t.NotError(rows.Close())
		}
	})
}
