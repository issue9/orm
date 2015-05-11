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
