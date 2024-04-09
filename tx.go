// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package orm

import (
	"context"
	"database/sql"
	"errors"

	"github.com/issue9/orm/v6/core"
	"github.com/issue9/orm/v6/internal/engine"
	"github.com/issue9/orm/v6/sqlbuilder"
)

// Tx 事务对象
type Tx struct {
	core.Engine
	tx *sql.Tx
	db *DB
}

// Begin 开始一个新的事务
func (db *DB) Begin() (*Tx, error) { return db.BeginTx(context.Background(), nil) }

// BeginTx 开始一个新的事务
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.DB().BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{
		tx:     tx,
		db:     db,
		Engine: engine.New(tx, db.TablePrefix(), db.Dialect()),
	}, nil
}

func (tx *Tx) LastInsertID(v TableNamer) (int64, error) { return lastInsertID(tx, v) }

func (tx *Tx) Insert(v TableNamer) (sql.Result, error) { return insert(tx, v) }

func (tx *Tx) Select(v TableNamer) (bool, error) { return find(tx, v) }

// ForUpdate 读数据并锁定
func (tx *Tx) ForUpdate(v TableNamer) error { return forUpdate(tx, v) }

func (tx *Tx) InsertMany(max int, v ...TableNamer) error {
	l := len(v)
	for i := 0; i < l; i += max {
		j := min(i+max, l)
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

func (tx *Tx) Update(v TableNamer, cols ...string) (sql.Result, error) { return update(tx, v, cols...) }

func (tx *Tx) Delete(v TableNamer) (sql.Result, error) { return del(tx, v) }

func (tx *Tx) Create(v TableNamer) error { return create(tx, v) }

func (tx *Tx) Drop(v TableNamer) error { return drop(tx, v) }

func (tx *Tx) Truncate(v TableNamer) error { return truncate(tx, v) }

func (tx *Tx) SQLBuilder() *sqlbuilder.SQLBuilder {
	return sqlbuilder.New(tx) // 事务一般是一个临时对象，没必要像 [DB] 一样固定 sqlbuilder 对象。
}

// Commit 提交事务
func (tx *Tx) Commit() error { return tx.Tx().Commit() }

func (tx *Tx) Rollback() error { return tx.Tx().Rollback() }

// Tx 返回标准库的事务接口 [sql.Tx]
func (tx *Tx) Tx() *sql.Tx { return tx.tx }

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
		return errors.Join(err, tx.Rollback())
	}

	return tx.Commit()
}
