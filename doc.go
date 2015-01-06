// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// orm提供了一个相对统一的数据库操作，目前内置了对部分数据库的支持，
// 用户也可以自行实现core.Dialect接口，来实现特定数据库的支持。
//
// 支持的数据库：
//  1. sqlite3: github.com/mattn/go-sqlite3
//  2. mysql:   github.com/go-sql-driver/mysql
//  3. postgres:github.com/lib/pq
// 其它数据库，用户可以通过实现orm/core.Dialect接口，
// 然后调用orm.Register()注册来实现支持。
//
// 初始化：
//
// 默认情况下，orm包并不会加载任何数据库的实例。所以想要用哪个数据库，需要手动初始化：
//  import (
//      github.com/issue9/orm          // orm.Register(...)注册dialect
//      _ github.com/mattn/go-sqlite3  // 加载数据库驱动
//  )
//
//  // 向orm包注册dialect
//  orm.Register("sqlite3", &dialect.Sqlite3{})
//
//  // 初始化一个Engine，表前缀为prefix_
//  db1 := orm.New("sqlite3", "./db1", "db1", "prefix_")
//
//  // 另一个Engine
//  db2 := orm.New("sqlite3", "./db2", "db2", "db2_")
//
// Model:
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
// 通过struct tag可以直接将一个结构体定义为一个数据表结构，
// struct tag的语法结构，如上面代码所示，目前支持以下的struct tag：
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
//  定义物理外键，最少需要指定fk_name,refTabl,refColName三个值。
//  分别对应约束名，引用的表和引用的字段，updateRule,deleteRule，
//  在不指定的情况下，使用数据库的默认值。
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
// 当然在go中receive区分值类型和指针类型，所以指定接口时，需要注意这个情况。
//
// 约束名：
//
// index,unique,check,fk都是可以指定约束名的，在表中，约束名必须是唯一的，
// 即便是不同类型的约束，比如已经有一个unique的约束名叫作name，
// 那么其它类型的约束，就不能再取这个名称了。
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
// Upgrade:
//  // 创建或更新表
//  e.Upgrade(&User{})
//  // 创建或是更新多个表
//  e.Upgrade(&User{},&Email{})
//
// Update:
//  // 将id为1的记录的FirstName更改为abc
//  e.Update(&User{Id:1,FirstName:"abc"})
//  e.Where("id=?", 1).Add("FirstName", "abc").Update()
//  e.Where("id=?").Columns("FirstName").Update("abc", 1)
//
// Delete:
//  // 删除id为1的记录
//  e.Delete(&User{Id:1})
//  e.Where("id=?").Delte(1)
//  e.Where("id=?", 1).Delete()
//
// Insert:
//  // 一次性插入一条数据
//  e.Insert(&User{Id:1,FirstName:"abc"})
//  // 一次性插入多条数据
//  e.Insert([]*User{&User{Id:1,FirstName:"abc"},&User{Id:1,FirstName:"abc"}})
//
// Select:
//  // 导出id=1的数据
//  m, err := e.Where("id=?", 1).FetchMap()
//  // 导出id<5的所有数据
//  m, err := e.Where("id<?", 1).FetchMaps(5)
//
// 更新表结构时，以下情况不能被正确处理：
//  1. 改字段名；
//  2. 字段改为与原来不兼容的类型，比如从varchar到int；
//
// 事务：
//
// 默认的Engine是不支持事务的，若需要事务支持，则需要调用Engine.Begin()
// 返回事务对象Tx，当然并不是所有的数据库都支持事务操作的。
// Tx拥有与Engine相似的接口。
package orm

// 版本号
const Version = "0.6.10.150105"
