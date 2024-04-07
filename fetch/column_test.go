// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package fetch_test

import (
	"reflect"
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/orm/v6/fetch"
	"github.com/issue9/orm/v6/internal/test"
)

func TestColumn(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	eq := func(s1, s2 []any) bool {
		if len(s1) != len(s2) {
			return false
		}
		for i, v := range s1 {
			if !reflect.DeepEqual(v, s2[i]) {
				return false
			}
		}
		return true
	}

	suite.Run(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)
		db := t.DB

		// 正常数据匹配，读取多行
		sql := `SELECT id,email FROM fetch_users WHERE id<3 ORDER BY id ASC`
		rows, err := db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err := fetch.Column(false, "id", rows)
		t.NotError(err).NotNil(cols)

		if t.DriverName == "mysql" { // mysql 返回的是 []byte 类型
			eq(cols, []any{[]byte{'1'}, []byte{'2'}})
		} else {
			eq(cols, []any{int64(1), int64(2)})
		}
		t.NotError(rows.Close())

		// 正常数据匹配，读取一行
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err = fetch.Column(true, "id", rows)
		t.NotError(err).NotNil(cols)

		if t.DriverName == "mysql" { // mysql 返回的是 []byte 类型
			eq([]any{[]byte{'1'}}, cols)
		} else {
			eq([]any{int64(1)}, cols)
		}
		t.NotError(rows.Close())

		// 没有数据匹配，读取多行
		sql = `SELECT id,email FROM fetch_users WHERE id<0 ORDER BY id ASC`
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err = fetch.Column(false, "id", rows)
		t.NotError(err)

		t.Empty(cols)
		t.NotError(rows.Close())

		// 没有数据匹配，读取一行
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err = fetch.Column(true, "id", rows)
		t.NotError(err)

		t.Empty(cols)
		t.NotError(rows.Close())

		// 指定错误的列名
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err = fetch.Column(true, "not-exists", rows)
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

		cols, err := fetch.ColumnString(false, "id", rows)
		t.NotError(err).NotNil(cols)

		t.Equal([]string{"1", "2"}, cols)
		t.NotError(rows.Close())

		// 正常数据匹配，读取一行
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err = fetch.ColumnString(true, "id", rows)
		t.NotError(err).NotNil(cols)

		t.Equal([]string{"1"}, cols)
		t.NotError(rows.Close())

		// 没有数据匹配，读取多行
		sql = `SELECT id FROM fetch_users WHERE id<0 ORDER BY id`
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err = fetch.ColumnString(false, "id", rows)
		t.NotError(err)

		t.Empty(cols)
		t.NotError(rows.Close())

		// 没有数据匹配，读取一行
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err = fetch.ColumnString(true, "id", rows)
		t.NotError(err)

		t.Empty(cols)
		t.NotError(rows.Close())

		// 指定错误的列名
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err = fetch.ColumnString(true, "not-exists", rows)
		t.Error(err)

		t.Empty(cols)
		t.NotError(rows.Close())
	})
}
