// SPDX-License-Identifier: MIT

package orm

import (
	"context"
	"database/sql"

	"github.com/issue9/orm/v4/core"
	"github.com/issue9/orm/v4/sqlbuilder"
)

// Tx 事务对象
type Tx struct {
	*sql.Tx
	db         *DB
	sqlBuilder *sqlbuilder.SQLBuilder
}

// Begin 开始一个新的事务
func (db *DB) Begin() (*Tx, error) {
	return db.BeginTx(context.Background(), nil)
}

// BeginTx 开始一个新的事务
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	inst := &Tx{
		Tx: tx,
		db: db,
	}
	inst.sqlBuilder = sqlbuilder.New(inst)

	return inst, nil
}

// TablePrefix 返回表名前缀内容内容
func (tx *Tx) TablePrefix() string { return tx.db.tablePrefix }

// Query 执行一条查询语句
func (tx *Tx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return tx.QueryContext(context.Background(), query, args...)
}

// QueryContext 执行一条查询语句
func (tx *Tx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	tx.db.printDebug(query)
	query = tx.db.replacer.Replace(query)
	query, args, err := tx.Dialect().Fix(query, args)
	if err != nil {
		return nil, err
	}

	return tx.Tx.QueryContext(ctx, query, args...)
}

// QueryRow 执行一条查询语句
//
// 如果生成语句出错，则会 panic
func (tx *Tx) QueryRow(query string, args ...interface{}) *sql.Row {
	return tx.QueryRowContext(context.Background(), query, args...)
}

// QueryRowContext 执行一条查询语句
//
// 如果生成语句出错，则会 panic
func (tx *Tx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	tx.db.printDebug(query)
	query = tx.db.replacer.Replace(query)
	query, args, err := tx.Dialect().Fix(query, args)
	if err != nil {
		panic(err)
	}

	return tx.Tx.QueryRowContext(ctx, query, args...)
}

// Exec 执行一条 SQL 语句
func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tx.ExecContext(context.Background(), query, args...)
}

// ExecContext 执行一条 SQL 语句
func (tx *Tx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	tx.db.printDebug(query)
	query = tx.db.replacer.Replace(query)
	query, args, err := tx.Dialect().Fix(query, args)
	if err != nil {
		return nil, err
	}

	return tx.Tx.ExecContext(ctx, query, args...)
}

// Prepare 将一条 SQL 语句进行预编译
func (tx *Tx) Prepare(query string) (*core.Stmt, error) {
	return tx.PrepareContext(context.Background(), query)
}

// PrepareContext 将一条 SQL 语句进行预编译
func (tx *Tx) PrepareContext(ctx context.Context, query string) (*core.Stmt, error) {
	tx.db.printDebug(query)
	query = tx.db.replacer.Replace(query)
	query, orders, err := tx.Dialect().Prepare(query)
	if err != nil {
		return nil, err
	}

	s, err := tx.db.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return core.NewStmt(s, orders), nil
}

// Dialect 返回对应的 Dialect 实例
func (tx *Tx) Dialect() Dialect {
	return tx.db.Dialect()
}

// LastInsertID 插入数据，并获取其自增的 ID
func (tx *Tx) LastInsertID(v TableNamer) (int64, error) {
	return lastInsertID(tx, v)
}

// Insert 插入一个或多个数据
func (tx *Tx) Insert(v TableNamer) (sql.Result, error) {
	return insert(tx, v)
}

// Select 读数据
func (tx *Tx) Select(v TableNamer) error {
	return find(tx, v)
}

// ForUpdate 读数据并锁定
func (tx *Tx) ForUpdate(v TableNamer) error {
	return forUpdate(tx, v)
}

func (tx *Tx) InsertMany(max int, v ...TableNamer) error {
	if len(v) == 0 {
		return nil
	}

	l := len(v)
	for i := 0; i < l; i += max {
		j := i + max
		if j > l {
			j = l
		}
		query, err := buildInsertManySQL(tx, v[i:j]...)
		if err != nil {
			return err
		}

		if _, err = query.Exec(); err != nil {
			return err
		}
	}

	return nil
}

// Update 更新一条类型
func (tx *Tx) Update(v TableNamer, cols ...string) (sql.Result, error) {
	return update(tx, v, cols...)
}

// Delete 删除一条数据
func (tx *Tx) Delete(v TableNamer) (sql.Result, error) {
	return del(tx, v)
}

// Create 创建数据表或是视图
func (tx *Tx) Create(v TableNamer) error {
	return create(tx, v)
}

// Drop 删除表或视图
func (tx *Tx) Drop(v TableNamer) error {
	return drop(tx, v)
}

// Truncate 清除表内容
//
// 会重置 ai，但保留表结构。
func (tx *Tx) Truncate(v TableNamer) error {
	return truncate(tx, v)
}

// SQLBuilder 返回 SQL 实例
func (tx *Tx) SQLBuilder() *sqlbuilder.SQLBuilder {
	return tx.sqlBuilder
}

// MultInsert 插入一个或多个数据
func (tx *Tx) MultInsert(objs ...TableNamer) error {
	for _, v := range objs {
		if _, err := tx.Insert(v); err != nil {
			return err
		}
	}
	return nil
}

// MultSelect 选择符合要求的一条或是多条记录
func (tx *Tx) MultSelect(objs ...TableNamer) error {
	return tx.multDo(tx.Select, objs...)
}

// MultUpdate 更新一条或多条类型
func (tx *Tx) MultUpdate(objs ...TableNamer) error {
	for _, v := range objs {
		if _, err := tx.Update(v); err != nil {
			return err
		}
	}
	return nil
}

// MultDelete 删除一条或是多条数据
func (tx *Tx) MultDelete(objs ...TableNamer) error {
	for _, v := range objs {
		if _, err := tx.Delete(v); err != nil {
			return err
		}
	}
	return nil
}

// MultCreate 创建数据表
func (tx *Tx) MultCreate(objs ...TableNamer) error {
	return tx.multDo(tx.Create, objs...)
}

// MultDrop 删除表结构及数据
func (tx *Tx) MultDrop(objs ...TableNamer) error {
	return tx.multDo(tx.Drop, objs...)
}

// MultTruncate 清除表内容
//
// 会重置 ai，但保留表结构。
func (tx *Tx) MultTruncate(objs ...TableNamer) error {
	return tx.multDo(tx.Truncate, objs...)
}

func (tx *Tx) multDo(f func(TableNamer) error, objs ...TableNamer) error {
	for _, v := range objs {
		if err := f(v); err != nil {
			return err
		}
	}
	return nil
}
