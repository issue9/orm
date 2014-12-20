// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"testing"

	"github.com/issue9/assert"
)

func TestMysqlLimitSQL(t *testing.T) {
	a := assert.New(t)

	sql, args := mysqlLimitSQL(5, 0)
	a.StringEqual(sql, " LIMIT ? OFFSET ? ", style).
		Equal(args, []interface{}{5, 0})

	sql, args = mysqlLimitSQL(5)
	a.StringEqual(sql, "LIMIT ?", style).
		Equal(args, []interface{}{5})
}

func TestOracleLimitSQL(t *testing.T) {
	a := assert.New(t)

	sql, args := oracleLimitSQL(5, 0)
	a.StringEqual(sql, " OFFSET ? ROWS FETCH NEXT ? ROWS ONLY ", style).
		Equal(args, []interface{}{0, 5})

	sql, args = oracleLimitSQL(5)
	a.StringEqual(sql, "FETCH NEXT ? ROWS ONLY ", style).
		Equal(args, []interface{}{5})
}
