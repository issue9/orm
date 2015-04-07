// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// 一个简单小巧的orm实现方案。
//
// 目前内置了对以下数据库的支持：
//  1. sqlite3:  github.com/mattn/go-sqlite3
//  2. mysql:    github.com/go-sql-driver/mysql
//  3. postgres: github.com/lib/pq
// 其它数据库，用户可以通过实现orm/core.Dialect接口，来实现相应的支持。
// 具体操作参照后面的如何实现Dialect章节。
//
//
//
// 初始化：
//
// 默认情况下，orm包并不会加载任何数据库的实例。所以想要用哪个数据库，需要手动初始化：
//  import (
//      github.com/issue9/orm/core
//      _ github.com/mattn/go-sqlite3  // 加载数据库驱动
//  )
//
//  // 注册dialect
//  core.Register("sqlite3", &dialect.Sqlite3{})
//
//  // 初始化一个Engine，表前缀为prefix_
//  db1 := orm.New("sqlite3", "./db1", "db1", "prefix_")
//
//  // 另一个Engine
//  db2 := orm.New("sqlite3", "./db2", "db2", "db2_")
//
//
//
// Model:
//
// orm包通过struct tag来描述model在数据库中的结构。大概格式如下：
//  type User struct {
//      Id          int64      `orm:"name(id);ai;"`
//      FirstName   string     `orm:"name(first_name);index(index_name)"`
//      LastName    string     `orm:"name(first_name);index(index_name)"`
//  }
//
//  // 通过orm/core.Metaer接口，指定表的额外数据。若不需要，可不用实现该接口
//  func(u *User) Meta() string {
//      return "name(user);engine(innodb);charset(utf-8)"
//  }
//
// 目前支持以下的struct tag：
//
//  name(fieldName): 将当前的字段映射到数据表中的fieldName字段。
//
//  len(l1, l2): 指定字段的长度，比如mysql中的int(5),varchar(255),double(1,2),
//  仅部分数据库支持，比如sqlite3不支持该属性。
//
//  nullable(true|false): 相当于定义表结构时的NULL，建议尽量少用该属性，
//  若非用不可的话，与之对应的Go属性必须声明为NullString之类的结构。
//
//  pk: 主键，支持联合主键，给多个字段加上pk的struct tag即可。
//
//  ai: 自增，若指定了自增列，则将自动取消其它的pk设置。无法指定起始值和步长。
//
//  unique(index_name): 唯一索引，支持联合索引，index_name为约束名，
//  会将index_name为一样的字段定义为一个联合索引。
//
//  index(index_name): 普通的关键字索引，同unique一样会将名称相同的索引定义为一个联合索引。
//
//  default(value): 指定默认值。相当于定义表结构时的DEFAULT。
//
//  fk(fk_name,refTable,refColName,updateRule,deleteRule):
//  定义物理外键，最少需要指定fk_name,refTabl,refColName三个值。分别对应约束名，
//  引用的表和引用的字段，updateRule,deleteRule，在不指定的情况下，使用数据库的默认值。
//
// 关于core.Metaer接口。
//
// 在go不能将struct tag作用于结构体，所以为了指定一些表级别的属性，
// 只能通过接口的形式，在接口方法中返回一段类似于struct tag的字符串，
// 以达到相同的目的。
//
// 在core.Metaer中除了可以指定name(table_name)和check(name,expr)两个属性之外，
// 还可指定一些自定义的属性，这些属性都将会被保存到Model.Meta中。
//
// NOTE:一定要注意receive的类型是值还是指针。
//
// 约束名：
//
// index,unique,check,fk都是可以指定约束名的，在表中，约束名必须是唯一的，
// 即便是不同类型的约束，比如已经有一个unique的约束名叫作name，
// 那么其它类型的约束，就不能再取这个名称了。
//
//
//
// 如何使用：
//
// Create:
// 可以通过Engine.Create()或是Tx.Create()创建一张表。
//  // 创建表
//  e.Create(&User{})
//  // 创建多个表
//  e.Create(&User{},&Email{})
//
// Update:
//  // 将id为1的记录的FirstName更改为abc；对象中的零值不会被提交。
//  e.Update(&User{Id:1,FirstName:"abc"})
//  e.Where("id=1").Set("FirstName", "abc").Update(nil)
//  e.Where("id=@id").Data(map[string]interface{"FirstName":"abc"}).Update(map[string]interface{"id":1})
//
// Delete:
//  // 删除id为1的记录
//  e.Delete(&User{Id:1})
//  e.Where("id=@id").Delte(map[string]interface{"id":1})
//  e.Where("id=1").Delete(nil)
//
// Insert:
//  // 一次性插入一条数据
//  e.Insert(&User{Id:1,FirstName:"abc"})
//  // 一次性插入多条数据
//  e.Insert([]*User{&User{Id:1,FirstName:"abc"},&User{Id:1,FirstName:"abc"}})
//
// Select:
//  // 导出id=1的数据
//  m, err := e.Where("id=1").FetchMap(nil)
//  // 导出id<5的所有数据
//  m, err := e.Where("id<@id").FetchMaps(map[string]interface{"id":5})
//
// 事务：
//
// 默认的Engine是不支持事务的，若需要事务支持，则需要调用Engine.Begin()
// 返回事务对象Tx，当然并不是所有的数据库都支持事务操作的。
// Tx拥有与Engine相似的接口。
//
//
// 如何实现自定义Dialect:
//
// 实现core.Dialect接口，需要使用时，加载数据库驱动，然后向core.Register()函数注册实例，
// 即可正常使用。具体的实现可以参照dialect子包的相关代码。
package orm

// 数据表的更改，涉及到很多方面：
//  1.字段名更改；
//  2.字段类型更改，且不兼容旧类型；
//  3.名称加了特殊修饰符：比如sqlite中的[group]与`group`为同一字段名，但表示形式不同；
// 所有这一切都让UpgradeTable()这一功能的实现变得很糟糕，
// 在没有比较完美的方法之前，不准备实现这个功能。

// 版本号
const Version = "0.9.17.150407"
