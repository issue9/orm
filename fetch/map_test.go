// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package fetch_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/fetch"
	"github.com/issue9/orm/v2/internal/test"
)

func TestMap(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		db := t.DB

		// 正常匹配数据，读取多行
		sql := `SELECT id,email FROM #user WHERE id<3 ORDER BY id`
		rows, err := db.Query(sql)
		t.NotError(err).NotNil(rows)

		mapped, err := fetch.Map(false, rows)
		t.NotError(err).NotNil(mapped)

		ok := assert.IsEqual([]map[string]interface{}{
			{"id": 1, "email": "email-1"},
			{"id": 2, "email": "email-2"},
		}, mapped) ||
			assert.IsEqual([]map[string]interface{}{
				{"id": []byte{'1'}, "email": []byte("email-1")},
				{"id": []byte{'2'}, "email": []byte("email-2")},
			}, mapped)
		t.True(ok)
		t.NotError(rows.Close())

		// 正常匹配数据，读取一行
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		mapped, err = fetch.Map(true, rows)
		t.NotError(err).NotNil(mapped)

		ok = assert.IsEqual([]map[string]interface{}{
			{"id": 1, "email": "email-1"},
		}, mapped) ||
			assert.IsEqual([]map[string]interface{}{
				{"id": []byte{'1'}, "email": []byte("email-1")},
			}, mapped)
		t.True(ok)
		t.NotError(rows.Close())

		// 没有匹配的数据，读取多行
		sql = `SELECT id,email FROM #user WHERE id<0 ORDER BY id`
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		mapped, err = fetch.Map(false, rows)
		t.NotError(err)

		t.Equal([]map[string]interface{}{}, mapped)
		t.NotError(rows.Close())

		// 没有匹配的数据，读取一行
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		mapped, err = fetch.Map(true, rows)
		t.NotError(err)

		t.Equal([]map[string]interface{}{}, mapped)
		t.NotError(rows.Close())
	})
}

func TestMapString(t *testing.T) {
	a := assert.New(t)
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

		mapped, err := fetch.MapString(false, rows)
		t.NotError(err).NotNil(mapped)

		t.Equal(mapped, []map[string]string{
			{"id": "1", "email": "email-1"},
			{"id": "2", "email": "email-2"},
		})
		t.NotError(rows.Close())

		// 正常数据匹配，读取一行
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		mapped, err = fetch.MapString(true, rows)
		t.NotError(err).NotNil(mapped)

		t.Equal(mapped, []map[string]string{
			{"id": "1", "email": "email-1"},
		})
		t.NotError(rows.Close())

		// 没有数据匹配，读取多行
		sql = `SELECT id,email FROM #user WHERE id<0 ORDER BY id`
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		mapped, err = fetch.MapString(false, rows)
		t.NotError(err)

		t.Equal(mapped, []map[string]string{})
		t.NotError(rows.Close())

		// 没有数据匹配，读取一行
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		mapped, err = fetch.MapString(true, rows)
		t.NotError(err)

		t.Equal(mapped, []map[string]string{})
		t.NotError(rows.Close())
	})
}
