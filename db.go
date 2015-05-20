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

// 数据库操作实例。
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
// 关闭之后，之前通过DB.StdDB()返回的实例也将失效。
// 通过调用DB.StdDB().Close()也将使当前实例失效。
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

// 执行一条查询语句，并返回相应的sql.Rows实例。
// 功能基本上等同于标准库database/sql的DB.Query()
// replace指示是否替换掉语句中的占位符，语句中可以指定两种占位符：
// - # 表示一个表名前缀；
// - {} 表示一对Quote字符。如：
//  select * from #user where {group}=1
// 在replace为false时，将原样输出，否则将被转换成以下字符串(以myql为例，假设当前的prefix为p_)
//  select * from prefix_user where `group`=1
func (db *DB) Query(replace bool, query string, args ...interface{}) (*sql.Rows, error) {
	if replace {
		query = db.replacer.Replace(query)
	}
	return db.stdDB.Query(query, args...)
}

// 功能等同于database/sql的DB.Exec()。
// replace参数可参考DB.Query()的说明。
func (db *DB) Exec(replace bool, query string, args ...interface{}) (sql.Result, error) {
	if replace {
		query = db.replacer.Replace(query)
	}
	return db.stdDB.Exec(query, args...)
}

// 功能等同于database/sql的DB.Prepare()。
// replace参数可参考DB.Query()的说明。
func (db *DB) Prepare(replace bool, query string) (*sql.Stmt, error) {
	if replace {
		query = db.replacer.Replace(query)
	}
	return db.stdDB.Prepare(query)
}

// 插入一个或是多个数据。v可以是多个不同类型的结构指针，
func (db *DB) Insert(v ...interface{}) error {
	return insertMult(db, v...)
}

// 删除一个或是多个数据。v可以是多个不同类型的结构指针，
// 查找条件以结构体定义的主键或是唯一约束(在没有主键的情况下)来查找，
// 若两者都不存在，则将返回error
func (db *DB) Delete(v ...interface{}) error {
	return deleteMult(db, v...)
}

// 更新一个或是多个数据。v可以是多个不同类型的结构指针，
// 查找条件以结构体定义的主键或是唯一约束(在没有主键的情况下)来查找，
// 若两者都不存在，则将返回error
func (db *DB) Update(v ...interface{}) error {
	return updateMult(db, v...)
}

// 查询一个或是多个数据。v可以是多个不同类型的结构指针，
// 查找条件以结构体定义的主键或是唯一约束(在没有主键的情况下)来查找，
// 若两者都不存在，则将返回error
func (db *DB) Select(v ...interface{}) error {
	return findMult(db, v...)
}

// 创建一张或是多张表。v可以是多个不同类型的结构指针。
func (db *DB) Create(v ...interface{}) error {
	return createMult(db, v...)
}

// 删除一张或是多张表。v可以是结构体指针或是表名字符串
func (db *DB) Drop(v ...interface{}) error {
	return dropMult(db, v...)
}

// 清空一张或是多张表。v可以是结构体指针或是表名字符串
func (db *DB) Truncate(v ...interface{}) error {
	return truncateMult(db, v...)
}

// 通过一组where()语句来定位数据。
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

// 获取当前实例的表名前缀
func (db *DB) Prefix() string {
	return db.tablePrefix
}

// 事务对象
type Tx struct {
	db    *DB
	stdTx *sql.Tx
}

// 返回标准库的*sql.Tx对象。
func (t *Tx) StdTx() *sql.Tx {
	return t.stdTx
}

// 执行一条查询语句，具体功能参考DB::Query()
func (tx *Tx) Query(replace bool, query string, args ...interface{}) (*sql.Rows, error) {
	if replace {
		query = tx.db.replacer.Replace(query)
	}
	return tx.stdTx.Query(query, args...)
}

// 执行一条SQL语句，具体功能参考DB::Exec()
func (tx *Tx) Exec(replace bool, query string, args ...interface{}) (sql.Result, error) {
	if replace {
		query = tx.db.replacer.Replace(query)
	}
	return tx.stdTx.Exec(query, args...)
}

// 将一条SQL语句进行预编译，具体功能参考DB::Prepare()
func (tx *Tx) Prepare(replace bool, query string) (*sql.Stmt, error) {
	if replace {
		query = tx.db.replacer.Replace(query)
	}
	return tx.stdTx.Prepare(query)
}

// 返回对应的Dialect实例
func (tx *Tx) Dialect() Dialect {
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
func (tx *Tx) Insert(v ...interface{}) error {
	return insertMult(tx, v...)
}

// 更新一个或多个类型。
func (tx *Tx) Update(v ...interface{}) error {
	return updateMult(tx, v...)
}

// 删除一个或是多个数据。
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

// 返回Where实例。
func (tx *Tx) Where(cond string, args ...interface{}) *Where {
	w := newWhere(tx)
	return w.And(cond, args...)
}

// 获取当前实例的表名前缀
func (tx *Tx) Prefix() string {
	return tx.db.tablePrefix
}
