// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package fetch

import (
	"testing"

	"github.com/issue9/assert"
)

func TestMap(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer closeDB(db, a)

	sql := `SELECT id,Email FROM user WHERE id<2 ORDER BY id`
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)

	mapped, err := Map(false, rows)
	a.NotError(err).NotNil(mapped)

	a.Equal([]map[string]interface{}{
		map[string]interface{}{"id": 0, "Email": "email-0"},
		map[string]interface{}{"id": 1, "Email": "email-1"},
	}, mapped)
	a.NotError(rows.Close())

	// 读取一行
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	mapped, err = Map(true, rows)
	a.NotError(err).NotNil(mapped)

	a.Equal([]map[string]interface{}{
		map[string]interface{}{"id": 0, "Email": "email-0"},
	}, mapped)
	a.NotError(rows.Close())
}

func TestMapString(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer closeDB(db, a)

	sql := `SELECT id,Email FROM user WHERE id<2 ORDER BY id`
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)

	mapped, err := MapString(false, rows)
	a.NotError(err).NotNil(mapped)

	a.Equal(mapped, []map[string]string{
		map[string]string{"id": "0", "Email": "email-0"},
		map[string]string{"id": "1", "Email": "email-1"},
	})

	a.NotError(rows.Close())

	// 读取一行
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	mapped, err = MapString(true, rows)
	a.NotError(err).NotNil(mapped)

	a.Equal(mapped, []map[string]string{
		map[string]string{"id": "0", "Email": "email-0"},
	})
	a.NotError(rows.Close())
}
