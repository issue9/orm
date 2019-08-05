orm
[![Build Status](https://travis-ci.org/issue9/orm.svg?branch=master)](https://travis-ci.org/issue9/orm)
[![Go version](https://img.shields.io/badge/Go-1.10-brightgreen.svg?style=flat)](https://golang.org)
[![Go Report Card](https://goreportcard.com/badge/github.com/issue9/orm)](https://goreportcard.com/report/github.com/issue9/orm)
[![codecov](https://codecov.io/gh/issue9/orm/branch/master/graph/badge.svg)](https://codecov.io/gh/issue9/orm)
[![GoDoc](https://godoc.org/github.com/issue9/orm?status.svg)](https://godoc.org/github.com/issue9/orm)
======

目前内置了对以下数据库的支持：

 1. sqlite3:  github.com/mattn/go-sqlite3
 2. mysql:    github.com/go-sql-driver/mysql
 3. postgres: github.com/lib/pq

其它数据库，用户可以通过实现 Dialect 接口，来实现相应的支持。


#### 初始化

默认情况下，orm 包并不会加载任何数据库的实例。所以想要用哪个数据库，需要手动初始化：

```go
import (
    github.com/issue9/orm/v2/dialect  // sqlite3 的 dialect 声明在此处
    _ github.com/mattn/go-sqlite3  // 加载数据库驱动
)

// 初始化一个 DB，表前缀为 prefix_
db1 := orm.NewDB("sqlite3", "./db1", "prefix_", dialect.Sqlite3())

// 另一个 DB 实例
db2 := orm.NewDB("sqlite3", "./db2", "db2_", dialect.Sqlite3())
```


#### 占位符


SQL 语句可以使用 # 字符在语句中暂替真实的表名前缀，也可以使用 {}
包含一个关键字，使其它成为普通列名，如：
```sql
SELECT * FROM #user WHERE {group}=1
```
在实际执行时，相关的占位符就会被替换成与当前环境想容的实例，如在表名前缀为 p_，
数据库为 mysql 时，会被替换成以下语句，然后再执行：
```sql
 SELECT * FROM p_user WHERE `group`=1
```
DB.Query(),DB.Exec(),DB.Prepare().DB.Where() 及 Tx 与之对应的函数都可以使用占位符。

Model 不能指定占位符，它们默认总会使用占位符，且无法取消。


#### Model:

orm 包通过 struct tag 来描述 model 在数据库中的结构。大概格式如下：

```go
type User struct {
    Id          int64      `orm:"name(id);ai;"`
    FirstName   string     `orm:"name(first_name);index(index_name)"`
    LastName    string     `orm:"name(first_name);index(index_name)"`

    // 此处group会自动加上引号，无须担心是否为关键字
    Group       string	   `orm:"name(group)"`
}

// 通过 orm.Metaer 接口，指定表的额外数据。若不需要，可不用实现该接口
// 表名 user 会被自动加上表名前缀。
func(u *User) Meta() string {
    return "name(user);engine(innodb);charset(utf8)"
}
```

目前支持以下的 struct tag：

##### name(fieldName): 

指定当前字段在数据表中的名称，如果未指定，则和字段名相同。
只有可导出的字段才有效果。

##### len(l1, l2): 

指定字段的长度。比如 mysql 中的int(5),varchar(255),double(1,2),
不支持该特性的数据，将会忽略该标签的内容，比如 sqlite3。
NOTE:字符串类型必须指定长度，若长度过大或是将长度设置了-1，
想使用类似于 TEXT 等不定长的形式表达。

如果是日期类型，则第一个可选参数表示日期精度。

##### nullable(true|false):

相当于定义表结构时的 NULL。

##### pk:

主键，支持联合主键，给多个字段加上 pk 的 struct tag 即可。

##### ai:

自增，若指定了自增列，则将自动取消其它的 pk 设置。无法指定起始值和步长。
可手动设置一个非零值来更改某条数据的 AI 行为。

##### unique(index_name):

唯一索引，支持联合索引，index_name 为约束名，
会将 index_name 为一样的字段定义为一个联合索引。

##### index(index_name):

普通的关键字索引，同 unique 一样会将名称相同的索引定义为一个联合索引。

##### occ(true|false)

当前列作为乐观锁字段。

作为乐观锁的字段，其值表示的是线上数据的值，在更新时，会自动给线上的值加 1。

##### default(value):

指定默认值。相当于定义表结构时的 DEFAULT。
当一个字段如果是个零值(reflect.Zero())时，将会使用它的默认值，
但是系统无法判断该零值是人为指定，还是未指定被默认初始化零值的，
所以在需要用到零值的字段，最好不要用 default 的 struct tag。

##### fk(fk_name,refTable,refColName,updateRule,deleteRule):

定义物理外键，最少需要指定 fk_name、refTable 和 refColName 三个值。分别对应约束名，
引用的表和引用的字段，updateRule,deleteRule，在不指定的情况下，使用数据库的默认值。
refTable 如果需要表名前缀，需要添加 # 符号。

##### check(chk_name, expr):

check 约束。chk_name 为约束名，expr 为该约束的表达式。
check 约束只能在 model.Metaer 接口中指定，而不是像其它约束一样，通过字段的 struct tag 指定。
因为 check 约束的表达式可以通过 and 或是 or 等符号连接多条基本表达式，
在字段 struct tag 中指定会显得有点怪异。


#### 接口:

##### Metaer

在 Go 不能将 struct tag 作用于结构体，所以为了指定一些表级别的属性，
只能通过接口的形式，在接口方法中返回一段类似于 struct tag 的字符串，
以达到相同的目的。

在 model.Metaer 每种数据库可以指定一些自定义的属性，这些属性都将会被保存到
Model.Meta 中，各个数据库的自定义属性以其名称开头，比如 mysql 的 charset 属性，
可以使用 mysql_charset。

同时还包含了以下几种通用的属性信息：
- name(val): 指定模型的名称，比如表名，或是视图名称；
- check(expr): 指定 check 约束内容。


##### AfterFetcher

在拉到数据之后，对该对象执行的一些额外操作。如果需要根据字段做额外工作的，可以使用该接口。


##### BeforeInserter

在执行插入之前，需要执行的操作。


##### BeforeUpdater

在执行更新之前，需要执行的操作。


#### 约束名：

index、unique、check 和 fk 都是可以指定约束名的，在当前数据库中，
约束名必须是唯一的，即便是不同类型的约束，比如已经有一个 unique 的约束名叫作 name，
那么其它类型的约束，就不能再取这个名称了。


#### 如何使用：

##### Create:

可以通过 DB.Create() 或是 Tx.Create() 创建一张表。

```go
// 创建表
db.Create(&User{})
// 创建多个表，同时创建多张表，主使用 Tx.Create
tx.MultCreate(&User{},&Email{})
```

##### Update:

```go
// 将 id 为 1 的记录的 FirstName 更改为 abc；对象中的零值不会被提交。
db.Update(&User{Id:1,FirstName:"abc"})
sqlbuilder.Update(db).Table("#table").Where("id=?",1).Set("FirstName", "abc").Exec()
```

##### Delete:

```go
// 删除 id 为 1 的记录
e.Delete(&User{Id:1})
sqlbuilder.Delete(e).Table("#table").Where("id=?",1).Exec()
```


##### Insert:

```go
// 插入一条数据
db.Insert(&User{Id:1,FirstName:"abc"})
// 一次性插入多条数据
tx.InsertMany(&User{Id:1,FirstName:"abc"},&User{Id:1,FirstName:"abc"})
```

##### Select:

```go
// 导出 id=1 的数据
_,err := sqlbuilder.Select(e, e.Dialect()).Select("*").From("{#table}").Where("id=1").QueryObj(obj)
// 导出 id 为 1 的数据，并回填到 user 实例中
user := &User{Id:1}
err := e.Select(u)
```

##### Query/Exec:

```go
// Query 返回参数与 sql.Query 是相同的
sql := "select * from #tbl_name where id=?"
rows, err := e.Query(sql, []interface{}{5})
// Exec 返回参数与 sql.Exec 是相同的
sql = "update #tbl_name set name=? where id=?"
r, err := e.Exec(sql, []interface{}{"name1", 5})
```


##### 多表查询

```go
type order struct {
    ID int `orm:"name(id);ai"`
    UID int `orm:"name(uid)"`
    User *User `orm:"name(u)"` // sql 语句的 u. 开头的列导入到此对象中
}

sql := "SELECT o.*, u.id AS 'u.id',u.name AS 'u.name' FROM orders AS o LEFT JOIN  users AS u ON o.uid= u.id where id=?"

// 以下查询会将 u.name 和 u.id 导入到 User 指向的对象中
data := []*order{}
rows,err := e.Query(sql, []interface{}{5})
r, err := fetch.Object(rows, &data)
```

##### InsertLastID:

像 postgres 之类的数据库，插入语句返回的 sql.Result 并不支持 LastInsertId()
所以得通过其它方式得到该 ID，可以直接尝试使用 InsertLastID 来代码 Insert 操作。

```go
lastID,err := db.InsertLastID(&User{FirstName:"abc"})
```

#### 事务：

默认的 DB 是不支持事务的，若需要事务支持，则需要调用 DB.Begin()
返回事务对象 Tx，当然并不是所有的数据库都支持事务操作的。
Tx 拥有一组与 DB 相同的接口。



### 版权

本项目采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
