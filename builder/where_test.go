// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package builder

import (
	"testing"

	"github.com/issue9/assert"
)

func TestSQL_Where(t *testing.T) {
	a := assert.New(t)
	sql := NewSQL(nil)

	sql.Columns("c1").
		Table("user").
		Where("id=@id").
		And("group=1").
		Or("name like @name")

	wont := "select c1 from user where(id=@id) and(group=1) or(name like @name)"
	chkSQLEqual(a, wont, sql.selectSQL())
	a.False(sql.HasErrors())

	// In
	sql.Reset().
		Columns("c1").
		Table("user").
		In("c1", 1, 2, "3").
		OrIn("c2", 4, 5, "6")

	wont = "select c1 from user where(c1 in(1,2,'3')) or (c2 in(4,5,'6'))"
	chkSQLEqual(a, wont, sql.selectSQL())
	a.False(sql.HasErrors())

	// between
	sql.Reset().
		Columns("c1").
		Table("user").
		Between("c1", 5, "@end").
		OrBetween("c2", 7, 8)

	wont = "SELECT c1 from user where (c1 between 5 and @end) or (c2 between 7 and 8)"
	chkSQLEqual(a, wont, sql.selectSQL())
	a.False(sql.HasErrors())

	// is null
	sql.Reset().
		Columns("*").
		Table("user").
		IsNotNull("c1").
		OrIsNotNull("c2").
		IsNull("c3").
		OrIsNull("c4")

	wont = "SELECT * FROM user where(c1 is not null) or (c2 is not null) and (c3 is null) OR (c4 is null)"
	chkSQLEqual(a, wont, sql.selectSQL())
	a.False(sql.HasErrors())

	// in参数错误
	sql.Reset().In("c1")
	a.True(sql.HasErrors())

	// where build 参数错误
	sql.Reset()
	a.False(sql.HasErrors())
	sql.whereBuild(3, "id=@id") // 第一次忽略op参数
	a.False(sql.HasErrors())
	sql.whereBuild(3, "id=@id") // 第二次错误的op参数，会报错
	a.True(sql.HasErrors())

}
