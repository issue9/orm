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

var (
	_ SQLer       = &UpdateStmt{}
	_ WhereStmter = &UpdateStmt{}
)

func TestUpdate_columnsHasDup(t *testing.T) {
	a := assert.New(t)
	u := Update(nil).Table("table")

	u.values = []*updateSet{
		&updateSet{
			column: "c1",
			value:  1,
		},
		&updateSet{
			column: "c2",
			value:  1,
		},

		&updateSet{
			column: "c2",
			value:  1,
		},
	}
	a.True(u.columnsHasDup())

	u.values = []*updateSet{
		&updateSet{
			column: "c1",
			value:  1,
		},
		&updateSet{
			column: "c1",
			value:  1,
		},

		&updateSet{
			column: "c2",
			value:  1,
		},
	}
	a.True(u.columnsHasDup())

	u.values = []*updateSet{
		&updateSet{
			column: "c1",
			value:  1,
		},
		&updateSet{
			column: "c2",
			value:  1,
		},

		&updateSet{
			column: "c1",
			value:  1,
		},
	}
	a.True(u.columnsHasDup())

	u.values = []*updateSet{
		&updateSet{
			column: "c1",
			value:  1,
		},
		&updateSet{
			column: "c2",
			value:  1,
		},

		&updateSet{
			column: "c3",
			value:  1,
		},
	}
	a.False(u.columnsHasDup())
}

func TestUpdate(t *testing.T) {
	a := assert.New(t)
	u := Update(nil).Table("table")
	a.NotNil(u)

	// 不带 where 部分
	u.Set("c1", 1).Set("c2", sql.Named("c2", 2))
	query, args, err := u.SQL()
	a.NotError(err)
	a.Equal(args, []interface{}{1, sql.Named("c2", 2)})
	sqltest.Equal(a, query, "update table SET c1=?,c2=@c2")

	// bug(caix): UpdateStmt 采用 map 保存修改的值，
	// 而 map 的顺序是不一定的，所以测试的比较内容，可能会出现值顺序不一样，
	// 从页导致测试失败
	u.Increase("c3", 3).
		Increase("c4", sql.Named("c4", 4)).
		Decrease("c5", 5).
		Decrease("c6", sql.Named("c6", 6))
	query, args, err = u.SQL()
	a.NotError(err)
	a.Equal(args, []interface{}{1, sql.Named("c2", 2), 3, sql.Named("c4", 4), 5, sql.Named("c6", 6)})
	sqltest.Equal(a, query, "update table SET c1=?,c2=@c2,c3=c3+?,c4=c4+@c4,c5=c5-?,c6=c6-@c6")

	// 重置
	u.Reset()
	a.Empty(u.table).Empty(u.values)

	u.Table("tb1").Table("tb2")
	u.Increase("c1", 1).
		Where("id=?", 1).
		Or("id=?", 2)
	query, args, err = u.SQL()
	a.NotError(err)
	a.Equal(args, []interface{}{1, 1, 2})
	sqltest.Equal(a, query, "update tb2 SET c1=c1+? where id=? or id=?")
}
