// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package orm

import (
	"database/sql"

	"github.com/issue9/orm/v6/sqlbuilder"
)

type dbPrefix struct {
	*DB
	p  string
	sb *sqlbuilder.SQLBuilder
}

type txPrefix struct {
	*Tx
	p  string
	sb *sqlbuilder.SQLBuilder
}

// Prefix 为所有操作的表名加上统一的前缀
//
// 如果要复用表结构，可以采此对象进行相关操作，而不是直接使用 DB 或 Tx。
func (db *DB) Prefix(p string) Engine {
	dp := &dbPrefix{
		p:  db.TablePrefix() + p,
		DB: db,
	}
	dp.sb = sqlbuilder.New(dp)

	return dp
}

// Prefix 为所有操作的表名加上统一的前缀
//
// 如果要复用表结构，可以采此对象进行相关操作，而不是直接使用 [DB] 或 [Tx]。
//
// 创建的 [Engine] 依然属于当前事务。
func (tx *Tx) Prefix(p string) Engine {
	dp := &txPrefix{
		p:  tx.TablePrefix() + p,
		Tx: tx,
	}
	dp.sb = sqlbuilder.New(dp)

	return dp
}

func (p *dbPrefix) TablePrefix() string { return p.p }

func (p *dbPrefix) LastInsertID(v TableNamer) (int64, error) { return lastInsertID(p, v) }

// Insert 插入数据
//
// NOTE: 若需一次性插入多条数据，请使用 tx.InsertMany()。
func (p *dbPrefix) Insert(v TableNamer) (sql.Result, error) { return insert(p, v) }

func (p *dbPrefix) Delete(v TableNamer) (sql.Result, error) { return del(p, v) }

func (p *dbPrefix) Update(v TableNamer, cols ...string) (sql.Result, error) {
	return update(p, v, cols...)
}

func (p *dbPrefix) Select(v TableNamer) (bool, error) { return find(p, v) }

func (p *dbPrefix) Create(v TableNamer) error { return create(p, v) }

func (p *dbPrefix) Drop(v TableNamer) error { return drop(p, v) }

func (p *dbPrefix) Truncate(v TableNamer) error { return truncate(p, v) }

func (p *dbPrefix) TableName(v TableNamer) string { return p.TablePrefix() + v.TableName() }

func (p *dbPrefix) InsertMany(max int, v ...TableNamer) error {
	return p.DB.DoTransaction(func(tx *Tx) error {
		return tx.Prefix(p.p).InsertMany(max, v...)
	})
}

func (p *dbPrefix) SQLBuilder() *sqlbuilder.SQLBuilder { return p.sb }

func (p *txPrefix) TablePrefix() string { return p.p }

func (p *txPrefix) LastInsertID(v TableNamer) (int64, error) { return lastInsertID(p, v) }

// Insert 插入数据
//
// NOTE: 若需一次性插入多条数据，请使用 InsertMany。
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
		j := i + max
		if j > l {
			j = l
		}
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

func (p *txPrefix) SQLBuilder() *sqlbuilder.SQLBuilder { return p.sb }

func (p *txPrefix) TableName(v TableNamer) string { return p.TablePrefix() + v.TableName() }
