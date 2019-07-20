// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/issue9/orm/v2/fetch"
)

// Tx 事务对象
type Tx struct {
	*sql.Tx
	db  *DB
	sql *SQL
}

// Begin 开始一个新的事务
func (db *DB) Begin() (*Tx, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}

	return db.begin(tx)
}

// BeginTx 开始一个新的事务
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return db.begin(tx)
}

func (db *DB) begin(tx *sql.Tx) (*Tx, error) {
	inst := &Tx{
		Tx: tx,
		db: db,
	}
	inst.sql = &SQL{engine: inst}

	return inst, nil
}

// Query 执行一条查询语句。
func (tx *Tx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return tx.QueryContext(context.Background(), query, args...)
}

// QueryContext 执行一条查询语句。
func (tx *Tx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	query = tx.db.replacer.Replace(query)
	query, args, err := tx.db.dialect.SQL(query, args)
	if err != nil {
		return nil, err
	}

	return tx.Tx.QueryContext(ctx, query, args...)
}

// QueryRow 执行一条查询语句。
//
// 如果生成语句出错，则会 panic
func (tx *Tx) QueryRow(query string, args ...interface{}) *sql.Row {
	return tx.QueryRowContext(context.Background(), query, args...)
}

// QueryRowContext 执行一条查询语句。
//
// 如果生成语句出错，则会 panic
func (tx *Tx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	query = tx.db.replacer.Replace(query)
	query, args, err := tx.db.dialect.SQL(query, args)
	if err != nil {
		panic(err)
	}

	return tx.Tx.QueryRowContext(ctx, query, args...)
}

// Exec 执行一条 SQL 语句。
func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tx.ExecContext(context.Background(), query, args...)
}

// ExecContext 执行一条 SQL 语句。
func (tx *Tx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	query = tx.db.replacer.Replace(query)
	query, args, err := tx.db.dialect.SQL(query, args)
	if err != nil {
		return nil, err
	}

	return tx.Tx.ExecContext(ctx, query, args...)
}

// Prepare 将一条 SQL 语句进行预编译。
func (tx *Tx) Prepare(query string) (*sql.Stmt, error) {
	return tx.PrepareContext(context.Background(), query)
}

// PrepareContext 将一条 SQL 语句进行预编译。
func (tx *Tx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	query = tx.db.replacer.Replace(query)
	query, _, err := tx.db.dialect.SQL(query, nil)
	if err != nil {
		return nil, err
	}

	return tx.Tx.PrepareContext(ctx, query)
}

// Dialect 返回对应的 Dialect 实例
func (tx *Tx) Dialect() Dialect {
	return tx.db.Dialect()
}

// LastInsertID 插入数据，并获取其自增的 ID。
func (tx *Tx) LastInsertID(v interface{}) (int64, error) {
	return lastInsertID(tx, v)
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
// max 表示一次最多插入的数量，如果超过此值，会分批执行，
// 但是依然在一个事务中完成。
//
// 与 MultInsert() 方法最大的不同在于:
//  // MultInsert() 可以每个参数的类型都不一样：
//  vs := []interface{}{&user{...}, &userInfo{...}}
//  db.Insert(vs...)
//  // db.InsertMany(vs) // 这里将出错，数组的元素的类型必须相同。
//  us := []*users{&user{}, &user{}}
//  db.InsertMany(us)
//  db.Insert(us...) // 这样也行，但是性能会差好多
func (tx *Tx) InsertMany(v interface{}, max int) error {
	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	switch rval.Kind() {
	case reflect.Struct: // 单个元素
		_, err := tx.Insert(v)
		return err
	case reflect.Array, reflect.Slice: // 跳出 switch
	default:
		return fetch.ErrInvalidKind
	}

	if rval.Len() == 0 { // 为空，则什么也不做
		return nil
	}

	for i := 0; i < rval.Len(); i += max {
		j := i + max
		if j > rval.Len() {
			j = rval.Len()
		}
		sql, err := buildInsertManySQL(tx, rval.Slice(i, j))
		if err != nil {
			return err
		}

		if _, err = sql.Exec(); err != nil {
			return err
		}
	}

	return nil
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
// 如果 v 为空值，则相当于 select count(*) from xx;
func (tx *Tx) Count(v interface{}) (int64, error) {
	return count(tx, v)
}

// Create 创建数据表。
func (tx *Tx) Create(v interface{}) error {
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
