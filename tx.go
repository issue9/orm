// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package orm

import (
	"context"
	"database/sql"
	"errors"

	"github.com/issue9/orm/v6/core"
	"github.com/issue9/orm/v6/sqlbuilder"
)

// Tx 事务对象
type Tx struct {
	core.Engine
	tx *sql.Tx
	db *DB
}

type txEngine struct {
	core.Engine
	tx *Tx
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
		Engine: db.models.NewEngine(tx, db.TablePrefix()),
		tx:     tx,
		db:     db,
	}, nil
}

func (tx *Tx) LastInsertID(v TableNamer) (int64, error) {
	return tx.LastInsertIDContext(context.Background(), v)
}

func (tx *Tx) LastInsertIDContext(ctx context.Context, v TableNamer) (int64, error) {
	return lastInsertID(ctx, tx, v)
}

func (tx *Tx) Insert(v TableNamer) (sql.Result, error) {
	return tx.InsertContext(context.Background(), v)
}

func (tx *Tx) InsertContext(ctx context.Context, v TableNamer) (sql.Result, error) {
	return insert(ctx, tx, v)
}

func (tx *Tx) Select(v TableNamer) (bool, error) { return tx.SelectContext(context.Background(), v) }

func (tx *Tx) SelectContext(ctx context.Context, v TableNamer) (bool, error) { return find(ctx, tx, v) }

// ForUpdate 读数据并锁定
func (tx *Tx) ForUpdate(v TableNamer) error { return tx.ForUpdateContext(context.Background(), v) }

func (tx *Tx) ForUpdateContext(ctx context.Context, v TableNamer) error { return forUpdate(ctx, tx, v) }

func (tx *Tx) InsertMany(max int, v ...TableNamer) error {
	return tx.InsertManyContext(context.Background(), max, v...)
}

func (tx *Tx) InsertManyContext(ctx context.Context, max int, v ...TableNamer) error {
	return txInsertMany(ctx, tx, max, v...)
}

func (tx *Tx) Update(v TableNamer, cols ...string) (sql.Result, error) {
	return tx.UpdateContext(context.Background(), v, cols...)
}

func (tx *Tx) UpdateContext(ctx context.Context, v TableNamer, cols ...string) (sql.Result, error) {
	return update(ctx, tx, v, cols...)
}

func (tx *Tx) SaveContext(ctx context.Context, v TableNamer, col ...string) (int64, bool, error) {
	return save(ctx, tx, v, col...)
}

func (tx *Tx) Save(v TableNamer, col ...string) (int64, bool, error) {
	return tx.SaveContext(context.Background(), v, col...)
}

func (tx *Tx) Delete(v TableNamer) (sql.Result, error) {
	return tx.DeleteContext(context.Background(), v)
}

func (tx *Tx) DeleteContext(ctx context.Context, v TableNamer) (sql.Result, error) {
	return del(ctx, tx, v)
}

func (tx *Tx) Create(v ...TableNamer) error { return tx.CreateContext(context.Background(), v...) }

func (tx *Tx) CreateContext(ctx context.Context, v ...TableNamer) error {
	for _, t := range v {
		if err := create(ctx, tx, t); err != nil {
			return err
		}
	}
	return nil
}

func (tx *Tx) Drop(v ...TableNamer) error { return tx.DropContext(context.Background(), v...) }

func (tx *Tx) DropContext(ctx context.Context, v ...TableNamer) error {
	for _, t := range v {
		if err := drop(ctx, tx, t); err != nil {
			return err
		}
	}
	return nil
}

func (tx *Tx) Truncate(v ...TableNamer) error { return tx.TruncateContext(context.Background(), v...) }

func (tx *Tx) TruncateContext(ctx context.Context, v ...TableNamer) error {
	for _, t := range v {
		if err := truncate(ctx, tx, t); err != nil {
			return err
		}
	}
	return nil
}

func (tx *Tx) SQLBuilder() *sqlbuilder.SQLBuilder {
	return sqlbuilder.New(tx) // 事务一般是一个临时对象，没必要像 [DB] 一样固定 sqlbuilder 对象。
}

// Commit 提交事务
func (tx *Tx) Commit() error { return tx.Tx().Commit() }

func (tx *Tx) Rollback() error { return tx.Tx().Rollback() }

// Tx 返回标准库的事务接口 [sql.Tx]
func (tx *Tx) Tx() *sql.Tx { return tx.tx }

// NewEngine 为当前事务创建一个不同表名前缀的 [Engine] 对象
//
// 如果要复用表模型，可以采此方法创建一个不同表名前缀的 [Engine] 进行操作表模型。
// 返回对象的生命周期与 [Tx] 相同。
func (tx *Tx) NewEngine(tablePrefix string) Engine {
	if tx.db.TablePrefix() == tablePrefix { // 事务的表名前缀必然是与创建他的 [DB] 是相同的
		return tx
	}

	return &txEngine{
		Engine: tx.db.models.NewEngine(tx.Tx(), tablePrefix),
		tx:     tx,
	}
}

func (e *txEngine) LastInsertID(v TableNamer) (int64, error) {
	return e.LastInsertIDContext(context.Background(), v)
}

func (e *txEngine) LastInsertIDContext(ctx context.Context, v TableNamer) (int64, error) {
	return lastInsertID(ctx, e, v)
}

func (e *txEngine) Insert(v TableNamer) (sql.Result, error) {
	return e.InsertContext(context.Background(), v)
}

func (e *txEngine) InsertContext(ctx context.Context, v TableNamer) (sql.Result, error) {
	return insert(ctx, e, v)
}

func (e *txEngine) Delete(v TableNamer) (sql.Result, error) {
	return e.DeleteContext(context.Background(), v)
}

func (e *txEngine) DeleteContext(ctx context.Context, v TableNamer) (sql.Result, error) {
	return del(ctx, e, v)
}

func (e *txEngine) Update(v TableNamer, cols ...string) (sql.Result, error) {
	return e.UpdateContext(context.Background(), v, cols...)
}

func (e *txEngine) UpdateContext(ctx context.Context, v TableNamer, cols ...string) (sql.Result, error) {
	return update(ctx, e, v, cols...)
}

func (e *txEngine) SaveContext(ctx context.Context, v TableNamer, col ...string) (int64, bool, error) {
	return save(ctx, e, v, col...)
}

func (e *txEngine) Save(v TableNamer, col ...string) (int64, bool, error) {
	return e.SaveContext(context.Background(), v, col...)
}

func (e *txEngine) Select(v TableNamer) (bool, error) {
	return e.SelectContext(context.Background(), v)
}

func (e *txEngine) SelectContext(ctx context.Context, v TableNamer) (bool, error) {
	return find(ctx, e, v)
}

func (e *txEngine) Create(v ...TableNamer) error { return e.CreateContext(context.Background(), v...) }

func (e *txEngine) CreateContext(ctx context.Context, v ...TableNamer) error {
	for _, t := range v {
		if err := create(ctx, e, t); err != nil {
			return err
		}
	}
	return nil
}

func (e *txEngine) Drop(v ...TableNamer) error { return e.DropContext(context.Background(), v...) }

func (e *txEngine) DropContext(ctx context.Context, v ...TableNamer) error {
	for _, t := range v {
		if err := drop(ctx, e, t); err != nil {
			return err
		}
	}
	return nil
}

func (e *txEngine) Truncate(v ...TableNamer) error {
	return e.TruncateContext(context.Background(), v...)
}

func (e *txEngine) TruncateContext(ctx context.Context, v ...TableNamer) error {
	for _, t := range v {
		if err := truncate(ctx, e, t); err != nil {
			return err
		}
	}
	return nil
}

func (e *txEngine) InsertMany(max int, v ...TableNamer) error {
	return e.InsertManyContext(context.Background(), max, v...)
}

func (e *txEngine) InsertManyContext(ctx context.Context, max int, v ...TableNamer) error {
	return txInsertMany(ctx, e, max, v...)
}

func (e *txEngine) SQLBuilder() *sqlbuilder.SQLBuilder {
	return sqlbuilder.New(e) // txPrefix 般是一个临时对象，没必要像 [DB] 一样固定 sqlbuilder 对象。
}

func txInsertMany(ctx context.Context, tx Engine, max int, v ...TableNamer) error {
	l := len(v)
	for i := 0; i < l; i += max {
		j := min(i+max, l)
		query, err := buildInsertManySQL(tx, v[i:j]...)
		if err != nil {
			return err
		}

		if _, err = query.ExecContext(ctx); err != nil {
			return err
		}
	}

	return nil
}

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
