// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// orm提供了一个相对统一的数据库操作，目前内置对mysql和sqlite3
// 的支持，用户可以通过实现orm/core.Dialect接口，实现对特定数据
// 库的支持。
//
// INSERT:
//  e := orm.New("mysql", "root:@/test", "maindb", "prefix_")
//  e.Insert().
//      Table().
//      Column(col1, col2).
//      Exec(args...)
//
// orm增加了数据操作的通用性，但相应的也损失了数据库独有的特性。
package orm
