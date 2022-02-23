// SPDX-License-Identifier: MIT

package orm

import "github.com/issue9/orm/v4/core"

type (
	ApplyModeler = core.ApplyModeler

	// Model 表示一个数据库的表模型
	Model = core.Model

	// ForeignKey 外键
	ForeignKey = core.ForeignKey

	TableNamer = core.TableNamer
)

// NewModel 从一个 obj 声明一个 Model 实例
//
// obj 可以是一个 struct 实例或是指针。
func (db *DB) NewModel(obj TableNamer) (*Model, error) { return db.models.New(obj) }

// NewModel 从一个 obj 声明一个 Model 实例
//
// obj 可以是一个 struct 实例或是指针。
func (tx *Tx) NewModel(obj TableNamer) (*Model, error) { return tx.db.NewModel(obj) }
