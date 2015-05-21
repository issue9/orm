// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/issue9/assert"
)

var _ engine = &DB{}
var _ engine = &Tx{}

func TestWhere(t *testing.T) {
	a := assert.New(t)
	db := newDB(a)
	defer func() {
		a.NotError(db.Close())
	}()

	fn := func(obj interface{}, sql string, vals []interface{}) {
		w := new(bytes.Buffer)
		m, err := newModel(obj)
		a.NotError(err).NotNil(m)

		rval := reflect.ValueOf(obj)
		for rval.Kind() == reflect.Ptr {
			rval = rval.Elem()
		}

		ret, err := where(db, w, m, rval)
		a.NotError(err).
			Equal(w.String(), sql).
			Equal(ret, vals)
	}

	// 测试有PK的情况
	fn(&user{ID: 1}, " WHERE `id`=?", []interface{}{1})

	// 测试只指定唯一索引的情况
	fn(&userInfo{
		LastName:  "l",
		FirstName: "f",
	}, " WHERE `firstName`=? AND `lastName`=?", []interface{}{"f", "l"})

	// 同时存在pk和唯一约束，以PK优先
	fn(&admin{
		user:  user{ID: 1},
		Email: "email@test.com",
	}, " WHERE `id`=?", []interface{}{1})
}
