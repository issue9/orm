// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package fetch_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/fetch"
)

func TestColumn(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer clearDB(a, db)

	// 正常数据匹配，读取多行
	sql := `SELECT id,email FROM #user WHERE id<3 ORDER BY id ASC`
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err := fetch.Column(false, "id", rows)
	a.NotError(err).NotNil(cols)

	// mysql 返回的是 []byte 类型
	ok := assert.IsEqual([]interface{}{1, 2}, cols) ||
		assert.IsEqual([]interface{}{[]byte{'1'}, []byte{'2'}}, cols)
	a.True(ok)
	a.NotError(rows.Close())

	// 正常数据匹配，读取一行
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err = fetch.Column(true, "id", rows)
	a.NotError(err).NotNil(cols)

	ok = assert.IsEqual([]interface{}{1}, cols) ||
		assert.IsEqual([]interface{}{[]byte{'1'}}, cols)
	a.True(ok)
	a.NotError(rows.Close())

	// 没有数据匹配，读取多行
	sql = `SELECT id,email FROM #user WHERE id<0 ORDER BY id ASC`
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err = fetch.Column(false, "id", rows)
	a.NotError(err)

	a.Empty(cols)
	a.NotError(rows.Close())

	// 没有数据匹配，读取一行
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err = fetch.Column(true, "id", rows)
	a.NotError(err)

	a.Empty(cols)
	a.NotError(rows.Close())

	// 指定错误的列名
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err = fetch.Column(true, "not-exists", rows)
	a.Error(err)

	a.Empty(cols)
	a.NotError(rows.Close())
}

func TestColumnString(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer clearDB(a, db)

	// 正常数据匹配，读取多行
	sql := `SELECT id,email FROM #user WHERE id<3 ORDER BY id`
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err := fetch.ColumnString(false, "id", rows)
	a.NotError(err).NotNil(cols)

	a.Equal([]string{"1", "2"}, cols)
	a.NotError(rows.Close())

	// 正常数据匹配，读取一行
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err = fetch.ColumnString(true, "id", rows)
	a.NotError(err).NotNil(cols)

	a.Equal([]string{"1"}, cols)
	a.NotError(rows.Close())

	// 没有数据匹配，读取多行
	sql = `SELECT id FROM #user WHERE id<0 ORDER BY id`
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err = fetch.ColumnString(false, "id", rows)
	a.NotError(err)

	a.Empty(cols)
	a.NotError(rows.Close())

	// 没有数据匹配，读取一行
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err = fetch.ColumnString(true, "id", rows)
	a.NotError(err)

	a.Empty(cols)
	a.NotError(rows.Close())

	// 指定错误的列名
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err = fetch.ColumnString(true, "not-exists", rows)
	a.Error(err)

	a.Empty(cols)
	a.NotError(rows.Close())
}
