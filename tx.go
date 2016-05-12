// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"database/sql"
	"reflect"

	"github.com/issue9/orm/forward"
)

// 事务对象
type Tx struct {
	db    *DB
	stdTx *sql.Tx
}

// 返回标准库的*sql.Tx对象。
func (tx *Tx) StdTx() *sql.Tx {
	return tx.stdTx
}

// 执行一条查询语句，具体功能参考DB::Query()
func (tx *Tx) Query(replace bool, query string, args ...interface{}) (*sql.Rows, error) {
	if replace {
		query = tx.db.replacer.Replace(query)
	}

	if err := tx.db.dialect.ReplaceMarks(&query); err != nil {
		return nil, err
	}

	return tx.stdTx.Query(query, args...)
}

// 执行一条SQL语句，具体功能参考DB::Exec()
func (tx *Tx) Exec(replace bool, query string, args ...interface{}) (sql.Result, error) {
	if replace {
		query = tx.db.replacer.Replace(query)
	}

	if err := tx.db.dialect.ReplaceMarks(&query); err != nil {
		return nil, err
	}

	return tx.stdTx.Exec(query, args...)
}

// 将一条SQL语句进行预编译，具体功能参考DB::Prepare()
func (tx *Tx) Prepare(replace bool, query string) (*sql.Stmt, error) {
	if replace {
		query = tx.db.replacer.Replace(query)
	}

	if err := tx.db.dialect.ReplaceMarks(&query); err != nil {
		return nil, err
	}

	return tx.stdTx.Prepare(query)
}

// 返回对应的Dialect实例
func (tx *Tx) Dialect() forward.Dialect {
	return tx.db.Dialect()
}

// 提交事务。
// 提交之后，整个Tx对象将不再有效。
func (tx *Tx) Commit() error {
	return tx.stdTx.Commit()
}

// 回滚事务。
// 回滚之后，整个Tx对象将不再有效。
func (tx *Tx) Rollback() error {
	return tx.stdTx.Rollback()
}

// 插入一个或多个数据。
func (tx *Tx) Insert(v interface{}) (sql.Result, error) {
	return insert(tx, v)
}

func (tx *Tx) Select(v interface{}) error {
	return find(tx, v)
}

// 插入多条相同的数据。若需要向某张表中插入多条记录，
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

// 更新一条类型。
func (tx *Tx) Update(v interface{}) (sql.Result, error) {
	return update(tx, v, false)
}

// 更新一条类型。
// 零值也会被提交。
func (tx *Tx) UpdateZero(v interface{}) (sql.Result, error) {
	return update(tx, v, true)
}

// 删除一条数据。
func (tx *Tx) Delete(v interface{}) (sql.Result, error) {
	return del(tx, v)
}

// 查询符合v条件的记录数量。
// v中的所有非零字段都将参与查询。
func (tx *Tx) Count(v interface{}) (int, error) {
	return count(tx, v)
}

// 创建数据表。
func (tx *Tx) Create(v interface{}) error {
	return create(tx, v)

}

// 删除表结构及数据。
func (tx *Tx) Drop(v interface{}) error {
	return drop(tx, v)
}

// 清除表内容，重置ai，但保留表结构。
func (tx *Tx) Truncate(v interface{}) error {
	return truncate(tx, v)
}

// 插入一个或多个数据。
func (tx *Tx) MultInsert(objs ...interface{}) error {
	for _, v := range objs {
		if _, err := tx.Insert(v); err != nil {
			return err
		}
	}
	return nil
}

// 选择符合要求的一条或是多条记录。
func (tx *Tx) MultSelect(objs ...interface{}) error {
	for _, v := range objs {
		if err := tx.Select(v); err != nil {
			return err
		}
	}
	return nil
}

// 更新一条或多条类型。
func (tx *Tx) MultUpdate(objs ...interface{}) error {
	for _, v := range objs {
		if _, err := tx.Update(v); err != nil {
			return err
		}
	}
	return nil
}

// 更新一条或多条类型。
func (tx *Tx) MultUpdateZero(objs ...interface{}) error {
	for _, v := range objs {
		if _, err := tx.UpdateZero(v); err != nil {
			return err
		}
	}
	return nil
}

// 删除一条或是多条数据。
func (tx *Tx) MultDelete(objs ...interface{}) error {
	for _, v := range objs {
		if _, err := tx.Delete(v); err != nil {
			return err
		}
	}
	return nil
}

// 创建数据表。
func (tx *Tx) MultCreate(objs ...interface{}) error {
	for _, v := range objs {
		if err := tx.Create(v); err != nil {
			return err
		}
	}
	return nil
}

// 删除表结构及数据。
func (tx *Tx) MultDrop(objs ...interface{}) error {
	for _, v := range objs {
		if err := tx.Drop(v); err != nil {
			return err
		}
	}

	return nil
}

// 清除表内容，重置ai，但保留表结构。
func (tx *Tx) MultTruncate(objs ...interface{}) error {
	for _, v := range objs {
		if err := tx.Truncate(v); err != nil {
			return err
		}
	}
	return nil
}

// 返回SQL实例。
func (tx *Tx) Where(cond string, args ...interface{}) *SQL {
	w := newSQL(tx)
	return w.And(cond, args...)
}

// 获取当前实例的表名前缀
func (tx *Tx) Prefix() string {
	return tx.db.tablePrefix
}

func (tx *Tx) SQL() *SQL {
	return newSQL(tx)
}
