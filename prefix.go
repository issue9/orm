// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package orm

import (
	"database/sql"

	"github.com/issue9/orm/v6/core"
	"github.com/issue9/orm/v6/internal/engine"
	"github.com/issue9/orm/v6/sqlbuilder"
)

type dbPrefix struct {
	core.Engine
	db *DB
}

type txPrefix struct {
	core.Engine
	tx *Tx
}

// Prefix 为所有操作的表名加上统一的前缀
//
// 如果要复用表结构，可以采此对象进行相关操作，而不是直接使用 [DB] 或 [Tx]。
func (db *DB) Prefix(p string) Engine { return newDBPrefix(db, db.TablePrefix()+p, db.Dialect()) }

func (p *dbPrefix) Prefix(pp string) Engine {
	return newDBPrefix(p.db, p.TablePrefix()+pp, p.Dialect())
}

func newDBPrefix(db *DB, tablePrefix string, d Dialect) Engine {
	return &dbPrefix{
		Engine: engine.New(db.DB(), tablePrefix, d),
		db:     db,
	}
}

// Prefix 为所有操作的表名加上统一的前缀
//
// 如果要复用表结构，可以采此对象进行相关操作，而不是直接使用 [DB] 或 [Tx]。
//
// 创建的 [Engine] 依然属于当前事务。
func (tx *Tx) Prefix(p string) Engine { return newTxPrefix(tx, tx.TablePrefix()+p, tx.Dialect()) }

func (p *txPrefix) Prefix(pp string) Engine {
	return newTxPrefix(p.tx, p.TablePrefix()+pp, p.Dialect())
}

func newTxPrefix(tx *Tx, tablePrefix string, d Dialect) Engine {
	return &txPrefix{
		Engine: engine.New(tx.Tx(), tablePrefix, d),
		tx:     tx,
	}
}

func (p *dbPrefix) LastInsertID(v TableNamer) (int64, error) { return lastInsertID(p, v) }

func (p *dbPrefix) Insert(v TableNamer) (sql.Result, error) { return insert(p, v) }

func (p *dbPrefix) Delete(v TableNamer) (sql.Result, error) { return del(p, v) }

func (p *dbPrefix) Update(v TableNamer, cols ...string) (sql.Result, error) {
	return update(p, v, cols...)
}

func (p *dbPrefix) Select(v TableNamer) (bool, error) { return find(p, v) }

func (p *dbPrefix) Create(v TableNamer) error { return create(p, v) }

func (p *dbPrefix) Drop(v TableNamer) error { return drop(p, v) }

func (p *dbPrefix) Truncate(v TableNamer) error { return truncate(p, v) }

func (p *dbPrefix) InsertMany(max int, v ...TableNamer) error {
	return p.db.DoTransaction(func(tx *Tx) error {
		return tx.Prefix(p.TablePrefix()).InsertMany(max, v...)
	})
}

func (p *dbPrefix) SQLBuilder() *sqlbuilder.SQLBuilder {
	return sqlbuilder.New(p) // dbPrefix 般是一个临时对象，没必要像 [DB] 一样固定 sqlbuilder 对象。
}

func (p *txPrefix) LastInsertID(v TableNamer) (int64, error) { return lastInsertID(p, v) }

func (p *txPrefix) Insert(v TableNamer) (sql.Result, error) { return insert(p, v) }

func (p *txPrefix) Delete(v TableNamer) (sql.Result, error) { return del(p, v) }

func (p *txPrefix) Update(v TableNamer, cols ...string) (sql.Result, error) {
	return update(p, v, cols...)
}

func (p *txPrefix) Select(v TableNamer) (bool, error) { return find(p, v) }

func (p *txPrefix) Create(v TableNamer) error { return create(p, v) }

func (p *txPrefix) Drop(v TableNamer) error { return drop(p, v) }

func (p *txPrefix) Truncate(v TableNamer) error { return truncate(p, v) }

func (p *txPrefix) InsertMany(max int, v ...TableNamer) error {
	l := len(v)
	for i := 0; i < l; i += max {
		j := min(i+max, l)
		query, err := buildInsertManySQL(p, v[i:j]...)
		if err != nil {
			return err
		}

		if _, err = query.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (p *txPrefix) SQLBuilder() *sqlbuilder.SQLBuilder {
	return sqlbuilder.New(p) // txPrefix 般是一个临时对象，没必要像 [DB] 一样固定 sqlbuilder 对象。
}
