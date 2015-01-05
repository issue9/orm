// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// 声明了一些测试用的虚假类：
// - fakeDB实现了DB接口的类，内部调用sqlite3实现。
// - fake1 fakeDriver1注册的数据库实例，与fakeDialect1组成一对。
// - fake2 fakeDriver2注册的数据库实例，与fakeDialect2组成一对。

package orm

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/dialect"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

func TestIsRegistedDriver(t *testing.T) {
	a := assert.New(t)

	a.True(isRegistedDriver("sqlite3"))
	a.False(isRegistedDriver("abcdeg"))
}

func TestDialects(t *testing.T) {
	a := assert.New(t)

	clearDialect()
	a.Empty(dialects.items)

	err := Register("sqlite3", &dialect.Sqlite3{})
	a.NotError(err).
		True(IsRegisted("sqlite3"))

	// 注册一个相同名称的
	err = Register("sqlite3", &dialect.Sqlite3{})
	a.Error(err)                    // 注册失败
	a.Equal(1, len(dialects.items)) // 数量还是1，注册没有成功

	// 再注册一个名称不相同的
	err = Register("mysql", &dialect.Mysql{})
	a.NotError(err)
	a.Equal(2, len(dialects.items))

	// 注册类型相同，但名称不同的实例
	err = Register("fake3", &dialect.Mysql{})
	a.Error(err)                    // 注册失败
	a.Equal(2, len(dialects.items)) // 数量还是2，注册没有成功

	// 清空
	clearDialect()
	a.Empty(dialects.items)

	// 恢复默认的dialects注册
	regDialects()
}
