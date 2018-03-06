// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"context"
	"database/sql"
	"strings"

	"github.com/issue9/orm/core"
)

// DB 数据库操作实例。
type DB struct {
	stdDB       *sql.DB
	dialect     core.Dialect
	tablePrefix string
	replacer    *strings.Replacer
}

// NewDB 声明一个新的 DB 实例。
func NewDB(driverName, dataSourceName, tablePrefix string, dialect core.Dialect) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	return NewDBWithStdDB(db, tablePrefix, dialect)
}

// NewDBWithStdDB 从 sql.DB 构建一个 DB 实例。
func NewDBWithStdDB(db *sql.DB, tablePrefix string, dialect core.Dialect) (*DB, error) {
	l, r := dialect.QuoteTuple()
	return &DB{
		stdDB:       db,
		dialect:     dialect,
		tablePrefix: tablePrefix,
		replacer: strings.NewReplacer(
			"#", tablePrefix,
			"{", string(l),
			"}", string(r),
		),
	}, nil
}

// Close 关闭当前数据库，释放所有的链接。
// 关闭之后，之前通过 DB.StdDB() 返回的实例也将失效。
// 通过调用 DB.StdDB().Close() 也将使当前实例失效。
func (db *DB) Close() error {
	return db.stdDB.Close()
}

// StdDB 返回标准包中的 sql.DB 指针。
func (db *DB) StdDB() *sql.DB {
	return db.stdDB
}

// Dialect 返回对应的 core.Dialect 接口实例。
func (db *DB) Dialect() core.Dialect {
	return db.dialect
}

// Query 执行一条查询语句，并返回相应的 sql.Rows 实例。
// 具体参数说明可参考 core.Engine 接口文档。
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	query = db.replacer.Replace(query)
	query, err := db.dialect.SQL(query)
	if err != nil {
		return nil, err
	}

	return db.stdDB.Query(query, args...)
}

// QueryContext 执行一条查询语句，并返回相应的 sql.Rows 实例。
// 具体参数说明可参考 core.Engine 接口文档。
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	query = db.replacer.Replace(query)
	query, err := db.dialect.SQL(query)
	if err != nil {
		return nil, err
	}

	return db.stdDB.QueryContext(ctx, query, args...)
}

// Exec 执行 SQL 语句。
// 具体参数说明可参考 core.Engine 接口文档。
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	query = db.replacer.Replace(query)
	query, err := db.dialect.SQL(query)
	if err != nil {
		return nil, err
	}

	return db.stdDB.Exec(query, args...)
}

// ExecContext 执行 SQL 语句。
// 具体参数说明可参考 core.Engine 接口文档。
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	query = db.replacer.Replace(query)
	query, err := db.dialect.SQL(query)
	if err != nil {
		return nil, err
	}

	return db.stdDB.ExecContext(ctx, query, args...)
}

// Prepare 预编译查询语句。
// 具体参数说明可参考 core.Engine 接口文档。
func (db *DB) Prepare(query string) (*sql.Stmt, error) {
	query = db.replacer.Replace(query)
	query, err := db.dialect.SQL(query)
	if err != nil {
		return nil, err
	}

	return db.stdDB.Prepare(query)
}

// PrepareContext 预编译查询语句。
// 具体参数说明可参考 core.Engine 接口文档。
func (db *DB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	query = db.replacer.Replace(query)
	query, err := db.dialect.SQL(query)
	if err != nil {
		return nil, err
	}

	return db.stdDB.PrepareContext(ctx, query)
}

// Insert 插入数据，若需一次性插入多条数据，请使用tx.Insert()。
func (db *DB) Insert(v interface{}) (sql.Result, error) {
	return insert(db, v)
}

// Delete 删除符合条件的数据。
// 查找条件以结构体定义的主键或是唯一约束(在没有主键的情况下)来查找，
// 若两者都不存在，则将返回error
func (db *DB) Delete(v interface{}) (sql.Result, error) {
	return del(db, v)
}

// Update 更新数据，零值不会被提交，cols 指定的列，即使是零值也会被更新。
//
// 查找条件以结构体定义的主键或是唯一约束(在没有主键的情况下)来查找，
// 若两者都不存在，则将返回 error
func (db *DB) Update(v interface{}, cols ...string) (sql.Result, error) {
	return update(db, v, cols...)
}

// Select 查询一个符合条件的数据。
// 查找条件以结构体定义的主键或是唯一约束(在没有主键的情况下)来查找，
// 若两者都不存在，则将返回error
// 若没有符合条件的数据，将不会对参数v做任何变动。
func (db *DB) Select(v interface{}) error {
	return find(db, v)
}

// Count 查询符合 v 条件的记录数量。
// v 中的所有非零字段都将参与查询。
// 若需要复杂的查询方式，请构建 core.SQL 对象查询。
func (db *DB) Count(v interface{}) (int, error) {
	return count(db, v)
}

// Create 创建一张表。
func (db *DB) Create(v interface{}) error {
	return create(db, v)
}

// Drop 删除一张表。
func (db *DB) Drop(v interface{}) error {
	return drop(db, v)
}

// Truncate 清空一张表。
func (db *DB) Truncate(v interface{}) error {
	return truncate(db, v)
}

// Begin 开始一个新的事务
func (db *DB) Begin() (*Tx, error) {
	tx, err := db.stdDB.Begin()
	if err != nil {
		return nil, err
	}

	return &Tx{
		db:    db,
		stdTx: tx,
	}, nil
}
