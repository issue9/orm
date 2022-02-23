// SPDX-License-Identifier: MIT

package orm

import (
	"database/sql"
	"time"

	"github.com/issue9/orm/v4/core"
)

type (
	ApplyModeler = core.ApplyModeler

	// Model 表示一个数据库的表模型
	Model = core.Model

	TableNamer = core.TableNamer

	// Table 表的基本字段
	//
	// 可嵌入到其它表中。
	Table struct {
		ID      int64        `orm:"name(id);ai"`
		Created time.Time    `orm:"name(created)"`
		Updated time.Time    `orm:"name(updated)"`
		Deleted sql.NullTime `orm:"name(deleted);nullable"`
	}
)

// NewModel 从一个 obj 声明一个 Model 实例
//
// obj 可以是一个 struct 实例或是指针。
func (db *DB) NewModel(obj TableNamer) (*Model, error) { return db.models.New(obj) }

// NewModel 从一个 obj 声明一个 Model 实例
//
// obj 可以是一个 struct 实例或是指针。
func (tx *Tx) NewModel(obj TableNamer) (*Model, error) { return tx.db.NewModel(obj) }

func (t *Table) BeforeUpdate() error {
	t.Updated = time.Now()
	return nil
}

func (t *Table) BeforeCreate() error {
	t.Created = time.Now()
	t.Updated = t.Created
	return nil
}
