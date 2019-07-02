// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/sqltest"
)

var _ DDLSQLer = &DropTableStmt{}

func TestDropTable(t *testing.T) {
	a := assert.New(t)

	drop := DropTable(nil).
		Table("table").
		Table("tbl2")
	sql, err := drop.DDLSQL()
	a.NotError(err).
		Equal(2, len(sql))
	sqltest.Equal(a, sql[0], "drop table if exists table")
	sqltest.Equal(a, sql[1], "drop table if exists tbl2")

	drop.Reset()
	sql, err = drop.DDLSQL()
	a.Equal(err, ErrTableIsEmpty).Empty(sql)
}
