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
