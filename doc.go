// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// orm提供了一个相对统一的数据库操作，目前内置了对部分数据库的支持，
// 用户可以通过实现自定义来实现对特定数据库的支持。
//
// 支持的数据库：
//  1. sqlite3: github.com/mattn/go-sqlite3
//  2. mysql:   github.com/go-sql-driver/mysql
//  3. postgres:github.com/lib/pq
// 其它数据库，用户可以通过实现orm/core.Dialect接口，
// 然后调用orm/dialect.Register()注册来实现支持。
//
// 初始化：
//
// 默认情况下，orm包并不会加载任何数据库的实例。所以想要用哪个数据库，需要手动初始化：
//  // 加载dialect管理包
//  import github.com/issue9/orm/dialect
//
//  // 加载数据库驱动
//  import _ github.com/mattn/go-sqlite3
//
//  // 向orm/dialect包注册dialect
//  dialect.Register("sqlite3", &dialect.Sqlite3{})
//
//  // 之后可以调用db1的各类方法操作数据库
//  db1 := orm.New("sqlite3", "./db1", "db1", "prefix_")
//
// Model:
//  type User struct {
//      // 对应表中的id字段，为自增列，从0开始
//      Id          int64      `orm:"name(id);ai(0);"`
//      // 对应表中的first_name字段，为索引index_name的一部分
//      FirstName   string     `orm:"name(first_name);index(index_name)"`
//      LastName    string     `orm:"name(first_name);index(index_name)"`
//  }
//
//  通过orm/core.Metaer接口，指定表的额外数据。若不需要，可不用实现该接口
//  func(u *User) Meta() string {
//      return "name(user);engine(innodb);charset(utf-8)"
//  }
//
// Create:
//  // 创建或是更新表
//  e.Create(&User{})
//  // 创建或是更新多个表
//  e.Create([]*User{&User{},&Email{}})
//
// Update:
//  将id为1的记录的FirstName更改为abc
//  e.Update(&User{Id:1,FirstName:"abc"})
//
// nullable 将一个varchar设置成null，可能在导出时提示scan error <nil>to *string的错误
// Limit 参数将被替换成占位符，所以若要在Query()中传递参数，不要忘记limit中相关的参数
package orm

// 版本号
const Version = "0.1.1.141229"
