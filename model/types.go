// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"github.com/issue9/orm/v2/core"
	"github.com/issue9/orm/v2/sqlbuilder"
)

// DefaultParser 提供了 ParseDefault 函数
//
// 在 struct tag 中可以通过 default 指定默认值，
// 该值的表示可能与数据库中的表示不尽相同，
// 所以自定义的数据类型，需要实现该接口，以便能正确转换成该类型的值。
//
// 如果用户不提供该接口实现，那么默认情况下，
// 系统会采用 github.com/issue9/conv.Value() 函数作默认转换。
type DefaultParser interface {
	// 将默认值从字符串解析成 t 类型的值。
	ParseDefault(v string) error
}

// Viewer 如果是视图模型，需要实现此接口
type Viewer interface {
	// 返回视图所需的 Select 语句实例。
	ViewAs(e core.Engine) *sqlbuilder.SelectStmt
}

// Type 表示数据模型的类别
type Type int8

// 目前支持的数据模型类别
//
// Table 表示为一张普通的数据表，默认的模型即为 Table；
// 如果实现了 Viewer 接口，则该模型改变视图类型，即 View。
//
// 两者的创建方式稍微有点不同：
// Table 类型创建时，会采用列、约束和索引等信息创建表；
// 而 View 创建时，只使用了 Viewer 接口返回的 Select
// 语句作为内容生成语句，像约束等信息，仅作为查询时的依据，
// 当然 select 语句中的列需要和 Columns 中的列要相对应，
// 否则可能出错。
//
// 在视图类型中，唯一约束、主键约束、自增约束依然是可以定义的，
// 虽然不会呈现在视图中，但是在查询时，可作为 orm 的一个判断依据。
const (
	Table Type = iota
	View
)

// Metaer 用于指定数据模型的元数据。
//
// 不同的数据库可以有各自的属性内容，具体的由 Dialect 的实现者定义。
// 但是 name、check 是通用的，分别表示名称和 check 约束。
//  "name(tbl_name);mysql_engine(myISAM);mysql_charset(utf8)"
type Metaer interface {
	Meta() string
}
