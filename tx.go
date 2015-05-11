// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"database/sql"
)

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
