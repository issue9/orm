// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/issue9/orm/fetch"
)

// Tx 事务对象
type Tx struct {
	db    *DB
	stdTx *sql.Tx
	sql   *SQL
}

// Begin 开始一个新的事务
func (db *DB) Begin() (*Tx, error) {
	tx, err := db.stdDB.Begin()
	if err != nil {
		return nil, err
	}

	inst := &Tx{
		db:    db,
		stdTx: tx,
	}
	inst.sql = &SQL{engine: inst}

	return inst, nil
}

// StdTx 返回标准库的 *sql.Tx 对象。
func (tx *Tx) StdTx() *sql.Tx {
	return tx.stdTx
}

// Query 执行一条查询语句。
func (tx *Tx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	query = tx.db.replacer.Replace(query)
	query, err := tx.db.dialect.SQL(query)
	if err != nil {
		return nil, err
	}

	return tx.stdTx.Query(query, args...)
}

// QueryContext 执行一条查询语句。
func (tx *Tx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	query = tx.db.replacer.Replace(query)
	query, err := tx.db.dialect.SQL(query)
	if err != nil {
		return nil, err
	}

	return tx.stdTx.QueryContext(ctx, query, args...)
}

// Exec 执行一条 SQL 语句。
func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	query = tx.db.replacer.Replace(query)
	query, err := tx.db.dialect.SQL(query)
	if err != nil {
		return nil, err
	}

	return tx.stdTx.Exec(query, args...)
}

// ExecContext 执行一条 SQL 语句。
func (tx *Tx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	query = tx.db.replacer.Replace(query)
	query, err := tx.db.dialect.SQL(query)
	if err != nil {
		return nil, err
	}

	return tx.stdTx.ExecContext(ctx, query, args...)
}

// Prepare 将一条 SQL 语句进行预编译。
func (tx *Tx) Prepare(query string) (*sql.Stmt, error) {
	query = tx.db.replacer.Replace(query)
	query, err := tx.db.dialect.SQL(query)
	if err != nil {
		return nil, err
	}

	return tx.stdTx.Prepare(query)
}

// PrepareContext 将一条 SQL 语句进行预编译。
func (tx *Tx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	query = tx.db.replacer.Replace(query)
	query, err := tx.db.dialect.SQL(query)
	if err != nil {
		return nil, err
	}

	return tx.stdTx.PrepareContext(ctx, query)
}

// Dialect 返回对应的 Dialect 实例
func (tx *Tx) Dialect() Dialect {
	return tx.db.Dialect()
}

// Commit 提交事务。
//
// 提交之后，整个 Tx 对象将不再有效。
func (tx *Tx) Commit() error {
	return tx.stdTx.Commit()
}

// Rollback 回滚事务。
//
// 回滚之后，整个 Tx 对象将不再有效。
func (tx *Tx) Rollback() error {
	return tx.stdTx.Rollback()
}

// Insert 插入一个或多个数据。
func (tx *Tx) Insert(v interface{}) (sql.Result, error) {
	return insert(tx, v)
}

// Select 读数据
func (tx *Tx) Select(v interface{}) error {
	return find(tx, v)
}

// ForUpdate 读数据并锁定
func (tx *Tx) ForUpdate(v interface{}) error {
	return forUpdate(tx, v)
}

// InsertMany 插入多条相同的数据。若需要向某张表中插入多条记录，
// InsertMany() 会比 Insert() 性能上好很多。
//
// 与 MultInsert() 方法最大的不同在于:
//  // MultInsert() 可以每个参数的类型都不一样：
//  vs := []interface{}{&user{...}, &userInfo{...}}
//  db.Insert(vs...)
//  // db.InsertMany(vs) // 这里将出错，数组的元素的类型必须相同。
//  us := []*users{&user{}, &user{}}
//  db.InsertMany(us)
//  db.Insert(us...) // 这样也行，但是性能会差好多
func (tx *Tx) InsertMany(v interface{}) error {
	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	switch rval.Kind() {
	case reflect.Struct: // 单个元素
		_, err := tx.Insert(v)
		return err
	case reflect.Array, reflect.Slice: // 支持多个插入，则由此处跳出 switch
		sql, err := buildInsertManySQL(tx, rval)
		if err != nil {
			return err
		}

		_, err = sql.Exec()
		return err
	default:
		return fetch.ErrInvalidKind
	}
}

// Update 更新一条类型。
func (tx *Tx) Update(v interface{}, cols ...string) (sql.Result, error) {
	return update(tx, v, cols...)
}

// Delete 删除一条数据。
func (tx *Tx) Delete(v interface{}) (sql.Result, error) {
	return del(tx, v)
}

// Count 查询符合 v 条件的记录数量。
// v 中的所有非零字段都将参与查询。
func (tx *Tx) Count(v interface{}) (int64, error) {
	return count(tx, v)
}

// Create 创建数据表。
func (tx *Tx) Create(v interface{}) error {
	if !tx.db.Dialect().TransactionalDDL() {
		return tx.db.Create(v)
	}

	return create(tx, v)
}

// Drop 删除表结构及数据。
func (tx *Tx) Drop(v interface{}) error {
	return drop(tx, v)
}

// Truncate 清除表内容，重置 ai，但保留表结构。
func (tx *Tx) Truncate(v interface{}) error {
	return truncate(tx, v)
}

// SQL 返回 SQL 实例
func (tx *Tx) SQL() *SQL {
	return tx.sql
}

// MultInsert 插入一个或多个数据。
func (tx *Tx) MultInsert(objs ...interface{}) error {
	for _, v := range objs {
		if _, err := tx.Insert(v); err != nil {
			return err
		}
	}
	return nil
}

// MultSelect 选择符合要求的一条或是多条记录。
func (tx *Tx) MultSelect(objs ...interface{}) error {
	for _, v := range objs {
		if err := tx.Select(v); err != nil {
			return err
		}
	}
	return nil
}

// MultUpdate 更新一条或多条类型。
func (tx *Tx) MultUpdate(objs ...interface{}) error {
	for _, v := range objs {
		if _, err := tx.Update(v); err != nil {
			return err
		}
	}
	return nil
}

// MultDelete 删除一条或是多条数据。
func (tx *Tx) MultDelete(objs ...interface{}) error {
	for _, v := range objs {
		if _, err := tx.Delete(v); err != nil {
			return err
		}
	}
	return nil
}

// MultCreate 创建数据表。
func (tx *Tx) MultCreate(objs ...interface{}) error {
	if !tx.db.Dialect().TransactionalDDL() {
		return tx.db.MultCreate(objs...)
	}

	for _, v := range objs {
		if err := tx.Create(v); err != nil {
			return err
		}
	}
	return nil
}

// MultDrop 删除表结构及数据。
func (tx *Tx) MultDrop(objs ...interface{}) error {
	for _, v := range objs {
		if err := tx.Drop(v); err != nil {
			return err
		}
	}

	return nil
}

// MultTruncate 清除表内容，重置 ai，但保留表结构。
func (tx *Tx) MultTruncate(objs ...interface{}) error {
	for _, v := range objs {
		if err := tx.Truncate(v); err != nil {
			return err
		}
	}
	return nil
}
