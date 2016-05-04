// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"database/sql"
	"strings"

	"github.com/issue9/orm/forward"
)

const (
	tablePrefixPlaceholder = "#"
	openQuotePlaceholder   = "{"
	closeQuotePlaceholder  = "}"
)

// 数据库操作实例。
type DB struct {
	stdDB       *sql.DB
	dialect     forward.Dialect
	tablePrefix string
	replacer    *strings.Replacer
}

// 声明一个新的DB实例。
func NewDB(driverName, dataSourceName, tablePrefix string, dialect forward.Dialect) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	return NewDBWithStdDB(db, tablePrefix, dialect)
}

// 从sql.DB构建一个DB实例。
func NewDBWithStdDB(db *sql.DB, tablePrefix string, dialect forward.Dialect) (*DB, error) {
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
func (db *DB) Dialect() forward.Dialect {
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

	if err := db.dialect.ReplaceMarks(&query); err != nil {
		return nil, err
	}

	return db.stdDB.Query(query, args...)
}

// 功能等同于database/sql的DB.Exec()。
// replace参数可参考DB.Query()的说明。
func (db *DB) Exec(replace bool, query string, args ...interface{}) (sql.Result, error) {
	if replace {
		query = db.replacer.Replace(query)
	}

	if err := db.dialect.ReplaceMarks(&query); err != nil {
		return nil, err
	}

	return db.stdDB.Exec(query, args...)
}

// 功能等同于database/sql的DB.Prepare()。
// replace参数可参考DB.Query()的说明。
func (db *DB) Prepare(replace bool, query string) (*sql.Stmt, error) {
	if replace {
		query = db.replacer.Replace(query)
	}

	if err := db.dialect.ReplaceMarks(&query); err != nil {
		return nil, err
	}

	return db.stdDB.Prepare(query)
}

// 插入数据，若需一次性插入多条数据，请使用tx.Insert()。
func (db *DB) Insert(v interface{}) (sql.Result, error) {
	return insert(db, v)
}

// 删除符合条件的数据。
// 查找条件以结构体定义的主键或是唯一约束(在没有主键的情况下)来查找，
// 若两者都不存在，则将返回error
func (db *DB) Delete(v interface{}) (sql.Result, error) {
	return del(db, v)
}

// 更新数据，零值不会被提交。
// 查找条件以结构体定义的主键或是唯一约束(在没有主键的情况下)来查找，
// 若两者都不存在，则将返回error
func (db *DB) Update(v interface{}) (sql.Result, error) {
	return update(db, v, false)
}

// 更新数据，包括零值的内容。
func (db *DB) UpdateZero(v interface{}) (sql.Result, error) {
	return update(db, v, true)
}

// 查询一个符合条件的数据。
// 查找条件以结构体定义的主键或是唯一约束(在没有主键的情况下)来查找，
// 若两者都不存在，则将返回error
// 若没有符合条件的数据，将不会对参数v做任何变动。
func (db *DB) Select(v interface{}) error {
	return find(db, v)
}

// 查询符合v条件的记录数量。
// v中的所有非零字段都将参与查询。
func (db *DB) Count(v interface{}) (int, error) {
	return count(db, v)
}

// 创建一张表。
func (db *DB) Create(v interface{}) error {
	return create(db, v)
}

// 删除一张表。
func (db *DB) Drop(v interface{}) error {
	return drop(db, v)
}

// 清空一张表。
func (db *DB) Truncate(v interface{}) error {
	return truncate(db, v)
}

// 通过SQL实例。
func (db *DB) Where(cond string, args ...interface{}) *SQL {
	w := newSQL(db)
	return w.And(cond, args...)
}

func (db *DB) SQL() *SQL {
	return newSQL(db)
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
