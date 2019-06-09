// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

// Metaer 用于指定一个表级别的元数据。如表名，存储引擎等：
//  "name(tbl_name);engine(myISAM);charset(utf8)"
type Metaer interface {
	Meta() string
}

// ForeignKey 外键
type ForeignKey struct {
	Col                      *Column
	RefTableName, RefColName string
	UpdateRule, DeleteRule   string
}

// 预定的约束类型，方便 Model 中使用。
const (
	Index ConType = iota
	Unique
	Fk
	Check
)

// ConType 约束类型
type ConType int8

func (t ConType) String() string {
	switch t {
	case Index:
		return "KEY INDEX"
	case Unique:
		return "UNIQUE INDEX"
	case Fk:
		return "FOREIGN KEY"
	case Check:
		return "CHECK"
	default:
		return "<unknown>"
	}
}
