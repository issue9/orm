// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"database/sql"

	"github.com/issue9/orm/model"
	"github.com/issue9/orm/sqlbuilder"
)

// Engine 是 DB 与 Tx 的共有接口。
type Engine interface {
	sqlbuilder.Engine

	// 获取与之关联的 Dialect 接口。
	Dialect() Dialect

	Insert(v interface{}) (sql.Result, error)

	Delete(v interface{}) (sql.Result, error)

	Update(v interface{}) (sql.Result, error)

	MustUpdate(v interface{}, cols ...string) (sql.Result, error)

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
}

// Dialect 数据库驱动特有的语言特性实现
type Dialect interface {
	sqlbuilder.Dialect

	// 生成创建表的 SQL 语句。
	//
	// 创建表可能生成多条语句，比如创建表，以及相关的创建索引语句。
	CreateTableSQL(m *model.Model) ([]string, error)
}

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
	return sqlbuilder.Select(sql.engine, sql.engine.Dialect())
}

// CreateIndex 生成创建索引的语句
func (sql *SQL) CreateIndex() *sqlbuilder.CreateIndexStmt {
	return sqlbuilder.CreateIndex(sql.engine)
}

// DropTable 生成删除表的语句
func (sql *SQL) DropTable() *sqlbuilder.DropTableStmt {
	return sqlbuilder.DropTable(sql.engine)
}

// Truncate 生成清空表的语句
func (sql *SQL) Truncate() *sqlbuilder.TruncateStmt {
	return sqlbuilder.Truncate(sql.engine, sql.engine.Dialect())
}
