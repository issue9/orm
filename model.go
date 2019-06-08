// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import "github.com/issue9/orm/v2/model"

type (
	// Metaer 用于指定一个表级别的元数据。如表名，存储引擎等：
	//  "name(tbl_name);engine(myISAM);charset(utf8)"
	Metaer = model.Metaer

	// Model 表示一个数据库的表模型。数据结构从字段和字段的 struct tag 中分析得出。
	Model = model.Model

	// Column 列结构
	Column = model.Column

	// ForeignKey 外键
	ForeignKey = model.ForeignKey
)

// NewModel 从一个 obj 声明一个 Model 实例。
// obj 可以是一个 struct 实例或是指针。
func (db *DB) NewModel(obj interface{}) (*Model, error) {
	return db.models.New(obj)
}

// NewModel 从一个 obj 声明一个 Model 实例。
// obj 可以是一个 struct 实例或是指针。
func (tx *Tx) NewModel(obj interface{}) (*Model, error) {
	return tx.db.models.New(obj)
}
