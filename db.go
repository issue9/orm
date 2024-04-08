// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package orm

import (
	"context"
	"database/sql"

	"github.com/issue9/orm/v6/core"
	"github.com/issue9/orm/v6/internal/engine"
	"github.com/issue9/orm/v6/internal/model"
	"github.com/issue9/orm/v6/sqlbuilder"
)

// DB 数据库操作实例
type DB struct {
	core.Engine
	db         *sql.DB
	sqlBuilder *sqlbuilder.SQLBuilder
	models     *model.Models
	version    string
}

// NewDB 声明一个新的 [DB] 实例
//
// NOTE: 不同驱动对时间的处理不尽相同，如果有在不同数据库之间移植的需求，
// 那么建议将保存时的时区都统一设置为 UTC：
//   - postgres 已经固定为 UTC；
//   - sqlite3 可以在 dsn 中通过 _loc=UTC 指定；
//   - mysql 默认是 UTC，也可以在 DSN 中通过 loc=UTC 指定；
func NewDB(tablePrefix, dsn string, dialect Dialect) (*DB, error) {
	db, err := sql.Open(dialect.DriverName(), dsn)
	if err != nil {
		return nil, err
	}
	return NewDBWithStdDB(tablePrefix, db, dialect)
}

// NewDBWithStdDB 从 [sql.DB] 构建 [DB] 实例
//
// NOTE: 请确保用于打开 db 的 driverName 参数与 dialect.DriverName() 是相同的，
// 否则后续操作的结果是未知的。
func NewDBWithStdDB(tablePrefix string, db *sql.DB, dialect Dialect) (*DB, error) {
	inst := &DB{
		db:     db,
		Engine: engine.New(db, tablePrefix, dialect),
	}

	inst.models = model.NewModels(inst)
	inst.sqlBuilder = sqlbuilder.New(inst)

	return inst, nil
}

// Close 关闭连接
//
// 同时会清除缓存的模型数据
func (db *DB) Close() error {
	db.models.Clear()
	return db.DB().Close()
}

// Version 数据库服务端的版本号
func (db *DB) Version() (string, error) {
	if db.version == "" {
		ver, err := sqlbuilder.Version(db)
		if err != nil {
			return "", err
		}
		db.version = ver
	}

	return db.version, nil
}

func (db *DB) LastInsertID(v TableNamer) (int64, error) { return lastInsertID(db, v) }

// Insert 插入数据
//
// NOTE: 若需一次性插入多条数据，请使用 [Tx.InsertMany]。
func (db *DB) Insert(v TableNamer) (sql.Result, error) { return insert(db, v) }

func (db *DB) Delete(v TableNamer) (sql.Result, error) { return del(db, v) }

func (db *DB) Update(v TableNamer, cols ...string) (sql.Result, error) {
	return update(db, v, cols...)
}

func (db *DB) Select(v TableNamer) (bool, error) { return find(db, v) }

func (db *DB) Create(v TableNamer) error { return create(db, v) }

func (db *DB) Drop(v TableNamer) error { return drop(db, v) }

func (db *DB) Truncate(v TableNamer) error {
	if !db.Dialect().TransactionalDDL() {
		return truncate(db, v)
	}
	return db.DoTransaction(func(tx *Tx) error { return truncate(tx, v) })
}

// InsertMany 一次插入多条数据
//
// 会自动转换成事务进行处理。
func (db *DB) InsertMany(max int, v ...TableNamer) error {
	return db.DoTransaction(func(tx *Tx) error { return tx.InsertMany(max, v...) })
}

func (db *DB) SQLBuilder() *sqlbuilder.SQLBuilder { return db.sqlBuilder }

func (db *DB) Ping() error { return db.PingContext(context.Background()) }

func (db *DB) PingContext(ctx context.Context) error { return db.DB().PingContext(ctx) }

func (db *DB) Stats() sql.DBStats { return db.DB().Stats() }

// DB 返回标准库的 [sql.DB] 实例
func (db *DB) DB() *sql.DB { return db.db }
