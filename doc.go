// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// 一个简单小巧的orm实现方案。
//
// user := &models.User{}
// sql.Table(sqlbuilder.Table("table").As("abc")).
// Select(As(user.ID,"id"),user.Name).
package orm

// 数据表的更改，涉及到很多方面：
//  1.字段名更改；
//  2.字段类型更改，且不兼容旧类型；
//  3.名称加了特殊修饰符：比如sqlite中的[group]与`group`为同一字段名，但表示形式不同；
// 所有这一切都让UpgradeTable()这一功能的实现变得很糟糕，
// 在没有比较完美的方法之前，不准备实现这个功能。

// 版本号
const Version = "0.11.24.150518"
