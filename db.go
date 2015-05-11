// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"database/sql"
)

type DB struct {
	stdDB   *sql.DB
	dialect Dialect
	prefix  string
}

func NewDB(driverName, dataSourceName, prefix string, dialect Dialect) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	return &DB{
		stdDB:   db,
		dialect: dialect,
		prefix:  prefix,
	}, nil
}

func (db *DB) Close() error {
	return db.Close()
}

// 返回标准包中的sql.DB指针
func (db *DB) StdDB() *sql.DB {
	return db.stdDB
}

func (db *DB) Dialect() Dialect {
	return db.dialect
}

func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.stdDB.Query(query, args...)
}

func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.stdDB.Exec(query, args...)
}

func (db *DB) Prepare(query string) (*sql.Stmt, error) {
	return db.stdDB.Prepare(query)
}

// 插入一个或是多个数据，v只能是struct或是Struct指针
func (db *DB) Insert(v interface{}) error {
	return insertMult(db, v)
}

func (db *DB) Delete(v interface{}) error {
	return deleteMult(db, v)
}

func (db *DB) Update(v interface{}) error {
	return updateMult(db, v)
}

func (db *DB) Select(v interface{}) error {
	return findMult(db, v)
}

func (db *DB) Create(v interface{}) error {
	return createMult(db, v)
}

// 删除表
func (db *DB) Drop(v interface{}) error {
	return nil
}

func (db *DB) Truncate(v interface{}) error {
	return nil
}

func (db *DB) Where(cond string, args ...interface{}) *Where {
	w := newWhere(db)
	return w.And(cond, args...)
}

// 事务对象
type Tx struct {
	db    *DB
	stdTx *sql.Tx
}

// 对orm/core.DB.DB()的实现。返回当前数据库对应的*sql.DB
func (t *Tx) StdTx() *sql.Tx {
	return t.stdTx
}

func (tx *Tx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return tx.stdTx.Query(query, args...)
}

func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tx.stdTx.Exec(query, args...)
}

func (tx *Tx) Prepare(query string) (*sql.Stmt, error) {
	return tx.stdTx.Prepare(query)
}

// 对orm/core.DB.Dialect()的实现。返回当前数据库对应的Dialect
func (tx *Tx) Dialect() Dialect {
	return tx.db.Dialect()
}

// 提交事务
// 提交之后，整个Tx对象将不再有效。
func (tx *Tx) Commit() error {
	return tx.stdTx.Commit()
}

// 回滚事务
func (tx *Tx) Rollback() error {
	return tx.stdTx.Rollback()
}

// 插入一个或多个数据。
// v可以是struct或是相同struct组成的数组。
// 若v中指定了自增字段，则该字段的值在插入数据库时，
// 会被自动忽略。
func (tx *Tx) Insert(v interface{}) error {
	return insertMult(tx, v)
}

// 更新一个或多个类型。
// 更新依据为每个对象的主键或是唯一索引列。
// 若不存在此两个类型的字段，则返回错误信息。
func (tx *Tx) Update(v interface{}) error {
	return updateMult(tx, v)
}

// 删除指定的数据对象。
func (tx *Tx) Delete(v interface{}) error {
	return deleteMult(tx, v)
}

// 创建数据表。
func (tx *Tx) Create(v ...interface{}) error {
	return createMult(tx, v...)
}

// 删除表结构及数据。
func (tx *Tx) Drop(tableName string) error {
	_, err := tx.Exec("DROP TABLE "+tableName, nil)
	return err
}

// 清除表内容，但保留表结构。
func (t *Tx) Truncate(tableName string) error {
	_, err := t.Exec(t.Dialect().TruncateTableSQL(tableName), nil)
	return err
}

func (tx *Tx) Where(cond string, args ...interface{}) *Where {
	w := newWhere(tx)
	return w.And(cond, args...)
}
