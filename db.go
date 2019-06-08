// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"context"
	"database/sql"
	"strings"

	"github.com/issue9/orm/v2/model"
)

// DB 数据库操作实例。
type DB struct {
	*sql.DB
	dialect     Dialect
	tablePrefix string
	replacer    *strings.Replacer
	sql         *SQL
	models      *model.Models
}

// NewDB 声明一个新的 DB 实例。
func NewDB(driverName, dataSourceName, tablePrefix string, dialect Dialect) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	return NewDBWithStdDB(db, tablePrefix, dialect)
}

// NewDBWithStdDB 从 sql.DB 构建一个 DB 实例。
func NewDBWithStdDB(db *sql.DB, tablePrefix string, dialect Dialect) (*DB, error) {
	l, r := dialect.QuoteTuple()
	inst := &DB{
		DB:          db,
		dialect:     dialect,
		tablePrefix: tablePrefix,
		replacer: strings.NewReplacer(
			"#", tablePrefix,
			"{", string(l),
			"}", string(r),
		),
		models: model.NewModels(),
	}

	inst.sql = &SQL{engine: inst}

	return inst, nil
}

// Dialect 返回对应的 Dialect 接口实例。
func (db *DB) Dialect() Dialect {
	return db.dialect
}

// Close 关闭连接
//
// 同时会清除缓存的模型数据
func (db *DB) Close() error {
	db.models.Clear()

	return db.DB.Close()
}

// QueryRow 执行一条查询语句，并返回相应的 sql.Rows 实例。
//
// 如果生成语句出错，则会 panic
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.QueryRowContext(context.Background(), query, args...)
}

// QueryRowContext 执行一条查询语句，并返回相应的 sql.Rows 实例。
//
// 如果生成语句出错，则会 panic
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	query = db.replacer.Replace(query)
	query, err := db.dialect.SQL(query)
	if err != nil {
		panic(err)
	}

	return db.DB.QueryRowContext(ctx, query, args...)
}

// Query 执行一条查询语句，并返回相应的 sql.Rows 实例。
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.QueryContext(context.Background(), query, args...)
}

// QueryContext 执行一条查询语句，并返回相应的 sql.Rows 实例。
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	query = db.replacer.Replace(query)
	query, err := db.dialect.SQL(query)
	if err != nil {
		return nil, err
	}

	return db.DB.QueryContext(ctx, query, args...)
}

// Exec 执行 SQL 语句。
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.ExecContext(context.Background(), query, args...)
}

// ExecContext 执行 SQL 语句。
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	query = db.replacer.Replace(query)
	query, err := db.dialect.SQL(query)
	if err != nil {
		return nil, err
	}

	return db.DB.ExecContext(ctx, query, args...)
}

// Prepare 预编译查询语句。
func (db *DB) Prepare(query string) (*sql.Stmt, error) {
	return db.PrepareContext(context.Background(), query)
}

// PrepareContext 预编译查询语句。
func (db *DB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	query = db.replacer.Replace(query)
	query, err := db.dialect.SQL(query)
	if err != nil {
		return nil, err
	}

	return db.DB.PrepareContext(ctx, query)
}

// LastInsertID 插入数据，并获取其自增的 ID。
func (db *DB) LastInsertID(v interface{}) (int64, error) {
	return lastInsertID(db, v)
}

// Insert 插入数据，若需一次性插入多条数据，请使用 tx.Insert()。
func (db *DB) Insert(v interface{}) (sql.Result, error) {
	return insert(db, v)
}

// Delete 删除符合条件的数据。
//
// 查找条件以结构体定义的主键或是唯一约束(在没有主键的情况下)来查找，
// 若两者都不存在，则将返回 error
func (db *DB) Delete(v interface{}) (sql.Result, error) {
	return del(db, v)
}

// Update 更新数据，零值不会被提交，cols 指定的列，即使是零值也会被更新。
//
// 查找条件以结构体定义的主键或是唯一约束(在没有主键的情况下)来查找，
// 若两者都不存在，则将返回 error
func (db *DB) Update(v interface{}, cols ...string) (sql.Result, error) {
	return update(db, v, cols...)
}

// Select 查询一个符合条件的数据。
//
// 查找条件以结构体定义的主键或是唯一约束(在没有主键的情况下 ) 来查找，
// 若两者都不存在，则将返回 error
// 若没有符合条件的数据，将不会对参数v做任何变动。
func (db *DB) Select(v interface{}) error {
	return find(db, v)
}

// Count 查询符合 v 条件的记录数量。
// v 中的所有非零字段都将参与查询。
// 如果 v 为空值，则相当于 select count(*) from xx;
//
// 若需要复杂的查询方式，请构建 SelectStmt 对象查询。
func (db *DB) Count(v interface{}) (int64, error) {
	return count(db, v)
}

// Create 创建一张表。
func (db *DB) Create(v interface{}) error {
	if !db.Dialect().TransactionalDDL() {
		return create(db, v)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if err = create(tx, v); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Drop 删除一张表。
func (db *DB) Drop(v interface{}) error {
	return drop(db, v)
}

// Truncate 清空一张表。
func (db *DB) Truncate(v interface{}) error {
	m, err := db.NewModel(v)
	if err != nil {
		return err
	}

	sqls := db.Dialect().TruncateTableSQL(m)

	// 单语句，直接执行
	if len(sqls) == 1 {
		_, err = db.Exec(sqls[0])
		return err
	}

	// 多语句，则初如化事务，再执行

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	for _, sql := range sqls {
		if _, err := tx.Exec(sql); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// MultInsert 插入一个或多个数据。
func (db *DB) MultInsert(objs ...interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if err := tx.MultInsert(objs...); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

// MultSelect 选择符合要求的一条或是多条记录。
func (db *DB) MultSelect(objs ...interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if err := tx.MultSelect(objs...); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

// MultUpdate 更新一条或多条类型。
func (db *DB) MultUpdate(objs ...interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if err := tx.MultUpdate(objs...); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

// MultDelete 删除一条或是多条数据。
func (db *DB) MultDelete(objs ...interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if err := tx.MultDelete(objs...); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

// MultCreate 创建数据表。
func (db *DB) MultCreate(objs ...interface{}) error {
	if !db.Dialect().TransactionalDDL() {
		for _, v := range objs {
			if err := db.Create(v); err != nil {
				return err
			}
		}
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if err := tx.MultCreate(objs...); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

// MultDrop 删除表结构及数据。
func (db *DB) MultDrop(objs ...interface{}) error {
	if !db.Dialect().TransactionalDDL() {
		for _, v := range objs {
			if err := db.Drop(v); err != nil {
				return err
			}
		}
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if err := tx.MultDrop(objs...); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

// MultTruncate 清除表内容，重置 ai，但保留表结构。
func (db *DB) MultTruncate(objs ...interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if err := tx.MultTruncate(objs...); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

// SQL 返回 SQL 实例
func (db *DB) SQL() *SQL {
	return db.sql
}
