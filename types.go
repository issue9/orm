// SPDX-License-Identifier: MIT

package orm

import (
	"database/sql"

	"github.com/issue9/orm/v4/core"
	"github.com/issue9/orm/v4/fetch"
	"github.com/issue9/orm/v4/sqlbuilder"
	"github.com/issue9/orm/v4/types"
)

type (
	// Unix 表示 Unix 时间戳的数据样式
	//
	// 表现为 time.Time，但是保存数据库时，以 unix 时间戳的形式保存。
	Unix = types.Unix

	Rat = types.Rat

	// AfterFetcher 从数据库查询到数据之后需要执行的操作
	AfterFetcher = fetch.AfterFetcher

	// Column 列结构
	Column = core.Column

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

	// Engine 是 DB 与 Tx 的共有接口
	Engine interface {
		core.Engine

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

		Delete(v TableNamer) (sql.Result, error)

		Update(v TableNamer, cols ...string) (sql.Result, error)

		Select(v TableNamer) error

		Create(v TableNamer) error

		Drop(v TableNamer) error

		Truncate(v TableNamer) error

		// InsertMany 插入多条相同的数据
		//
		// 若需要向某张表中插入多条记录，InsertMany() 会比 Insert() 性能上好很多。
		//
		// max 表示一次最多插入的数量，如果超过此值，会分批执行，
		// 但是依然在一个事务中完成。
		//
		// 与 MultInsert() 方法最大的不同在于:
		//  // MultInsert() 可以每个参数的类型都不一样：
		//  vs := []interface{}{&user{...}, &userInfo{...}}
		//  db.Insert(vs...)
		//  // db.InsertMany(500, vs) // 这里将出错，数组的元素的类型必须相同。
		//  us := []*users{&user{}, &user{}}
		//  db.InsertMany(500, us)
		//  db.Insert(us...) // 这样也行，但是性能会差好多
		InsertMany(max int, v ...TableNamer) error

		// MultInsert 一次性插入多条不同的数据
		//
		// 本质上是对 Insert 的多次调用，合并在一个事务中完成。
		MultInsert(objs ...TableNamer) error

		MultSelect(objs ...TableNamer) error

		MultUpdate(objs ...TableNamer) error

		MultDelete(objs ...TableNamer) error

		MultCreate(objs ...TableNamer) error

		MultDrop(objs ...TableNamer) error

		MultTruncate(objs ...TableNamer) error

		SQLBuilder() *sqlbuilder.SQLBuilder

		NewModel(v TableNamer) (*Model, error)
	}
)
