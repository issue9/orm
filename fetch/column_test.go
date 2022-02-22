// SPDX-License-Identifier: MIT

package fetch_test

import (
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/orm/v4/fetch"
	"github.com/issue9/orm/v4/internal/test"
)

func TestColumn(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)
		db := t.DB

		// 正常数据匹配，读取多行
		sql := `SELECT id,email FROM #user WHERE id<3 ORDER BY id ASC`
		rows, err := db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err := fetch.Column(false, "id", rows)
		t.NotError(err).NotNil(cols)

		// mysql 返回的是 []byte 类型
		ok := assert.IsEqual([]any{1, 2}, cols) ||
			assert.IsEqual([]any{[]byte{'1'}, []byte{'2'}}, cols)
		t.True(ok)
		t.NotError(rows.Close())

		// 正常数据匹配，读取一行
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		cols, err = fetch.Column(true, "id", rows)
		t.NotError(err).NotNil(cols)

		ok = assert.IsEqual([]any{1}, cols) ||
			assert.IsEqual([]any{[]byte{'1'}}, cols)
		t.True(ok)
		t.NotError(rows.Close())

		// 没有数据匹配，读取多行
		sql = `SELECT id,email FROM #user WHERE id<0 ORDER BY id ASC`
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
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)
		db := t.DB

		// 正常数据匹配，读取多行
		sql := `SELECT id,email FROM #user WHERE id<3 ORDER BY id`
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
		sql = `SELECT id FROM #user WHERE id<0 ORDER BY id`
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
