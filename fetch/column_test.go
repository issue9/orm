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

func TestColumn(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)
		db := t.DB

		// 正常数据匹配，读取多行
		sql := `SELECT id,email FROM fetch_users WHERE id<3 ORDER BY id ASC`
		rows, err := db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err := fetch.Column[int64](false, "id", rows)
		t.NotError(err).NotNil(cols)

		t.Equal(cols, []int64{int64(1), int64(2)})
		t.NotError(rows.Close())

		// 正常数据匹配，读取一行
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err = fetch.Column[int64](true, "id", rows)
		t.NotError(err).NotNil(cols)

		t.Equal(cols, []int64{int64(1)})
		t.NotError(rows.Close())

		// 没有数据匹配，读取多行
		sql = `SELECT id,email FROM fetch_users WHERE id<0 ORDER BY id ASC`
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err = fetch.Column[int64](false, "id", rows)
		t.NotError(err)

		t.Empty(cols)
		t.NotError(rows.Close())

		// 没有数据匹配，读取一行
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err = fetch.Column[int64](true, "id", rows)
		t.NotError(err)

		t.Empty(cols)
		t.NotError(rows.Close())

		// 指定错误的列名
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err = fetch.Column[int64](true, "not-exists", rows)
		t.Error(err)

		t.Empty(cols)
		t.NotError(rows.Close())
	})
}

func TestColumnString(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)
		db := t.DB

		// 正常数据匹配，读取多行
		sql := `SELECT id,email FROM fetch_users WHERE id<3 ORDER BY id`
		rows, err := db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err := fetch.Column[string](false, "id", rows)
		t.NotError(err).NotNil(cols)

		t.Equal([]string{"1", "2"}, cols)
		t.NotError(rows.Close())

		// 正常数据匹配，读取一行
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err = fetch.Column[string](true, "id", rows)
		t.NotError(err).NotNil(cols)

		t.Equal([]string{"1"}, cols)
		t.NotError(rows.Close())

		// 没有数据匹配，读取多行
		sql = `SELECT id FROM fetch_users WHERE id<0 ORDER BY id`
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err = fetch.Column[string](false, "id", rows)
		t.NotError(err)

		t.Empty(cols)
		t.NotError(rows.Close())

		// 没有数据匹配，读取一行
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err = fetch.Column[string](true, "id", rows)
		t.NotError(err)

		t.Empty(cols)
		t.NotError(rows.Close())

		// 指定错误的列名
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err = fetch.Column[string](true, "not-exists", rows)
		t.Error(err)

		t.Empty(cols)
		t.NotError(rows.Close())
	})
}
