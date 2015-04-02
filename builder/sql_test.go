// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package builder

import (
	"errors"
	"strings"
	"testing"

	"github.com/issue9/assert"
)

func TestErrors(t *testing.T) {
	a := assert.New(t)

	errs := Errors([]error{errors.New("1"), errors.New("2")})
	a.True(strings.Index(errs.Error(), "1") > -1)
	a.True(strings.Index(errs.Error(), "2") > -1)
}

func TestSQL_Delete(t *testing.T) {
	a := assert.New(t)
	db := newDB(a)
	defer db.Close(a)

	sql := NewSQL(db)
	sql.Where("id>@id").
		Table("user").
		Columns("abc"). // 这行应该过滤掉
		Or("account=@account")

	// 删除两条记录id>6,id=2
	wont := "DELETE FROM user where (id>@id) or (account=@account)"
	chkSQLEqual(a, wont, sql.deleteSQL())
	_, err := sql.Delete(map[string]interface{}{"id": 6, "account": "account-2"})
	a.NotError(err)

	// 比较剩余记录
	m, err := sql.Reset().
		Table("user").
		Columns("*").
		Asc("id").
		Fetch2Maps(nil)
	a.NotError(err).NotNil(m)
	a.Equal(m, []map[string]interface{}{
		map[string]interface{}{"id": 1, "account": []byte("account-1")},
		map[string]interface{}{"id": 3, "account": []byte("account-3")},
		map[string]interface{}{"id": 4, "account": []byte("account-4")},
		map[string]interface{}{"id": 5, "account": []byte("account-5")},
		map[string]interface{}{"id": 6, "account": []byte("account-6")},
	})

	// 删除id=4,5,6的记录
	sql.Reset().
		In("id", 4, 5, 6).
		Table("user")
	wont = "DELETE FROM user where (id in (4,5,6))"
	chkSQLEqual(a, wont, sql.deleteSQL())
	_, err = sql.Delete(nil)
	a.NotError(err)

	// 比较剩余记录
	m, err = sql.Reset().
		Table("user").
		Columns("*").
		Asc("id").
		Fetch2Maps(nil)
	a.NotError(err).NotNil(m)
	a.Equal(m, []map[string]interface{}{
		map[string]interface{}{"id": 1, "account": []byte("account-1")},
		map[string]interface{}{"id": 3, "account": []byte("account-3")},
	})
}

func TestSQL_Update(t *testing.T) {
	a := assert.New(t)
	db := newDB(a)
	defer db.Close(a)

	// 更新一条记录,id=1
	sql := NewSQL(db).
		Table("user").
		Where("id=@id").
		Columns("abc"). // 被过滤的内容
		Data(map[string]interface{}{"account": "account-upd"})

	wont := "UPDATE user set account='account-upd' where(id=@id)"
	chkSQLEqual(a, wont, sql.updateSQL())
	_, err := sql.Update(map[string]interface{}{"id": 1})
	a.NotError(err)

	m, err := sql.Reset().
		Where("id=@id").
		Table("user").
		Columns("*").
		Fetch2Map(map[string]interface{}{"id": 1})
	a.NotError(err).NotNil(m)

	a.Equal(m, map[string]interface{}{"id": 1, "account": []byte("account-upd")})

	// 更新多条记录
	sql.Reset().
		Table("user").
		Where("id<3").
		Set("account", "abc")

	wont = "UPDATE user set account='abc' where(id<3)"
	chkSQLEqual(a, wont, sql.updateSQL())
	_, err = sql.Update(nil)
	a.NotError(err)

	mapped, err := sql.Reset().
		Where("id<4").
		Table("user").
		Columns("*").
		Fetch2Maps(nil)
	a.NotError(err).NotNil(m)

	a.Equal(mapped, []map[string]interface{}{
		map[string]interface{}{"id": 1, "account": []byte("abc")},
		map[string]interface{}{"id": 2, "account": []byte("abc")},
		map[string]interface{}{"id": 3, "account": []byte("account-3")},
	})
}

func TestSQL_Insert(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer db.Close(a)

	sql := NewSQL(db).
		Table("user").
		Data(map[string]interface{}{"account": "@account"})
	stmt, err := sql.Prepare(Insert, "insert-user")
	a.NotError(err).NotNil(stmt)
	defer func() {
		a.NotError(stmt.Close())
	}()

	r, err := stmt.Exec(map[string]interface{}{"account": "insert"})
	a.NotError(err).NotNil(r)
	id, err := r.LastInsertId()
	a.NotError(err).Equal(11, id)
}

func TestSQL_Exec(t *testing.T) {
	a := assert.New(t)

	db := newDB(a)
	defer db.Close(a)

	sql := NewSQL(db).
		Table("user").
		Where("id=@id").
		Data(map[string]interface{}{"account": "@account"})

	// 错误的Action类型
	r, err := sql.Exec(100, map[string]interface{}{"id": 1, "account": "upd"})
	a.Error(err).Nil(r)

	// 错误的Action类型:Select
	r, err = sql.Exec(Select, map[string]interface{}{"id": 1, "account": "upd"})
	a.Error(err).Nil(r)

	// 执行Exec:Update
	r, err = sql.Exec(Update, map[string]interface{}{"id": 1, "account": "upd"})
	a.NotError(err).NotNil(r)

	// 验证Exec执行的结果
	m, err := sql.Reset().
		Table("user").
		Where("id=1").
		Columns("*").
		Fetch2Map(nil)
	a.NotError(err).NotNil(m)
	a.Equal(m, map[string]interface{}{"id": 1, "account": []byte("upd")})
}
