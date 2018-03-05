// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"database/sql"
	"testing"

	"github.com/issue9/orm/internal/sqltest"

	"github.com/issue9/assert"
)

var (
	_ SQLer       = &SelectStmt{}
	_ WhereStmter = &SelectStmt{}
	_ queryer     = &SelectStmt{}
)

func TestSelect(t *testing.T) {
	a := assert.New(t)
	s := Select(nil).Select("c1", "column2 as c2", "c3").
		From("table").
		And("c1=?", 1).
		Or("c2=?", sql.Named("c2", 2)).
		Limit(10, 0).
		Desc("c1")
	a.NotNil(s)
	query, args, err := s.SQL()
	a.NotError(err)
	a.Equal(args, []interface{}{1, sql.Named("c2", 2)})
	sqltest.Equal(a, query, "select c1,colun2 as c2,c3 from table where c1=? and c2=@c2 order by c1 desc limit 10,0")

	// count
	s.Count("count(*) as cnt")
	query, args, err = s.SQL()
	a.NotError(err)
	a.Equal(args, []interface{}{1, sql.Named("c2", 2)})
	sqltest.Equal(a, query, "select count(*) as cnt from table where c1=? and c2=@c2 order by c1 desc")
}
