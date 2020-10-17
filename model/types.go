// SPDX-License-Identifier: MIT

package model

import "github.com/issue9/orm/v3/core"

// DefaultParser 提供了 ParseDefault 函数
//
// 在 struct tag 中可以通过 default 指定默认值，
// 该值的表示可能与数据库中的表示不尽相同，
// 所以自定义的数据类型，需要实现该接口，以便能正确转换成该类型的值。
//
// 如果用户不提供该接口实现，那么默认情况下，
// 系统会采用 github.com/issue9/conv.Value() 函数作默认转换。
type DefaultParser interface {
	// 将默认值从字符串解析成 t 类型的值
	ParseDefault(v string) error
}

// Viewer 如果是视图模型，需要实现此接口
type Viewer interface {
	// 返回视图所需的 Select 语句。
	ViewAs(e core.Engine) (string, error)
}

// Metaer 用于指定数据模型的元数据
//
// 不同的数据库可以有各自的属性内容，具体的由 Dialect 的实现者定义。
// 但是 name、check 是通用的，分别表示名称和 check 约束。
//  "name(tbl_name);mysql_engine(myISAM);mysql_charset(utf8)"
type Metaer interface {
	Meta() string
}
