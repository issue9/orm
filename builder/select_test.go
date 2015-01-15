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

	sql.RightJoin("#group as g", "g.id=u.gid")
	wont = "SELECT c1,c2 FROM #user as u LEFT JOIN #group as g on g.id=u.gid RIGHT join #group as g on g.id=u.gid WHERE(true)"
	chkSQLEqual(a, wont, sql.selectSQL())
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

	sql.Limit(5, 7)
	wont = "SELECT c1,c2 FROM #user WHERE(true) limit 5 offset 7"
	chkSQLEqual(a, wont, sql.selectSQL())
}

func TestSQL_Fetch(t *testing.T) {
	a := assert.New(t)
	db := newDB(a)
	defer db.Close(a)

	sql := NewSQL(db)
	m, err := sql.Columns("*").
		Table("user").
		Limit(2).
		Asc("id").
		Fetch2Maps(nil)
	a.NotError(err).NotNil(m)
	a.Equal(m, []map[string]interface{}{
		map[string]interface{}{"id": 1, "account": []byte("account-1")}, // 字符串默认被转换成[]byte
		map[string]interface{}{"id": 2, "account": []byte("account-2")},
	})

	// desc
	m, err = sql.Reset().
		Columns("*").
		Table("user").
		Limit("@limit").
		Desc("id").
		Fetch2Maps(map[string]interface{}{"limit": 2})
	a.NotError(err).NotNil(m)
	a.Equal(m, []map[string]interface{}{
		map[string]interface{}{"id": 10, "account": []byte("account-10")}, // 字符串默认被转换成[]byte
		map[string]interface{}{"id": 9, "account": []byte("account-9")},
	})
}
