// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package orm

import (
	"context"
	"database/sql"
	"strings"

	"github.com/issue9/orm/v6/core"
	"github.com/issue9/orm/v6/internal/model"
	"github.com/issue9/orm/v6/sqlbuilder"
)

// DB 数据库操作实例
type DB struct {
	*sql.DB
	tablePrefix string
	dialect     Dialect
	sqlBuilder  *sqlbuilder.SQLBuilder
	models      *model.Models
	version     string
	replacer    *strings.Replacer

	sqlLogger func(string)
}

func defaultSQLLogger(string) {}

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
func NewDBWithStdDB(tablePRefix string, db *sql.DB, dialect Dialect) (*DB, error) {
	l, r := dialect.Quotes()
	inst := &DB{
		DB:          db,
		tablePrefix: tablePRefix,
		dialect:     dialect,
		replacer: strings.NewReplacer(
			string(core.QuoteLeft), string(l),
			string(core.QuoteRight), string(r),
		),

		sqlLogger: defaultSQLLogger,
	}

	inst.models = model.NewModels(inst)
	inst.sqlBuilder = sqlbuilder.New(inst)

	return inst, nil
}

func (db *DB) TablePrefix() string { return db.tablePrefix }

// Debug 指定调输出调试内容通道
//
// 如果 l 不为 nil，则每次 SQL 调用都会输出 SQL 语句，预编译的语句，仅在预编译时输出；
// 如果为 nil，则表示关闭调试。
func (db *DB) Debug(l func(string)) {
	if l == nil {
		l = defaultSQLLogger
	}
	db.sqlLogger = l
}

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

// QueryRowContext 执行一条查询语句
//
// 如果生成语句出错，则会 panic
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	db.sqlLogger(query)
	query, args, err := db.dialect.Fix(query, args)
	if err != nil {
		panic(err)
	}

	query = db.replacer.Replace(query)
	return db.DB.QueryRowContext(ctx, query, args...)
}

func (db *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return db.QueryContext(context.Background(), query, args...)
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	db.sqlLogger(query)
	query, args, err := db.dialect.Fix(query, args)
	if err != nil {
		return nil, err
	}

	query = db.replacer.Replace(query)
	return db.DB.QueryContext(ctx, query, args...)
}

func (db *DB) Exec(query string, args ...any) (sql.Result, error) {
	return db.ExecContext(context.Background(), query, args...)
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	db.sqlLogger(query)
	query, args, err := db.dialect.Fix(query, args)
	if err != nil {
		return nil, err
	}

	query = db.replacer.Replace(query)
	return db.DB.ExecContext(ctx, query, args...)
}

func (db *DB) Prepare(query string) (*core.Stmt, error) {
	return db.PrepareContext(context.Background(), query)
}

func (db *DB) PrepareContext(ctx context.Context, query string) (*core.Stmt, error) {
	db.sqlLogger(query)
	query, orders, err := db.Dialect().Prepare(query)
	if err != nil {
		return nil, err
	}

	query = db.replacer.Replace(query)
	s, err := db.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return core.NewStmt(s, orders), nil
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

func (db *DB) TableName(v TableNamer) string { return db.TablePrefix() + v.TableName() }
