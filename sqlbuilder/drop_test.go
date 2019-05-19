// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/sqltest"
)

var _ SQLer = &UpdateStmt{}

func TestDropTable(t *testing.T) {
	a := assert.New(t)

	drop := DropTable(nil).
		Table("table").
		Table("tbl2")
	sql, args, err := drop.SQL()
	a.NotError(err).Nil(args)
	sqltest.Equal(a, sql, "drop table if exists tbl2")

	drop.Reset()
	sql, args, err = drop.SQL()
	a.Equal(err, ErrTableIsEmpty).Nil(args).Empty(sql)
}
