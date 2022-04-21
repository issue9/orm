// SPDX-License-Identifier: MIT

package orm

import (
	"database/sql"

	"github.com/issue9/orm/v5/core"
	"github.com/issue9/orm/v5/internal/model"
	"github.com/issue9/orm/v5/sqlbuilder"
)

// Prefix 为所有操作的表名加上统一的前缀
//
// 如果要复用表结构，可以采此对象进行相关操作，而不是直接使用 DB 或 Tx。
type Prefix struct {
	p    string
	ms   *model.Models
	e    core.Engine
	sb   *sqlbuilder.SQLBuilder
	isTx bool
}

func (db *DB) Prefix(p string) *Prefix {
	return &Prefix{
		p:    p,
		ms:   db.models,
		e:    db,
		sb:   db.SQLBuilder(),
		isTx: false,
	}
}

func (tx *Tx) Prefix(p string) *Prefix {
	return &Prefix{
		p:    p,
		ms:   tx.db.models,
		e:    tx,
		sb:   tx.SQLBuilder(),
		isTx: true,
	}
}

func (p *Prefix) LastInsertID(v TableNamer) (int64, error) { return lastInsertID(p, v) }

// Insert 插入数据
//
// NOTE: 若需一次性插入多条数据，请使用 tx.InsertMany()。
func (p *Prefix) Insert(v TableNamer) (sql.Result, error) { return insert(p, v) }

func (p *Prefix) Delete(v TableNamer) (sql.Result, error) { return del(p, v) }

func (p *Prefix) Update(v TableNamer, cols ...string) (sql.Result, error) {
	return update(p, v, cols...)
}

func (p *Prefix) Select(v TableNamer) error { return find(p, v) }

func (p *Prefix) Create(v TableNamer) error { return create(p, v) }

func (p *Prefix) Drop(v TableNamer) error { return drop(p, v) }

func (p *Prefix) Truncate(v TableNamer) error { return truncate(p, v) }

func (p *Prefix) InsertMany(max int, v ...TableNamer) error {
	// TODO 非事务模式下，创建事务再执行？

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

func (p *Prefix) SQLBuilder() *sqlbuilder.SQLBuilder { return p.sb }
