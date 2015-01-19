// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"github.com/issue9/orm/builder"
	"github.com/issue9/orm/core"
)

// 确保Engine和Tx都实现了这些接口
type db interface {
	core.DB
	Close() error
	SQL() *builder.SQL
	Where(cond string) *builder.SQL
	Insert(v interface{}) error
	Update(v interface{}) error
	Create(models ...interface{}) error
	Drop(tableName string) error
	Truncate(tableName string) error
}

var _ db = &Engine{}
var _ db = &Tx{}
