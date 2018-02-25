// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"database/sql"
	"reflect"

	"github.com/issue9/orm/forward"
)

// Tx 事务对象
type Tx struct {
	db    *DB
	stdTx *sql.Tx
}

// StdTx 返回标准库的 *sql.Tx 对象。
func (tx *Tx) StdTx() *sql.Tx {
	return tx.stdTx
}

// Query 执行一条查询语句。
// 具体参数说明可参考 forward.Engine 接口文档。
func (tx *Tx) Query(replace bool, query string, args ...interface{}) (*sql.Rows, error) {
	if replace {
		query = tx.db.replacer.Replace(query)
	}

	if err := tx.db.dialect.ReplaceMarks(&query); err != nil {
		return nil, err
	}

	return tx.stdTx.Query(query, args...)
}

// Exec 执行一条SQL语句。
// 具体参数说明可参考 forward.Engine 接口文档。
func (tx *Tx) Exec(replace bool, query string, args ...interface{}) (sql.Result, error) {
	if replace {
		query = tx.db.replacer.Replace(query)
	}

	if err := tx.db.dialect.ReplaceMarks(&query); err != nil {
		return nil, err
	}

	return tx.stdTx.Exec(query, args...)
}

// Prepare 将一条 SQL 语句进行预编译。
// 具体参数说明可参考 forward.Engine 接口文档。
func (tx *Tx) Prepare(replace bool, query string) (*sql.Stmt, error) {
	if replace {
		query = tx.db.replacer.Replace(query)
	}

	if err := tx.db.dialect.ReplaceMarks(&query); err != nil {
		return nil, err
	}

	return tx.stdTx.Prepare(query)
}

// Dialect 返回对应的Dialect实例
func (tx *Tx) Dialect() forward.Dialect {
	return tx.db.Dialect()
}

// Commit 提交事务。
// 提交之后，整个Tx对象将不再有效。
func (tx *Tx) Commit() error {
	return tx.stdTx.Commit()
}

// Roolback 回滚事务。
// 回滚之后，整个Tx对象将不再有效。
func (tx *Tx) Rollback() error {
	return tx.stdTx.Rollback()
}

// Insert 插入一个或多个数据。
func (tx *Tx) Insert(v interface{}) (sql.Result, error) {
	return insert(tx, v)
}

// Select 读数据并锁定
func (tx *Tx) Select(v interface{}) error {
	return find(tx, v)
}

// ForUpdate 读数据并锁定
func (tx *Tx) ForUpdate(v interface{}) error {
	return forUpdate(tx, v)
}

// InsertMany 插入多条相同的数据。若需要向某张表中插入多条记录，
// InsertMany()会比Insert()性能上好很多。
// 与DB::Insert()方法最大的不同在于:
//  // Insert()可以每个参数的类型都不一样：
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
	case reflect.Array, reflect.Slice:
		if !tx.Dialect().SupportInsertMany() {
			for i := 0; i < rval.Len(); i++ {
				if _, err := tx.Insert(rval.Index(i).Interface()); err != nil {
					return err
				}
			}
			return nil
		}
		// 支持多个插入，则由此处跳出 switch
	default:
		return ErrInvalidKind
	}

	//sql := new(bytes.Buffer)
	sql, err := buildInsertManySQL(tx, rval)
	if err != nil {
		return err
	}

	if _, err := sql.Exec(true); err != nil {
		return err
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
func (tx *Tx) Count(v interface{}) (int, error) {
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

// SQL 返回一个 forward.SQL 实例。
func (tx *Tx) SQL() *forward.SQL {
	return forward.NewSQL(tx)
}
