// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"github.com/issue9/assert"
)

func newDB(a *assert.Assertion) *DB {
	db, err := NewDB("sqlite3", "./test.db", "sqlite3_", &sqlite3{})
	a.NotError(err).NotNil(db)
	return db
}
