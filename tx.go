// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package orm

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/issue9/orm/v5/core"
	"github.com/issue9/orm/v5/sqlbuilder"
)

// Tx 事务对象
type Tx struct {
	*sql.Tx
	db         *DB
	sqlBuilder *sqlbuilder.SQLBuilder
}

// Begin 开始一个新的事务
func (db *DB) Begin() (*Tx, error) { return db.BeginTx(context.Background(), nil) }

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

func (tx *Tx) Query(query string, args ...any) (*sql.Rows, error) {
	return tx.QueryContext(context.Background(), query, args...)
}

func (tx *Tx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	tx.db.sqlLogger(query)
	query, args, err := tx.Dialect().Fix(query, args)
	if err != nil {
		return nil, err
	}

	query = tx.db.replacer.Replace(query)
	return tx.Tx.QueryContext(ctx, query, args...)
}

// QueryRow 执行一条查询语句
//
// 如果生成语句出错，则会 panic
func (tx *Tx) QueryRow(query string, args ...any) *sql.Row {
	return tx.QueryRowContext(context.Background(), query, args...)
}

// QueryRowContext 执行一条查询语句
//
// 如果生成语句出错，则会 panic
func (tx *Tx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	tx.db.sqlLogger(query)
	query, args, err := tx.Dialect().Fix(query, args)
	if err != nil {
		panic(err)
	}

	query = tx.db.replacer.Replace(query)
	return tx.Tx.QueryRowContext(ctx, query, args...)
}

func (tx *Tx) Exec(query string, args ...any) (sql.Result, error) {
	return tx.ExecContext(context.Background(), query, args...)
}

func (tx *Tx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	tx.db.sqlLogger(query)
	query, args, err := tx.Dialect().Fix(query, args)
	if err != nil {
		return nil, err
	}

	query = tx.db.replacer.Replace(query)
	return tx.Tx.ExecContext(ctx, query, args...)
}

func (tx *Tx) Prepare(query string) (*core.Stmt, error) {
	return tx.PrepareContext(context.Background(), query)
}

func (tx *Tx) PrepareContext(ctx context.Context, query string) (*core.Stmt, error) {
	tx.db.sqlLogger(query)
	query, orders, err := tx.Dialect().Prepare(query)
	if err != nil {
		return nil, err
	}

	query = tx.db.replacer.Replace(query)
	s, err := tx.Tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return core.NewStmt(s, orders), nil
}

func (tx *Tx) Dialect() Dialect { return tx.db.Dialect() }

func (tx *Tx) LastInsertID(v TableNamer) (int64, error) {
	return lastInsertID(tx, v)
}

func (tx *Tx) Insert(v TableNamer) (sql.Result, error) { return insert(tx, v) }

func (tx *Tx) Select(v TableNamer) (bool, error) { return find(tx, v) }

// ForUpdate 读数据并锁定
func (tx *Tx) ForUpdate(v TableNamer) error { return forUpdate(tx, v) }

func (tx *Tx) InsertMany(max int, v ...TableNamer) error {
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

func (tx *Tx) Update(v TableNamer, cols ...string) (sql.Result, error) {
	return update(tx, v, cols...)
}

func (tx *Tx) Delete(v TableNamer) (sql.Result, error) { return del(tx, v) }

func (tx *Tx) Create(v TableNamer) error { return create(tx, v) }

func (tx *Tx) Drop(v TableNamer) error { return drop(tx, v) }

func (tx *Tx) Truncate(v TableNamer) error { return truncate(tx, v) }

func (tx *Tx) SQLBuilder() *sqlbuilder.SQLBuilder { return tx.sqlBuilder }

// DoTransaction 将 f 中的内容以事务的方式执行
func (db *DB) DoTransaction(f func(tx *Tx) error) error {
	return db.DoTransactionTx(context.Background(), nil, f)
}

// DoTransactionTx 将 f 中的内容以事务的方式执行
//
// 如果执行失败，自动回滚，且返回错误信息。否则会直接提交。
func (db *DB) DoTransactionTx(ctx context.Context, opt *sql.TxOptions, f func(tx *Tx) error) error {
	tx, err := db.BeginTx(ctx, opt)
	if err != nil {
		return err
	}

	if err := f(tx); err != nil {
		if err1 := tx.Rollback(); err1 != nil {
			return fmt.Errorf("在抛出错误 %w 时再次发生错误 %s", err, err1.Error())
		}
		return err
	}

	return tx.Commit()
}
