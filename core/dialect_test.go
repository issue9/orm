// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package core

import (
	"testing"

	"github.com/issue9/assert"
	//"github.com/issue9/orm/dialect"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

func TestIsRegistedDriver(t *testing.T) {
	a := assert.New(t)

	a.True(isRegistedDriver("sqlite3"))
	a.True(isRegistedDriver("mysql"))

	a.False(isRegistedDriver("abcdeg"))
}

func TestDialects(t *testing.T) {
	a := assert.New(t)

	clearDialect()
	a.Empty(dialects.items)

	err := Register("sqlite3", &sqlite3{})
	a.NotError(err).
		True(IsRegisted("sqlite3"))

	// Get
	d, found := Get("sqlite3") // 已注册
	a.True(found).NotNil(d)
	d, found = Get("sqlite4") // 未注册
	a.False(found).Nil(d)

	// 注册一个相同名称的
	err = Register("sqlite3", &sqlite3{})
	a.Error(err)                    // 注册失败
	a.Equal(1, len(dialects.items)) // 数量还是1，注册没有成功

	// 再注册一个名称不相同的
	err = Register("mysql", &mysql{})
	a.NotError(err)
	a.Equal(2, len(dialects.items))

	// 注册类型相同，但名称不同的实例
	err = Register("fake3", &mysql{})
	a.Error(err)                    // 注册失败
	a.Equal(2, len(dialects.items)) // 数量还是2，注册没有成功

	// 清空
	clearDialect()
	a.Empty(dialects.items)
	d, found = Get("sqlite3")
	a.False(found).Nil(d)
}
