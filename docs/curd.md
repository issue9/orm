
以下展示了 CRUD 的一些基本操作。假设操作对象为以下结构：

```go
type User struct {
    ID   int64  `orm:"name(id);ai"`
    Name string `orm:"name(name);len(20);index(i_user_name)"`
    Age  int    `orm:"name(age)"`
}
```


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
id, err := db.LastInsertID(&User{
    Name: "name",
})
```
如果需要 insert last id 的值，建议采用 `InsertLastID` 方法获取，
而不是通过 `sql.Result.LastInsertID`，部分数据库（比如 postgres）
无法通过 `sql.Result.LastInertID` 获取 ID，但是 `db.LastInsertID`
会处理这种情况。


### update

```go
result, err := db.Update(&User{
    ID:   1,
    Name: "test",
    Age:  0,
}, "age")
```

update 会根据当前传递对象的非零值字段中查找 AI、PK 和唯一约束，
只要找到了就符根据这些约束作为查询条件，其它值作为更新内容进行更新。

默认情况下，零值不会被更新，这在大对象中，会节省不少操作。
当然如果需要更新零值到数据库，则需要在 `Update()` 的第二个参中指定。

如果需要更新 AI、PK 和唯一约束本身的内容，可以通过 sqlbuilder
进行一些高级的操作。


### delete

delete 和 update 一样，通过唯一查询条件确定需要删除的列，并执行删除操作。
```go
result, err := db.Delete(&User{ID: 1})
```



### select


### count
