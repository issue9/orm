// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/dialect"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

// 测试orm的一些常用操作：New,Get,Close,CloseAll
func TestEngines(t *testing.T) {
	a := assert.New(t)

	// 注册dialect
	a.NotError(dialect.Register("sqlite3", &dialect.Sqlite3{}))
	a.NotError(dialect.Register("mysql", &dialect.Mysql{}))

	e, err := New("sqlite3", "./test", "main", "main_")
	a.NotError(err).NotNil(e)

	// 不存在的实例
	e, found := Get("test1test")
	a.False(found).Nil(e)

	// 获取注册的名为test的Engine实例
	e, found = Get("main")
	a.True(found).NotNil(e)

	// 关闭之后，是否能再次正常获取
	Close("main")
	e, found = Get("main")
	a.False(found).Nil(e)

	// 重新添加2个Engine

	e, err = New("mysql", "root:@/", "second", "second_")
	a.NotError(err).NotNil(e)

	e, err = New("sqlite3", "./test", "main", "main_")
	a.NotError(err).NotNil(e)

	e, found = Get("main")
	a.True(found).NotNil(e)

	e, found = Get("second")
	a.True(found).NotNil(e)

	// 关闭所有
	CloseAll()
	e, found = Get("main")
	a.False(found).Nil(e)
	e, found = Get("second")
	a.False(found).Nil(e)
}
