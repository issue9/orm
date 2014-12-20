// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"testing"

	"github.com/issue9/assert"
)

var _ base = &Sqlite3{}

var s = &Sqlite3{}

func TestSqlite3GetDBName(t *testing.T) {
	a := assert.New(t)

	a.Equal(s.GetDBName("./dbname.db"), "dbname")
	a.Equal(s.GetDBName("./dbname"), "dbname")
	a.Equal(s.GetDBName("abc/dbname.abc"), "dbname")
	a.Equal(s.GetDBName("dbname"), "dbname")
	a.Equal(s.GetDBName(""), "")
}
