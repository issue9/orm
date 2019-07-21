// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"database/sql"

	"github.com/issue9/orm/v2/fetch"
	"github.com/issue9/orm/v2/sqlbuilder"
)

// BeforeUpdater 在更新之前调用的函数
type BeforeUpdater interface {
	BeforeUpdate() error
}

// BeforeInserter 在插入之前调用的函数
type BeforeInserter interface {
	BeforeInsert() error
}

// AfterFetcher 从数据库查询到数据之后，需要执行的操作。
type AfterFetcher = fetch.AfterFetcher

// Engine 是 DB 与 Tx 的共有接口。
type Engine interface {
	sqlbuilder.Engine

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

	Count(v interface{}) (int64, error)

	Create(v interface{}) error

	Drop(v interface{}) error

	Truncate(v interface{}) error

	MultInsert(objs ...interface{}) error

	MultSelect(objs ...interface{}) error

	MultUpdate(objs ...interface{}) error

	MultDelete(objs ...interface{}) error

	MultCreate(objs ...interface{}) error

	MultDrop(objs ...interface{}) error

	MultTruncate(objs ...interface{}) error

	SQL() *SQL

	NewModel(v interface{}) (*Model, error)
}

// Dialect 数据库驱动特有的语言特性实现
type Dialect = sqlbuilder.Dialect

// SQL 用于生成 SQL 语句
type SQL struct {
	engine Engine
}

// Delete 生成删除语句
func (sql *SQL) Delete() *sqlbuilder.DeleteStmt {
	return sqlbuilder.Delete(sql.engine)
}

// Update 生成更新语句
func (sql *SQL) Update() *sqlbuilder.UpdateStmt {
	return sqlbuilder.Update(sql.engine)
}

// Insert 生成插入语句
func (sql *SQL) Insert() *sqlbuilder.InsertStmt {
	return sqlbuilder.Insert(sql.engine)
}

// Select 生成插入语句
func (sql *SQL) Select() *sqlbuilder.SelectStmt {
	return sqlbuilder.Select(sql.engine)
}

// CreateIndex 生成创建索引的语句
func (sql *SQL) CreateIndex() *sqlbuilder.CreateIndexStmt {
	return sqlbuilder.CreateIndex(sql.engine)
}

// DropTable 生成删除表的语句
func (sql *SQL) DropTable() *sqlbuilder.DropTableStmt {
	return sqlbuilder.DropTable(sql.engine)
}

// TruncateTable 生成清空表的语句，同时重置 AI 计算
func (sql *SQL) TruncateTable() *sqlbuilder.TruncateTableStmt {
	return sqlbuilder.TruncateTable(sql.engine)
}
