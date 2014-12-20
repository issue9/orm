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

	sql := `SELECT id FROM user WHERE id<2 ORDER BY id`
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err := Column(false, "id", rows)
	a.NotError(err).NotNil(cols)

	a.Equal([]interface{}{0, 1}, cols)
	a.NotError(rows.Close())

	// 读取一行
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err = Column(true, "id", rows)
	a.NotError(err).NotNil(cols)

	a.Equal([]interface{}{0}, cols)
	a.NotError(rows.Close())
}

func TestColumnString(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer closeDB(db, a)

	sql := `SELECT id FROM user WHERE id<2 ORDER BY id`
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err := ColumnString(false, "id", rows)
	a.NotError(err).NotNil(cols)

	a.Equal([]string{"0", "1"}, cols)
	a.NotError(rows.Close())

	// 读取一行
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	cols, err = ColumnString(true, "id", rows)
	a.NotError(err).NotNil(cols)

	a.Equal([]string{"0"}, cols)
	a.NotError(rows.Close())
}
