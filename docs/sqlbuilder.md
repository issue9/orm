
sqlbuilder 提供了一组以链式操作构建 SQL 语句的方法。
且对象本身可以复用。

```go
stmt := sqlbuilder.Select(db)
    Column("*").
    From("users", "u").
    Where("id>@id", sql.Named("id", 1)).
    And("age<@age", sql.Named("age", 100)).
    Desc("id")

// 所有符合条件的数据会导出到 list 中，如果 list 不够长，会自动添加元素。
list := make([]*User, 0, 10)
stmt.QueryObject(false, &list)
```

### 命名参数

支持 Go 1.8 之后提供的 sql.NamedArgs 格式的命名参数。
在链式操作中，并不要求语句的顺序，比如：
```go
stmt1, err := sqlbuilder.Select(db)
    Column("*").
    From("users", "u").
    Where("id>?", 1).
    Limit(20, 10).
    Prepare()

stmt2, err := sqlbuilder.Select(db)
    From("users", "u").
    Column("*").
    Limit(20, 10).
    Where("id>?", 1).
    Prepare()
```
以上两个语句生成的 SQL 是一样的：
```sql
SELECT * FROM users AS u WHERE id>? LIMIT ? OFFSET ?
```
所以在调用预编译的语句时，给的参数也必须是一样的，
而不是按照链式语句的参数顺序就行了。
这就成造成了，一旦你的语句需要预编译，那么链式操作带来的只有麻烦而不是便捷。

所以在这类操作中，推荐使用命名参数的方式调用：
```go
stmt2, err := sqlbuilder.Select(db)
    From("users", "u").
    Column("*").
    Limit(sql.Named("limit", 20), sql.Named("offset", 10)).
    Where("id>@id", sql.Named("id", 1)).
    Prepare()
stmt2.Query([]interface{}{
    sql.Named("id", 1),
    sql.Named("offset", 30),
    sql.Named("limit", 20),
})
```


### 占位符

如果你表中的字段是 SQL 的关键字，那么在查询时，需要将该字段加特殊的引号，
各个数据库不尽相同。
在 sqlbuilder 中，你不需要关心具体是什么唯一，我们统一采用 `{}`，
系统会自动处理成当前数据库的符号。

当然大部分时候，你都不需要手动添加，系统会自动帮你添加，
只有在 `WHERE` 和 `CHECK` 表达式才需要：
```go
query := sqlbuilder.Select(db).
    From("users").
    Column("group").              // 此处不需要
    Where("{group} IS NOT NULL")  // WHERE 表达式需要加 {}
```


### CrateTable/TruncateTable/DropTable

```go
// 创建表
creator := sqlbuilder.CreateTable(db).
    Table("users").
    AutoIncrement("id", core.Int64Type). // 自增列
    Column("name", core.StringType, false, false, nil).
    Column("username", core.StringType, false, false, nil).
    Column("age", core.IntType, false, false, nil).
    PK("u_unique_username", "username").  // 唯一约束
    Check("chk_age_great_18", "age>18")

// 清空表
truncate := sqlbuilder.TruncateTable(db).
    Table("users", "id")

// 删除表
drop := sqlbuilder.DropTable(db).
    Table("users")
```

### Insert


### Select


### Delete
