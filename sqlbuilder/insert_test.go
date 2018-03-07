// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"database/sql"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/internal/sqltest"
)

var _ SQLer = &InsertStmt{}

func TestInsert(t *testing.T) {
	a := assert.New(t)
	i := Insert(nil, "table")
	a.NotNil(i)

	i.Columns("c1", "c2", "c3").Values(1, 2, 3).Values(4, 5, 6)
	query, args, err := i.SQL()
	a.NotError(err)
	a.Equal(args, []interface{}{1, 2, 3, 4, 5, 6})
	sqltest.Equal(a, query, "insert into table (c1,c2,c3) values (?,?,?),(?,?,?)")

	i.Reset()
	a.Empty(i.cols).Empty(i.args)
	i.Table("tb1").
		Table("tb2").
		Columns("c1", "c2").
		Values(1, 2).
		Values(3, sql.Named("c2", 4))
	query, args, err = i.SQL()
	a.NotError(err)
	a.Equal(args, []interface{}{1, 2, 3, sql.Named("c2", 4)})
	sqltest.Equal(a, query, "insert into tb2 (c1,c2) values (?,?),(?,@c2)")
}

func TestInsert_KeyValue(t *testing.T) {
	a := assert.New(t)
	i := Insert(nil, "table")
	i.KeyValue("c1", 1).KeyValue("c2", sql.Named("c2", 2))
	query, args, err := i.SQL()
	a.NotError(err)
	a.Equal(args, []interface{}{1, sql.Named("c2", 2)})
	sqltest.Equal(a, query, "insert into table (c1,c2) values(?,@c2)")

	i.Reset()
	i.Table("table")
	i.Columns("c1", "c2")
	i.Values(1, 2)
	i.KeyValue("c3", 3)
	query, args, err = i.SQL()
	a.NotError(err)
	a.Equal(args, []interface{}{1, 2, 3})
	sqltest.Equal(a, query, "insert into table (c1,c2,c3) values(?,?,?)")

	// 添加第二行数据，就不能再次使用 KeyValue 了
	i.Values(1, 2, 3)
	a.Panic(func() {
		i.KeyValue("c4", 4)
	})
}

func TestInsertError(t *testing.T) {
	a := assert.New(t)
	i := Insert(nil, "table")
	a.NotNil(i)

	query, args, err := i.Columns("c1", "c2").SQL()
	a.Error(err).Nil(args).Empty(query)

	i.Reset()
	i.Table("tb1")
	query, args, err = i.Columns("c1", "c2").Values(1).SQL()
	a.Error(err).Nil(args).Empty(query)
}
