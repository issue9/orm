// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"testing"

	"github.com/issue9/assert"
)

var _ base = &Postgres{}

var p = &Postgres{}

func TestPostgresGetDBName(t *testing.T) {
	a := assert.New(t)

	a.Equal(p.GetDBName("user=abc dbname = dbname password=abc"), "dbname")
	a.Equal(p.GetDBName("dbname=\tdbname user=abc"), "dbname")
	a.Equal(p.GetDBName("dbname=dbname\tuser=abc"), "dbname")
	a.Equal(p.GetDBName("\tdbname=dbname user=abc"), "dbname")
	a.Equal(p.GetDBName("\tdbname = dbname user=abc"), "dbname")
}
