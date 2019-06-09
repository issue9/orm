// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package fetch_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/fetch"
	"github.com/issue9/orm/v2/internal/testconfig"
)

func TestMap(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer testconfig.CloseDB(db, a)

	// 正常匹配数据，读取多行
	sql := `SELECT id,Email FROM #user WHERE id<2 ORDER BY id`
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)

	mapped, err := fetch.Map(false, rows)
	a.NotError(err).NotNil(mapped)

	a.Equal([]map[string]interface{}{
		map[string]interface{}{"id": 0, "Email": "email-0"},
		map[string]interface{}{"id": 1, "Email": "email-1"},
	}, mapped)
	a.NotError(rows.Close())

	// 正常匹配数据，读取一行
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	mapped, err = fetch.Map(true, rows)
	a.NotError(err).NotNil(mapped)

	a.Equal([]map[string]interface{}{
		map[string]interface{}{"id": 0, "Email": "email-0"},
	}, mapped)
	a.NotError(rows.Close())

	// 没有匹配的数据，读取多行
	sql = `SELECT id,Email FROM #user WHERE id<0 ORDER BY id`
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
	defer testconfig.CloseDB(db, a)

	// 正常数据匹配，读取多行
	sql := `SELECT id,Email FROM #user WHERE id<2 ORDER BY id`
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)

	mapped, err := fetch.MapString(false, rows)
	a.NotError(err).NotNil(mapped)

	a.Equal(mapped, []map[string]string{
		map[string]string{"id": "0", "Email": "email-0"},
		map[string]string{"id": "1", "Email": "email-1"},
	})
	a.NotError(rows.Close())

	// 正常数据匹配，读取一行
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	mapped, err = fetch.MapString(true, rows)
	a.NotError(err).NotNil(mapped)

	a.Equal(mapped, []map[string]string{
		map[string]string{"id": "0", "Email": "email-0"},
	})
	a.NotError(rows.Close())

	// 没有数据匹配，读取多行
	sql = `SELECT id,Email FROM #user WHERE id<0 ORDER BY id`
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
