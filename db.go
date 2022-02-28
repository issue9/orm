// SPDX-License-Identifier: MIT

package orm

import (
	"context"
	"database/sql"
	"log"

	"github.com/issue9/orm/v5/core"
	"github.com/issue9/orm/v5/internal/model"
	"github.com/issue9/orm/v5/sqlbuilder"
)

// DB 数据库操作实例
type DB struct {
	*sql.DB
	dialect    Dialect
	sqlBuilder *sqlbuilder.SQLBuilder
	models     *model.Models
	version    string

	sqlLogger *log.Logger
}

// NewDB 声明一个新的 DB 实例
//
// 不同驱动对时间的处理不尽相同，如果有在不同数据库之间移植的需求，
// 那么建议将保存时的时区都统一设置为 UTC：
// postgres 已经固定为 UTC，sqlite3 可以在 dsn 中通过 _loc=UTC 指定，
// mysql 默认是 UTC，也可以在 DSN 中通过 loc=UTC 指定。
func NewDB(dsn string, dialect Dialect) (*DB, error) {
	db, err := sql.Open(dialect.DriverName(), dsn)
	if err != nil {
		return nil, err
	}

	return NewDBWithStdDB(db, dialect)
}

// NewDBWithStdDB 从 sql.DB 构建 DB 实例
//
// NOTE: 请确保用于打开 db 的 driverName 参数与 dialect.DriverName() 是相同的，
// 否则后续操作的结果是未知的。
func NewDBWithStdDB(db *sql.DB, dialect Dialect) (*DB, error) {
	inst := &DB{
		DB:      db,
		dialect: dialect,
	}

	inst.models = model.NewModels(inst)
	inst.sqlBuilder = sqlbuilder.New(inst)

	return inst, nil
}

func (db *DB) TablePrefix() string { return db.Dialect().TablePrefix() }

// Debug 指定调输出调试内容通道
//
// 如果 l 不为 nil，则每次 SQL 调用都会输出 SQL 语句，
// 预编译的语句，仅在预编译时输出；
// 如果为 nil，则表示关闭调试。
func (db *DB) Debug(l *log.Logger) { db.sqlLogger = l }

func (db *DB) Dialect() Dialect { return db.dialect }

// Close 关闭连接
//
// 同时会清除缓存的模型数据
func (db *DB) Close() error {
	db.models.Clear()
	return db.DB.Close()
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

// QueryRow 执行一条查询语句
//
// 如果生成语句出错，则会 panic
func (db *DB) QueryRow(query string, args ...any) *sql.Row {
	return db.QueryRowContext(context.Background(), query, args...)
}

func (db *DB) printDebug(query string) {
	if db.sqlLogger != nil {
		db.sqlLogger.Println(query)
	}
}

// QueryRowContext 执行一条查询语句
//
// 如果生成语句出错，则会 panic
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	db.printDebug(query)
	query, args, err := db.dialect.Fix(query, args)
	if err != nil {
		panic(err)
	}

	return db.DB.QueryRowContext(ctx, query, args...)
}

func (db *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return db.QueryContext(context.Background(), query, args...)
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	db.printDebug(query)
	query, args, err := db.dialect.Fix(query, args)
	if err != nil {
		return nil, err
	}

	return db.DB.QueryContext(ctx, query, args...)
}

func (db *DB) Exec(query string, args ...any) (sql.Result, error) {
	return db.ExecContext(context.Background(), query, args...)
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	db.printDebug(query)
	query, args, err := db.dialect.Fix(query, args)
	if err != nil {
		return nil, err
	}

	return db.DB.ExecContext(ctx, query, args...)
}

func (db *DB) Prepare(query string) (*core.Stmt, error) {
	return db.PrepareContext(context.Background(), query)
}

func (db *DB) PrepareContext(ctx context.Context, query string) (*core.Stmt, error) {
	db.printDebug(query)
	query, orders, err := db.Dialect().Prepare(query)
	if err != nil {
		return nil, err
	}

	s, err := db.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return core.NewStmt(s, orders), nil
}

func (db *DB) LastInsertID(v TableNamer) (int64, error) { return lastInsertID(db, v) }

// Insert 插入数据
//
// NOTE: 若需一次性插入多条数据，请使用 tx.InsertMany()。
func (db *DB) Insert(v TableNamer) (sql.Result, error) { return insert(db, v) }

func (db *DB) Delete(v TableNamer) (sql.Result, error) { return del(db, v) }

func (db *DB) Update(v TableNamer, cols ...string) (sql.Result, error) {
	return update(db, v, cols...)
}

func (db *DB) Select(v TableNamer) error { return find(db, v) }

func (db *DB) Create(v TableNamer) error { return create(db, v) }

func (db *DB) Drop(v TableNamer) error { return drop(db, v) }

func (db *DB) Truncate(v TableNamer) error {
	if !db.Dialect().TransactionalDDL() {
		return truncate(db, v)
	}

	return db.DoTransaction(func(tx *Tx) error {
		return truncate(tx, v)
	})
}

// InsertMany 一次插入多条数据
//
// 会自动转换成事务进行处理。
func (db *DB) InsertMany(max int, v ...TableNamer) error {
	return db.DoTransaction(func(tx *Tx) error {
		return tx.InsertMany(max, v...)
	})
}

// MultInsert 插入一个或多个数据
//
// 会自动转换成事务进行处理。
func (db *DB) MultInsert(objs ...TableNamer) error {
	return db.DoTransaction(func(tx *Tx) error {
		return tx.MultInsert(objs...)
	})
}

// MultSelect 选择符合要求的一条或是多条记录
//
// 会自动转换成事务进行处理。
func (db *DB) MultSelect(objs ...TableNamer) error {
	return db.DoTransaction(func(tx *Tx) error {
		return tx.MultSelect(objs...)
	})
}

// MultUpdate 更新一条或多条类型。
//
// 会自动转换成事务进行处理。
func (db *DB) MultUpdate(objs ...TableNamer) error {
	return db.DoTransaction(func(tx *Tx) error {
		return tx.MultUpdate(objs...)
	})
}

// MultDelete 删除一条或是多条数据
//
// 会自动转换成事务进行处理。
func (db *DB) MultDelete(objs ...TableNamer) error {
	return db.DoTransaction(func(tx *Tx) error {
		return tx.MultDelete(objs...)
	})
}

// MultCreate 创建数据表
//
// 如果数据库支持事务 DDL，则会在事务中完成此操作。
func (db *DB) MultCreate(objs ...TableNamer) error {
	if !db.Dialect().TransactionalDDL() {
		for _, v := range objs {
			if err := db.Create(v); err != nil {
				return err
			}
		}
		return nil
	}

	return db.DoTransaction(func(tx *Tx) error {
		return tx.MultCreate(objs...)
	})
}

// MultDrop 删除表结构及数据
//
// 如果数据库支持事务 DDL，则会在事务中完成此操作。
func (db *DB) MultDrop(objs ...TableNamer) error {
	if !db.Dialect().TransactionalDDL() {
		for _, v := range objs {
			if err := db.Drop(v); err != nil {
				return err
			}
		}
		return nil
	}

	return db.DoTransaction(func(tx *Tx) error {
		return tx.MultDrop(objs...)
	})
}

func (db *DB) MultTruncate(objs ...TableNamer) error {
	if !db.Dialect().TransactionalDDL() {
		for _, v := range objs {
			if err := truncate(db, v); err != nil {
				return err
			}
		}
		return nil
	}

	return db.DoTransaction(func(tx *Tx) error {
		return tx.MultTruncate(objs...)
	})
}
