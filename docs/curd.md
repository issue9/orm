DB 和 Tx 对象都提供了一套基于数据模型的基本操作，
功能上比较单一，如果需要复杂的 SQL 操作，则需要使用 sqlbuilder
下的内容。

像 Update、Delete 和 Select 等操作，需要指定查询条件的，
在 DB 和 Tx 中会从当前提交的对象中查找可用的查询条件。
可用的查询条件是指 AI、PK 和唯一约束中，所有值都不为零值的那一个约束。
所以 Update、Delete 和 Select 操作都是单一对象。


以下展示了 CRUD 的一些基本操作。假设操作对象为以下结构：

```go
type User struct {
    ID       int64  `orm:"name(id);ai"`
    Name     string `orm:"name(name);len(20);index(i_user_name)"`
    Age      int    `orm:"name(age)"`
    Username string `orm:"name(username);unique(u_unique_username)"`
}
```

### TransactionalDDL

`core.Dialect.TransactionalDDL()` 指定了当前数据是否支持在事务中执行 DDL 语句。


像 `db.Create()` 可能存在执行多条语句，比如：
```sql
CREATE TABLE users (
    id INT NOT NULL,
    name VARCHAR(20) NOT NULL,
);
CREATE INDEX i_user_index ON users (name);
```
两条 create 才组成一个完整的创建表的操作。


如果不支持 TransactionalDDL 的，那么这些语句会分开执行，中断出错了，也没法回滚；
而支持 TransactionalDDL 的，这些步骤只出错，都会被撤消。

所在以 TransactionalDDL 值不同的数据库中，执行某些操作，其行为可能会有稍微的差别。


### create

用于创建数据表。
```go
err := db.Create(&User{})
```
会创建一张 User 表，表名由 User.Meta() 方法指定中的 name 属性指定，
如果没有 Meta() 方法，则直接采用结构体名称作为表名。
当然如果你在初始化 DB 时，指定了表名前缀，这里也会加上。

创建表属于 DDL，如果数据不支持事务中执行 DDL，那么即使在事务中，
创建表也依然是逐条提交的。


同时还提供了 MultCreate() 方法，用于同时创建多张表。
会根据是否支持事务内 DDL 属性，自动决定是否在一个事务中完成。


### insert

```go
result, err := db.Insert(&User{
    Name: "name",
})
```

插入 User{} 对象到数据库，不需要指定自增列的值，会自动生成。
其 name 字段的值为 name，其它字段都采用默认值。

#### lastInsertID

```go
// id 为当前插入数据的自增 ID
id, err := db.LastInsertID(&User{
    Name: "name",
})
```
如果需要获得 Last Insert ID 的值，建议采用 `LastInsertID` 方法获取，
而不是通过 `sql.Result.LastInsertID`，部分数据库（比如 postgres）
无法通过 `sql.Result.LastInertID` 获取 ID，但是 `db.LastInsertID`
会处理这种情况。

必须要有自增列，否则会出错！


### update

```go
result, err := db.Update(&User{
    ID:   1,
    Name: "test",
    Age:  0, // 零值，不会更新到数据库
})

result, err := db.Update(&User{
    ID:   1,
    Name: "test",
    Age:  0,
}, "age") // 指定了 age 必须更新，即使是零值
```

update 会根据当前传递对象的非零值字段中查找 AI、PK 和唯一约束，
只要找到了就符根据这些约束作为查询条件，其它值作为更新内容进行更新。

默认情况下，零值不会被更新，这在大对象中，会节省不少操作。
当然如果需要更新零值到数据库，则需要在 `Update()` 的第二个参中指定列名。

如果需要更新 AI、PK 和唯一约束本身的内容，可以通过 sqlbuilder
进行一些高级的操作。


### delete

delete 和 update 一样，通过唯一查询条件确定需要删除的列，并执行删除操作。
```go
// 删除 ID 为 1 的行。
result, err := db.Delete(&User{ID: 1})

// 删除 username 值为 example 的行
result, err = db.Delete(&User{Usrname: "example"})


// 同时指这了 AI 和唯一约束，则优先 AI 作查询。
result, err = db.Delete(&User{
    ID: 1,
    Usrname: "example",
})

// 返回错误，查询条件必须要有表达唯一性。
result, err = db.Delete(&User{
    Age: 18,
})
```


### truncate

truncate 会清空表内容，同时将该的自增计数重置为从 1 开始。
```go
err :=db.Truncate(&User{})
```

### select

```go
// 查找 ID 为 1 的 User 数据。会将 u 的其它字段填上。
u := &User{ID: 1}
err := db.Select(u)
```

### count

count 用于统计符合指定条件的所有数据。所有非零值都参与计算，
以 `AND` 作为各个查询条件的连接。
```go
// 相当于 SELECT count(*) FROM users WHERE name='name' AND age=18
count, err := db.Count(&User{
    Name: "name",
    Age: 18,
})
```
