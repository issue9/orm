// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package testconfig

import (
	"testing"

	"github.com/issue9/assert"
)

func TestNewDB(t *testing.T) {
	a := assert.New(t)

	driver = "sqlite3"
	db := NewDB(a)
	CloseDB(db, a)

	driver = "not-exists-driver"
	a.Panic(func() {
		db = NewDB(a)
	})
}
