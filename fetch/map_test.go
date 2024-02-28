// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package fetch_test

import (
	"reflect"
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/orm/v5/fetch"
	"github.com/issue9/orm/v5/internal/test"
)

func TestMap(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a)

	eq := func(m1, m2 []map[string]any) bool {
		if len(m1) != len(m2) {
			return false
		}

		for i, s1 := range m1 {
			s2 := m2[i]
			if len(s2) != len(s1) {
				return false
			}

			for k, v := range s1 {
				if !reflect.DeepEqual(s2[k], v) {
					return false
				}
			}
		}
		return true
	}

	suite.Run(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		db := t.DB

		// 正常匹配数据，读取多行
		sql := `SELECT id,email FROM fetch_users WHERE id<3 ORDER BY id`
		rows, err := db.Query(sql)
		t.NotError(err).NotNil(rows)

		mapped, err := fetch.Map(false, rows)
		t.NotError(err).NotNil(mapped)

		ok := eq([]map[string]any{
			{"id": int64(1), "email": "email-1"},
			{"id": int64(2), "email": "email-2"},
		}, mapped) ||
			eq([]map[string]any{
				{"id": []byte{'1'}, "email": []byte("email-1")},
				{"id": []byte{'2'}, "email": []byte("email-2")},
			}, mapped)
		t.True(ok)
		t.NotError(rows.Close())

		// 正常匹配数据，读取一行
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		mapped, err = fetch.Map(true, rows)
		t.NotError(err).NotNil(mapped)

		ok = eq([]map[string]any{
			{"id": int64(1), "email": "email-1"},
		}, mapped) ||
			eq([]map[string]any{
				{"id": []byte{'1'}, "email": []byte("email-1")},
			}, mapped)
		t.True(ok)
		t.NotError(rows.Close())

		// 没有匹配的数据，读取多行
		sql = `SELECT id,email FROM fetch_users WHERE id<0 ORDER BY id`
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		mapped, err = fetch.Map(false, rows)
		t.NotError(err)

		t.Equal([]map[string]any{}, mapped)
		t.NotError(rows.Close())

		// 没有匹配的数据，读取一行
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		mapped, err = fetch.Map(true, rows)
		t.NotError(err)

		t.Equal([]map[string]any{}, mapped)
		t.NotError(rows.Close())
	})
}

func TestMapString(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a)

	suite.Run(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		db := t.DB

		// 正常数据匹配，读取多行
		sql := `SELECT id,email FROM fetch_users WHERE id<3 ORDER BY id`
		rows, err := db.Query(sql)
		t.NotError(err).NotNil(rows)

		mapped, err := fetch.MapString(false, rows)
		t.NotError(err).NotNil(mapped)

		t.Equal(mapped, []map[string]string{
			{"id": "1", "email": "email-1"},
			{"id": "2", "email": "email-2"},
		})
		t.NotError(rows.Close())

		// 正常数据匹配，读取一行
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		mapped, err = fetch.MapString(true, rows)
		t.NotError(err).NotNil(mapped)

		t.Equal(mapped, []map[string]string{
			{"id": "1", "email": "email-1"},
		})
		t.NotError(rows.Close())

		// 没有数据匹配，读取多行
		sql = `SELECT id,email FROM fetch_users WHERE id<0 ORDER BY id`
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		mapped, err = fetch.MapString(false, rows)
		t.NotError(err)

		t.Equal(mapped, []map[string]string{})
		t.NotError(rows.Close())

		// 没有数据匹配，读取一行
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		mapped, err = fetch.MapString(true, rows)
		t.NotError(err)

		t.Equal(mapped, []map[string]string{})
		t.NotError(rows.Close())
	})
}
