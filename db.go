// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package orm

import (
	"context"
	"database/sql"

	"github.com/issue9/orm/v6/core"
	"github.com/issue9/orm/v6/internal/model"
	"github.com/issue9/orm/v6/sqlbuilder"
)

// DB 数据库操作实例
type DB struct {
	core.Engine
	tablePrefix string
	sqlBuilder  *sqlbuilder.SQLBuilder
	models      *model.Models
	dsn         string
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

	ms, e, err := model.NewModels(db, dialect, tablePrefix)
	if err != nil {
		return nil, err
	}

	return &DB{
		Engine:      e,
		tablePrefix: tablePrefix,
		sqlBuilder:  sqlbuilder.New(e),
		models:      ms,
		dsn:         dsn,
	}, nil
}

// Backup 备份数据库至 dest
//
// 具体格式由各个数据库自行决定。
func (db *DB) Backup(dest string) error { return db.Dialect().Backup(db.dsn, dest) }

// New 重新指定表名前缀为 tablePrefix
//
// 如果要复用表模型，可以采此方法创建一个不同表名前缀的 [DB] 对表模型进行操作。
func (db *DB) New(tablePrefix string) *DB {
	if tablePrefix == db.TablePrefix() {
		return db
	}

	e := db.models.NewEngine(db.DB(), tablePrefix)
	return &DB{
		Engine:      e,
		tablePrefix: tablePrefix,
		sqlBuilder:  sqlbuilder.New(e),
		models:      db.models,
		dsn:         db.dsn,
	}
}

// Close 关闭连接
//
// 同时会清除缓存的模型数据。
// 此操作会让数据库不再可用，包括由 [DB.New] 派生的对象。
func (db *DB) Close() error { return db.models.Close() }

// Version 数据库服务端的版本号
func (db *DB) Version() string { return db.models.Version() }

func (db *DB) LastInsertID(v TableNamer) (int64, error) {
	return db.LastInsertIDContext(context.Background(), v)
}

func (db *DB) LastInsertIDContext(ctx context.Context, v TableNamer) (int64, error) {
	return lastInsertID(ctx, db, v)
}

// Insert 插入数据
//
// NOTE: 若需一次性插入多条数据，请使用 [Tx.InsertMany]。
func (db *DB) Insert(v TableNamer) (sql.Result, error) {
	return db.InsertContext(context.Background(), v)
}

func (db *DB) InsertContext(ctx context.Context, v TableNamer) (sql.Result, error) {
	return insert(ctx, db, v)
}

func (db *DB) Delete(v TableNamer) (sql.Result, error) {
	return db.DeleteContext(context.Background(), v)
}

func (db *DB) DeleteContext(ctx context.Context, v TableNamer) (sql.Result, error) {
	return del(ctx, db, v)
}

func (db *DB) Update(v TableNamer, cols ...string) (sql.Result, error) {
	return db.UpdateContext(context.Background(), v, cols...)
}

func (db *DB) UpdateContext(ctx context.Context, v TableNamer, cols ...string) (sql.Result, error) {
	return update(ctx, db, v, cols...)
}

func (db *DB) Select(v TableNamer) (bool, error) { return db.SelectContext(context.Background(), v) }

func (db *DB) SelectContext(ctx context.Context, v TableNamer) (bool, error) { return find(ctx, db, v) }

func (db *DB) Create(v ...TableNamer) error { return db.CreateContext(context.Background(), v...) }

func (db *DB) CreateContext(ctx context.Context, v ...TableNamer) error {
	if !db.Dialect().TransactionalDDL() {
		for _, t := range v {
			if err := create(ctx, db, t); err != nil {
				return err
			}
		}
		return nil
	}

	return db.DoTransaction(func(tx *Tx) error {
		for _, t := range v {
			if err := create(ctx, tx, t); err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *DB) Drop(v ...TableNamer) error { return db.DropContext(context.Background(), v...) }

func (db *DB) DropContext(ctx context.Context, v ...TableNamer) error {
	if !db.Dialect().TransactionalDDL() {
		for _, t := range v {
			if err := drop(ctx, db, t); err != nil {
				return err
			}
		}
		return nil
	}

	return db.DoTransaction(func(tx *Tx) error {
		for _, t := range v {
			if err := drop(ctx, tx, t); err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *DB) Truncate(v ...TableNamer) error {
	return db.TruncateContext(context.Background(), v...)
}

func (db *DB) TruncateContext(ctx context.Context, v ...TableNamer) error {
	if !db.Dialect().TransactionalDDL() {
		for _, t := range v {
			if err := truncate(ctx, db, t); err != nil {
				return err
			}
		}
		return nil
	}

	return db.DoTransaction(func(tx *Tx) error {
		for _, t := range v {
			if err := truncate(ctx, tx, t); err != nil {
				return err
			}
		}
		return nil
	})
}

// InsertMany 一次插入多条数据
//
// 会自动转换成事务进行处理。
func (db *DB) InsertMany(max int, v ...TableNamer) error {
	return db.InsertManyContext(context.Background(), max, v...)
}

func (db *DB) InsertManyContext(ctx context.Context, max int, v ...TableNamer) error {
	return db.DoTransaction(func(tx *Tx) error { return tx.InsertManyContext(ctx, max, v...) })
}

func (db *DB) SQLBuilder() *sqlbuilder.SQLBuilder { return db.sqlBuilder }

func (db *DB) Ping() error { return db.PingContext(context.Background()) }

func (db *DB) PingContext(ctx context.Context) error { return db.DB().PingContext(ctx) }

func (db *DB) Stats() sql.DBStats { return db.DB().Stats() }

// DB 返回标准库的 [sql.DB] 实例
func (db *DB) DB() *sql.DB { return db.models.DB() }

// TablePrefix 所有数据表拥有的统一表名前缀
//
// 当需要在一个数据库中创建不同的实例，
// 或是同一个数据表结构应用在不同的对象是，可以通过不同的表名前缀对数据表进行区分。
func (db *DB) TablePrefix() string { return db.tablePrefix }
