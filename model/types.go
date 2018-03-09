// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

// 预定的约束类型，方便 Model 中使用。
const (
	none conType = iota
	index
	unique
	fk
	check
)

type conType int8

// Metaer 用于指定一个表级别的元数据。如表名，存储引擎等：
//  "name(tbl_name);engine(myISAM);charset(utf-8)"
type Metaer interface {
	Meta() string
}

// ForeignKey 外键
type ForeignKey struct {
	Col                      *Column
	RefTableName, RefColName string
	UpdateRule, DeleteRule   string
}

func (t conType) String() string {
	switch t {
	case none:
		return "<none>"
	case index:
		return "KEY INDEX"
	case unique:
		return "UNIQUE INDEX"
	case fk:
		return "FOREIGN KEY"
	case check:
		return "CHECK"
	default:
		return "<unknown>"
	}
}
