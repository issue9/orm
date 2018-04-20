// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package fetch

import (
	"testing"

	"github.com/issue9/assert"
)

func TestColumn(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer closeDB(db, a)

	// 正常数据匹配，读取多行
	sql := `SELECT id,email FROM user WHERE id<2 ORDER BY id`
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err := Column(false, "id", rows)
	a.NotError(err).NotNil(cols)

	a.Equal([]interface{}{0, 1}, cols)
	a.NotError(rows.Close())

	// 正常数据匹配，读取一行
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err = Column(true, "id", rows)
	a.NotError(err).NotNil(cols)

	a.Equal([]interface{}{0}, cols)
	a.NotError(rows.Close())

	// 没有数据匹配，读取多行
	sql = `SELECT id,email FROM user WHERE id<0 ORDER BY id`
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err = Column(false, "id", rows)
	a.NotError(err)

	a.Empty(cols)
	a.NotError(rows.Close())

	// 没有数据匹配，读取一行
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err = Column(true, "id", rows)
	a.NotError(err)

	a.Empty(cols)
	a.NotError(rows.Close())

	// 指定错误的列名
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err = Column(true, "not-exists", rows)
	a.Error(err)

	a.Empty(cols)
	a.NotError(rows.Close())
}

func TestColumnString(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer closeDB(db, a)

	// 正常数据匹配，读取多行
	sql := `SELECT id,email FROM user WHERE id<2 ORDER BY id`
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err := ColumnString(false, "id", rows)
	a.NotError(err).NotNil(cols)

	a.Equal([]string{"0", "1"}, cols)
	a.NotError(rows.Close())

	// 正常数据匹配，读取一行
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err = ColumnString(true, "id", rows)
	a.NotError(err).NotNil(cols)

	a.Equal([]string{"0"}, cols)
	a.NotError(rows.Close())

	// 没有数据匹配，读取多行
	sql = `SELECT id FROM user WHERE id<0 ORDER BY id`
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err = ColumnString(false, "id", rows)
	a.NotError(err)

	a.Empty(cols)
	a.NotError(rows.Close())

	// 没有数据匹配，读取一行
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err = ColumnString(true, "id", rows)
	a.NotError(err)

	a.Empty(cols)
	a.NotError(rows.Close())

	// 指定错误的列名
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err = ColumnString(true, "not-exists", rows)
	a.Error(err)

	a.Empty(cols)
	a.NotError(rows.Close())
}
