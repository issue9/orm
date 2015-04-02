// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package builder

import (
	"testing"

	"github.com/issue9/assert"
)

func TestSQL_Join(t *testing.T) {
	a := assert.New(t)
	sql := NewSQL(nil)

	sql.Columns("c1", "c2").
		Table("#user as u").
		Where("true").
		LeftJoin("#group as g", "g.id=u.gid")

	wont := "SELECT c1,c2 FROM #user as u LEFT JOIN #group as g on g.id=u.gid WHERE(true)"
	chkSQLEqual(a, wont, sql.selectSQL())
	a.False(sql.HasErrors())

	sql.RightJoin("#group as g", "g.id=u.gid")
	wont = "SELECT c1,c2 FROM #user as u LEFT JOIN #group as g on g.id=u.gid RIGHT join #group as g on g.id=u.gid WHERE(true)"
	chkSQLEqual(a, wont, sql.selectSQL())
	a.False(sql.HasErrors())
}

func TestSQL_Order(t *testing.T) {
	a := assert.New(t)
	sql := NewSQL(nil)

	sql.Columns("c1", "c2").
		Table("#user").
		Where("true").
		Asc("c1").
		Desc("c2,c3")

	wont := "SELECT c1,c2 FROM #user WHERE(true)ORDER BY c1 ASC , c2, c3 desc"
	chkSQLEqual(a, wont, sql.selectSQL())
	a.False(sql.HasErrors())
}

func TestSQL_Limit(t *testing.T) {
	a := assert.New(t)
	db := newDB(a)
	defer db.Close(a)

	sql := NewSQL(db)

	sql.Columns("c1", "c2").
		Table("#user").
		Where("true").
		Limit(5)

	wont := "SELECT c1,c2 FROM #user WHERE(true) limit 5"
	chkSQLEqual(a, wont, sql.selectSQL())
	a.False(sql.HasErrors())

	sql.Limit(5, 7)
	wont = "SELECT c1,c2 FROM #user WHERE(true) limit 5 offset 7"
	chkSQLEqual(a, wont, sql.selectSQL())
	a.False(sql.HasErrors())

	// 多余的参数
	sql.Limit(5, 6, 7)
	a.True(sql.HasErrors())
	rs, err := sql.Query(nil) // 不处理错误，直接招待Query
	a.Error(err).Nil(rs)
}

// 测试SQL.Fetch*系列函数
func TestSQL_Fetch(t *testing.T) {
	a := assert.New(t)
	db := newDB(a)
	defer db.Close(a)

	// Fetch2Maps
	sql := NewSQL(db)
	maps, err := sql.Columns("*").
		Table("user").
		Limit(2).
		Asc("id").
		Fetch2Maps(nil)
	a.NotError(err).NotNil(maps)
	a.Equal(maps, []map[string]interface{}{
		map[string]interface{}{"id": 1, "account": []byte("account-1")}, // 字符串默认被转换成[]byte
		map[string]interface{}{"id": 2, "account": []byte("account-2")},
	})

	// Fetch2Map
	m, err := sql.Fetch2Map(nil)
	a.NotError(err).NotNil(m)
	a.Equal(m, map[string]interface{}{"id": 1, "account": []byte("account-1")})

	// FetchColumn
	col, err := sql.FetchColumn("id", nil)
	a.NotError(err).NotNil(col)
	a.Equal(col, 1)

	// FetchColumns
	cols, err := sql.FetchColumns("id", nil)
	a.NotError(err).NotNil(cols)
	a.Equal(cols, []interface{}{1, 2})

	// FetchObj=>单个对象
	obj := &user{}
	a.NotError(sql.FetchObj(obj, nil))
	a.Equal(obj, &user{ID: 1, Account: "account-1"})

	// FetchObj=>对象数组
	objs := []*user{&user{}}
	a.NotError(sql.FetchObj(&objs, nil))
	a.Equal(objs, []*user{&user{ID: 1, Account: "account-1"}, &user{ID: 2, Account: "account-2"}})

	// 以desc排序

	// Fetch2Maps
	maps, err = sql.Reset().
		Columns("*").
		Table("user").
		Limit("@limit").
		Desc("id").
		Fetch2Maps(map[string]interface{}{"limit": 2})
	a.NotError(err).NotNil(maps)
	a.Equal(maps, []map[string]interface{}{
		map[string]interface{}{"id": 10, "account": []byte("account-10")}, // 字符串默认被转换成[]byte
		map[string]interface{}{"id": 9, "account": []byte("account-9")},
	})

	// Fetch2Map
	m, err = sql.Fetch2Map(map[string]interface{}{"limit": 2})
	a.NotError(err).NotNil(m)
	a.Equal(m, map[string]interface{}{"id": 10, "account": []byte("account-10")})

	// FetchColumn
	col, err = sql.FetchColumn("id", map[string]interface{}{"limit": 2})
	a.NotError(err).NotNil(col)
	a.Equal(col, 10)

	// FetchColumns
	cols, err = sql.FetchColumns("id", map[string]interface{}{"limit": 2})
	a.NotError(err).NotNil(cols)
	a.Equal(cols, []interface{}{10, 9})

	// FetchObj=>单个对象
	obj = &user{}
	a.NotError(sql.FetchObj(obj, map[string]interface{}{"limit": 2}))
	a.Equal(obj, &user{ID: 10, Account: "account-10"})

	// FetchObj=>对象数组
	objs = []*user{&user{}}
	a.NotError(sql.FetchObj(&objs, map[string]interface{}{"limit": 2}))
	a.Equal(objs, []*user{&user{ID: 10, Account: "account-10"}, &user{ID: 9, Account: "account-9"}})
}
