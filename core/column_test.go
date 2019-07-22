// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package core

import (
	"reflect"
	"testing"

	"github.com/issue9/assert"
)

func TestColumn_IsZero(t *testing.T) {
	a := assert.New(t)
	col := &Column{}

	col.GoType = IntType
	col.GoZero = reflect.Zero(col.GoType).Interface()
	a.True(col.IsZero(reflect.ValueOf(int(0))))
	a.False(col.IsZero(reflect.ValueOf(1)))

	col.GoType = reflect.TypeOf([]byte{})
	col.GoZero = reflect.Zero(col.GoType).Interface()
	a.True(col.IsZero(reflect.ValueOf([]byte{})))
	a.True(col.IsZero(reflect.ValueOf([]byte(""))))
	a.False(col.IsZero(reflect.ValueOf([]byte{'0'})))

	col.GoType = RawBytesType
	col.GoZero = reflect.Zero(col.GoType).Interface()
	a.True(col.IsZero(reflect.ValueOf([]byte{})))
	a.True(col.IsZero(reflect.ValueOf([]byte(""))))
	a.False(col.IsZero(reflect.ValueOf([]byte{'0'})))
}
