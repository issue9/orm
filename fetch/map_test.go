// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package fetch_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/fetch"
)

func TestMap(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer clearDB(a, db)

	// 正常匹配数据，读取多行
	sql := `SELECT id,email FROM #user WHERE id<3 ORDER BY id`
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)

	mapped, err := fetch.Map(false, rows)
	a.NotError(err).NotNil(mapped)

	ok := assert.IsEqual([]map[string]interface{}{
		map[string]interface{}{"id": 1, "email": "email-1"},
		map[string]interface{}{"id": 2, "email": "email-2"},
	}, mapped) ||
		assert.IsEqual([]map[string]interface{}{
			map[string]interface{}{"id": []byte{'1'}, "email": []byte("email-1")},
			map[string]interface{}{"id": []byte{'2'}, "email": []byte("email-2")},
		}, mapped)
	a.True(ok)
	a.NotError(rows.Close())

	// 正常匹配数据，读取一行
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	mapped, err = fetch.Map(true, rows)
	a.NotError(err).NotNil(mapped)

	ok = assert.IsEqual([]map[string]interface{}{
		map[string]interface{}{"id": 1, "email": "email-1"},
	}, mapped) ||
		assert.IsEqual([]map[string]interface{}{
			map[string]interface{}{"id": []byte{'1'}, "email": []byte("email-1")},
		}, mapped)
	a.True(ok)
	a.NotError(rows.Close())

	// 没有匹配的数据，读取多行
	sql = `SELECT id,email FROM #user WHERE id<0 ORDER BY id`
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	mapped, err = fetch.Map(false, rows)
	a.NotError(err)

	a.Equal([]map[string]interface{}{}, mapped)
	a.NotError(rows.Close())

	// 没有匹配的数据，读取一行
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	mapped, err = fetch.Map(true, rows)
	a.NotError(err)

	a.Equal([]map[string]interface{}{}, mapped)
	a.NotError(rows.Close())
}

func TestMapString(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer clearDB(a, db)

	// 正常数据匹配，读取多行
	sql := `SELECT id,email FROM #user WHERE id<3 ORDER BY id`
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)

	mapped, err := fetch.MapString(false, rows)
	a.NotError(err).NotNil(mapped)

	a.Equal(mapped, []map[string]string{
		map[string]string{"id": "1", "email": "email-1"},
		map[string]string{"id": "2", "email": "email-2"},
	})
	a.NotError(rows.Close())

	// 正常数据匹配，读取一行
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	mapped, err = fetch.MapString(true, rows)
	a.NotError(err).NotNil(mapped)

	a.Equal(mapped, []map[string]string{
		map[string]string{"id": "1", "email": "email-1"},
	})
	a.NotError(rows.Close())

	// 没有数据匹配，读取多行
	sql = `SELECT id,email FROM #user WHERE id<0 ORDER BY id`
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	mapped, err = fetch.MapString(false, rows)
	a.NotError(err)

	a.Equal(mapped, []map[string]string{})
	a.NotError(rows.Close())

	// 没有数据匹配，读取一行
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	mapped, err = fetch.MapString(true, rows)
	a.NotError(err)

	a.Equal(mapped, []map[string]string{})
	a.NotError(rows.Close())
}
