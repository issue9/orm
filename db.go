// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"database/sql"
	"strings"
)

const (
	tablePrefixPlaceholder = "#"
	openQuotePlaceholder   = "{"
	closeQuotePlaceholder  = "}"
)

// 可以以对象的方式存取数据库。
type DB struct {
	stdDB       *sql.DB
	dialect     Dialect
	tablePrefix string
	replacer    *strings.Replacer
}

// 声明一个新的DB实例。
func NewDB(driverName, dataSourceName, tablePrefix string, dialect Dialect) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	l, r := dialect.QuoteTuple()
	return &DB{
		stdDB:       db,
		dialect:     dialect,
		tablePrefix: tablePrefix,
		replacer: strings.NewReplacer(
			tablePrefixPlaceholder, tablePrefix,
			openQuotePlaceholder, string(l),
			closeQuotePlaceholder, string(r),
		),
	}, nil
}

// 关闭当前数据库，释放所有的链接。
// 关闭之后，之前通过db.StdDB()返回的实例也将失效。
func (db *DB) Close() error {
	return db.stdDB.Close()
}

// 返回标准包中的sql.DB指针。
func (db *DB) StdDB() *sql.DB {
	return db.stdDB
}

// 返回对应的Dialect接口实例。
func (db *DB) Dialect() Dialect {
	return db.dialect
}

func (db *DB) Query(replace bool, query string, args ...interface{}) (*sql.Rows, error) {
	if replace {
		query = db.replacer.Replace(query)
	}
	return db.stdDB.Query(query, args...)
}

func (db *DB) Exec(replace bool, query string, args ...interface{}) (sql.Result, error) {
	if replace {
		query = db.replacer.Replace(query)
	}
	return db.stdDB.Exec(query, args...)
}

func (db *DB) Prepare(replace bool, query string) (*sql.Stmt, error) {
	if replace {
		query = db.replacer.Replace(query)
	}
	return db.stdDB.Prepare(query)
}

// 插入一个或是多个数据。v可以是struct或是struct指针，
// 每个v的o类型可以不相同。
func (db *DB) Insert(v ...interface{}) error {
	return insertMult(db, v...)
}

func (db *DB) Delete(v ...interface{}) error {
	return deleteMult(db, v...)
}

func (db *DB) Update(v ...interface{}) error {
	return updateMult(db, v...)
}

func (db *DB) Select(v ...interface{}) error {
	return findMult(db, v...)
}

func (db *DB) Create(v ...interface{}) error {
	return createMult(db, v...)
}

// 删除表
func (db *DB) Drop(v ...interface{}) error {
	return dropMult(db, v...)
}

func (db *DB) Truncate(v ...interface{}) error {
	return truncateMult(db, v...)
}

func (db *DB) Where(cond string, args ...interface{}) *Where {
	w := newWhere(db)
	return w.And(cond, args...)
}

// 开始一个新的事务
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

func (db *DB) prefix() string {
	return db.tablePrefix
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

func (tx *Tx) Query(replace bool, query string, args ...interface{}) (*sql.Rows, error) {
	if replace {
		query = tx.db.replacer.Replace(query)
	}
	return tx.stdTx.Query(query, args...)
}

func (tx *Tx) Exec(replace bool, query string, args ...interface{}) (sql.Result, error) {
	if replace {
		query = tx.db.replacer.Replace(query)
	}
	return tx.stdTx.Exec(query, args...)
}

func (tx *Tx) Prepare(replace bool, query string) (*sql.Stmt, error) {
	if replace {
		query = tx.db.replacer.Replace(query)
	}
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
func (tx *Tx) Insert(v ...interface{}) error {
	return insertMult(tx, v...)
}

// 更新一个或多个类型。
// 更新依据为每个对象的主键或是唯一索引列。
// 若不存在此两个类型的字段，则返回错误信息。
func (tx *Tx) Update(v ...interface{}) error {
	return updateMult(tx, v...)
}

// 删除指定的数据对象。
func (tx *Tx) Delete(v ...interface{}) error {
	return deleteMult(tx, v...)
}

// 创建数据表。
func (tx *Tx) Create(v ...interface{}) error {
	return createMult(tx, v...)
}

// 删除表结构及数据。
func (tx *Tx) Drop(v ...interface{}) error {
	return dropMult(tx, v...)
}

// 清除表内容，但保留表结构。
func (tx *Tx) Truncate(v ...interface{}) error {
	return truncateMult(tx, v...)
}

func (tx *Tx) Where(cond string, args ...interface{}) *Where {
	w := newWhere(tx)
	return w.And(cond, args...)
}

func (tx *Tx) prefix() string {
	return tx.db.tablePrefix
}
