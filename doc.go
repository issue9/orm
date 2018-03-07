// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package orm 一个简单小巧的 orm 实现方案。
//
// 目前内置了对以下数据库的支持：
//  1. sqlite3:  github.com/mattn/go-sqlite3
//  2. mysql:    github.com/go-sql-driver/mysql
//  3. postgres: github.com/lib/pq
// 其它数据库，用户可以通过实现 Dialect 接口，来实现相应的支持。
//
//
//
// 初始化：
//
// 默认情况下，orm 包并不会加载任何数据库的实例。所以想要用哪个数据库，需要手动初始化：
//  import (
//      _ github.com/mattn/go-sqlite3    // 加载数据库驱动
//      _ github.com/issue9/orm/dialect  // sqlite3 的 dialect 声明在此处
//  )
//
//  // 初始化一个 DB，表前缀为 prefix_
//  db1 := orm.NewDB("sqlite3", "./db1", "prefix_", dialect.Sqlite3())
//
//  // 另一个 DB 实例
//  db2 := orm.NewDB("sqlite3", "./db2", "db2_", dialect.Sqlite3())
//
//
//
// 占位符
//
//
// SQL 语句可以使用 # 字符在语句中暂替真实的表名前缀，也可以使用 {}
// 包含一个关键字，使其它成为普通列名，如：
//  select * from #user where {group}=1
// 在实际执行时，如 DB.Query()，将第一个参数 replace 指定为 true，
// 相关的占位符就会被替换成与当前环境想容的实例，如在表名前缀为 p_，
// 数据库为 mysql 时，会被替换成以下语句，然后再执行：
//  select * from p_user where `group`=1
// DB.Query(),DB.Exec(),DB.Prepare().DB.Where() 及 Tx 与之对应的函数都可以使用占位符。
//
// Model 不能指定占位符，它们默认总会使用占位符，且无法取消。
//
//
//
// Model:
//
// orm 包通过 struct tag 来描述 model 在数据库中的结构。大概格式如下：
//  type User struct {
//      Id          int64      `orm:"name(id);ai;"`
//      FirstName   string     `orm:"name(first_name);index(index_name)"`
//      LastName    string     `orm:"name(first_name);index(index_name)"`
//
//      // 此处group会自动加上引号，无须担心是否为关键字
//      Group       string	   `orm:"name(group)"`
//  }
//
//  // 通过 orm.Metaer 接口，指定表的额外数据。若不需要，可不用实现该接口
//  // 表名 user 会被自动加上表名前缀。
//  func(u *User) Meta() string {
//      return "name(user);engine(innodb);charset(utf-8)"
//  }
//
// 目前支持以下的 struct tag：
//
//  name(fieldName): 将当前的字段映射到数据表中的 fieldName 字段。
//
//  len(l1, l2): 指定字段的长度。比如 mysql 中的int(5),varchar(255),double(1,2),
//  不支持该特性的数据，将会忽略该标签的内容，比如 sqlite3。
//  NOTE:字符串类型必须指定长度，若长度过大或是将长度设置了-1，
//  想使用类似于 TEXT 等不定长的形式表达。
//
//  nullable(true|false): 相当于定义表结构时的 NULL，建议尽量少用该属性，
//  若非用不可的话，与之对应的 Go 属性必须声明为 NullString之类的结构。
//
//  pk: 主键，支持联合主键，给多个字段加上pk的struct tag即可。
//
//  ai: 自增，若指定了自增列，则将自动取消其它的 pk 设置。无法指定起始值和步长。
//  可手动设置一个非零值来更改某条数据的 AI 行为。
//
//  unique(index_name): 唯一索引，支持联合索引，index_name 为约束名，
//  会将 index_name 为一样的字段定义为一个联合索引。
//
//  index(index_name): 普通的关键字索引，同 unique 一样会将名称相同的索引定义为一个联合索引。
//
//  default(value): 指定默认值。相当于定义表结构时的 DEFAULT。
//  当一个字段如果是个零值(reflect.Zero())时，将会使用它的默认值，
//  但是系统无法判断该零值是人为指定，还是未指定被默认初始化零值的，
//  所以在需要用到零值的字段，最好不要用 default 的 struct tag。
//
//  fk(fk_name,refTable,refColName,updateRule,deleteRule):
//  定义物理外键，最少需要指定 fk_name,refTabl,refColName 三个值。分别对应约束名，
//  引用的表和引用的字段，updateRule,deleteRule，在不指定的情况下，使用数据库的默认值。
//
//  check(chk_name, expr): check 约束。chk_name 为约束名，expr 为该约束的表达式。
//  check 约束只能在 core.Metaer 接口中指定，而不是像其它约束一样，通过字段的 struct tag 指定。
//  因为 check 约束的表达式可以通过 and 或是 or 等符号连接多条基本表达式，
//  在字段 struct tag 中指定会显得有点怪异。
//
//
// core.Metaer:
//
// 在 Go 不能将 struct tag 作用于结构体，所以为了指定一些表级别的属性，
// 只能通过接口的形式，在接口方法中返回一段类似于 struct tag 的字符串，
// 以达到相同的目的。
//
// 在 core.Metaer 中除了可以指定 name(table_name) 和 check(name,expr) 两个属性之外，
// 还可指定一些自定义的属性，这些属性都将会被保存到 Model.Meta 中。
//
//
//
// 约束名：
//
// index,unique,check,fk 都是可以指定约束名的，在表中，约束名必须是唯一的，
// 即便是不同类型的约束，比如已经有一个 unique 的约束名叫作 name，那么其它类
// 型的约束，就不能再取这个名称了。
//
//
//
// 如何使用：
//
// Create:
// 可以通过 DB.Create() 或是 Tx.Create() 创建一张表。
//  // 创建表
//  db.Create(&User{})
//  // 创建多个表，同时创建多张表，主使用 Tx.Create
//  tx.MultCreate(&User{},&Email{})
//
// Update:
//  // 将 id 为 1 的记录的 FirstName 更改为 abc；对象中的零值不会被提交。
//  db.Update(&User{Id:1,FirstName:"abc"})
//  sqlbuilder.Update(db, "#table").Where("id=?",1).Set("FirstName", "abc").Exec()
//
// Delete:
//  // 删除 id 为 1 的记录
//  e.Delete(&User{Id:1})
//  sqlbuilder.Delete(e, "#table").Where("id=?",1).Exec()
//
// Insert:
//  // 插入一条数据
//  db.Insert(&User{Id:1,FirstName:"abc"})
//  // 一次性插入多条数据
//  tx.InsertMany(&User{Id:1,FirstName:"abc"},&User{Id:1,FirstName:"abc"})
//
// Select:
//  // 导出 id=1 的数据
//  _,err := sqlbuilder.Select(e).Select("*").From("{#table}").Where("id=1").QueryObj(obj)
//  // 导出 id 为 1 的数据，并回填到 user 实例中
//  user := &User{Id:1}
//  err := e.Select(u)
//
// Query/Exec:
//  // Query 返回参数与 sql.Query 是相同的
//  sql := "select * from #tbl_name where id=?"
//  rows, err := e.Query(true, sql, []interface{}{5})
//  // Exec 返回参数与 sql.Exec 是相同的
//  sql = "update #tbl_name set name=? where id=?"
//  r, err := e.Exec(true, sql, []interface{}{"name1", 5})
//
// 事务：
//
// 默认的 DB 是不支持事务的，若需要事务支持，则需要调用 DB.Begin()
// 返回事务对象 Tx，当然并不是所有的数据库都支持事务操作的。
// Tx拥有一组与 DB 相同的接口，另外还提供了一组以 `Mult` 开头的函数，
// 用以同时操作多条记录的。
package orm

// 数据表的更改，涉及到很多方面：
//  1.字段名更改；
//  2.字段类型更改，且不兼容旧类型；
//  3.名称加了特殊修饰符：比如sqlite中的[group]与`group`为同一字段名，但表示形式不同；
// 所有这一切都让UpgradeTable()这一功能的实现变得很糟糕，
// 在没有比较完美的方法之前，不准备实现这个功能。
