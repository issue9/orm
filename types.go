// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package orm

import (
	"database/sql"
	"time"

	"github.com/issue9/orm/v6/core"
	"github.com/issue9/orm/v6/fetch"
	"github.com/issue9/orm/v6/sqlbuilder"
	"github.com/issue9/orm/v6/types"
)

type (
	// Unix 表示 Unix 时间戳的数据样式
	//
	// 表现为 [time.Time]，但是保存数据库时，以 unix 时间戳的形式保存。
	Unix = types.Unix

	Rat = types.Rat

	Decimal = types.Decimal

	// AfterFetcher 从数据库查询到数据之后需要执行的操作
	AfterFetcher = fetch.AfterFetcher

	// Dialect 数据库驱动特有的语言特性实现
	Dialect = core.Dialect

	// BeforeUpdater 在更新之前调用的函数
	BeforeUpdater interface {
		BeforeUpdate() error
	}

	// BeforeInserter 在插入之前调用的函数
	BeforeInserter interface {
		BeforeInsert() error
	}

	// Engine 数据操作引擎
	//
	// 相对于 [core.Engine]，添加了针对 [TableNamer] 的操作。
	// 所有针对 [TableNamer] 的操作与 sqlbuilder 的拼接方式有以下区别：
	//  - 针对 [TableNamer] 的操作会自动为表名加上 # 表名前缀；
	//  - 针对 [TableNamer] 的操作会为约束名加上表名，以确保约束名的唯一性；
	Engine interface {
		core.Engine

		// LastInsertID 插入一条数据并返回其自增 ID
		//
		// 理论上功能等同于以下两步操作：
		//  rslt, err := engine.Insert(obj)
		//  id, err := rslt.LastInsertId()
		// 但是实际上部分数据库不支持直接在 sql.Result 中获取 LastInsertId，
		// 比如 postgresql，所以使用 LastInsertID() 会是比 sql.Result
		// 更简单和安全的方法。
		//
		// NOTE: 要求 v 有定义自增列。
		LastInsertID(v TableNamer) (int64, error)

		Insert(v TableNamer) (sql.Result, error)

		// Delete 删除符合条件的数据
		//
		// 查找条件以结构体定义的主键或是唯一约束(在没有主键的情况下)来查找，
		// 若两者都不存在，则将返回 error
		Delete(v TableNamer) (sql.Result, error)

		// Update 更新数据
		//
		// 零值不会被提交，cols 指定的列，即使是零值也会被更新。
		//
		// 查找条件以结构体定义的主键或是唯一约束(在没有主键的情况下)来查找，
		// 若两者都不存在，则将返回 error
		Update(v TableNamer, cols ...string) (sql.Result, error)

		// Select 查询一个符合条件的数据
		//
		// 查找条件以结构体定义的主键或是唯一约束(在没有主键的情况下 ) 来查找，
		// 若两者都不存在，则将返回 error
		// 若没有符合条件的数据，将不会对参数 v 做任何变动。
		//
		// 查找条件的查找顺序是为 自增 > 主键 > 唯一约束，
		// 如果同时存在多个唯一约束满足条件(可能每个唯一约束查询至的结果是不一样的)，则返回错误信息。
		Select(v TableNamer) (found bool, err error)

		Create(v TableNamer) error

		Drop(v TableNamer) error

		// Truncate 清空表并重置 ai 但保留表结构
		Truncate(v TableNamer) error

		// InsertMany 插入多条相同的数据
		//
		// 若需要向某张表中插入多条记录，InsertMany() 会比 Insert() 性能上好很多。
		//
		// max 表示一次最多插入的数量，如果超过此值，会分批执行，
		// 但是依然在一个事务中完成。
		InsertMany(max int, v ...TableNamer) error

		// Where 生成 [WhereStmt] 语句
		Where(cond string, args ...any) *WhereStmt

		SQLBuilder() *sqlbuilder.SQLBuilder

		// Prefix 声明带有统一表名前缀的 [Engine]
		//
		// 返回的实例表名前缀为当前实例的表名前缀+p
		Prefix(p string) Engine

		// newModel 获取 v 的 [core.Model] 实例
		//
		// 内部使用不公开，[Engine] 也不会有外部的实现。
		newModel(v TableNamer) (*core.Model, error)
	}
)

// NowUnix 返回当前时间
func NowUnix() Unix { return Unix{Time: time.Now()} }

// NowNullTime 返回当前时间
func NowNullTime() sql.NullTime { return sql.NullTime{Time: time.Now(), Valid: true} }
