// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"database/sql"

	"github.com/issue9/orm/v3/core"
	"github.com/issue9/orm/v3/fetch"
	"github.com/issue9/orm/v3/sqlbuilder"
)

// AfterFetcher 从数据库查询到数据之后，需要执行的操作。
type AfterFetcher = fetch.AfterFetcher

// Column 列结构
type Column = core.Column

// Dialect 数据库驱动特有的语言特性实现
type Dialect = core.Dialect

// BeforeUpdater 在更新之前调用的函数
type BeforeUpdater interface {
	BeforeUpdate() error
}

// BeforeInserter 在插入之前调用的函数
type BeforeInserter interface {
	BeforeInsert() error
}

// Engine 是 DB 与 Tx 的共有接口。
type Engine interface {
	core.Engine

	// 理论上功能等同于以下两步操作：
	//  rslt, err := engine.Insert(obj)
	//  id, err := rslt.LastInsertId()
	// 但是实际上部分数据库不支持直接在 sql.Result 中获取 LastInsertId，
	// 比如 postgresql，所以使用 LastInsertID() 会是比 sql.Result
	// 更简单和安全的方法。
	//
	// NOTE: 要求 v 有定义自增列。
	LastInsertID(v interface{}) (int64, error)

	Insert(v interface{}) (sql.Result, error)

	Delete(v interface{}) (sql.Result, error)

	Update(v interface{}, cols ...string) (sql.Result, error)

	Select(v interface{}) error

	Create(v interface{}) error

	Drop(v interface{}) error

	Truncate(v interface{}) error

	InsertMany(v interface{}, max int) error

	MultInsert(objs ...interface{}) error

	MultSelect(objs ...interface{}) error

	MultUpdate(objs ...interface{}) error

	MultDelete(objs ...interface{}) error

	MultCreate(objs ...interface{}) error

	MultDrop(objs ...interface{}) error

	MultTruncate(objs ...interface{}) error

	SQLBuilder() *sqlbuilder.SQLBuilder

	NewModel(v interface{}) (*Model, error)
}
