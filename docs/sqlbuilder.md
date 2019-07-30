
sqlbuilder 提供了一组以链式操作构建 SQL 语句的方法。
且对象本身可以复用。

```go
stmt := sqlbuilder.Select("*").
    From("users", "u").
    Where("id>@id", sql.Named("id", 1)).
    And("age<@age", sql.Named("age", 100)).
    Desc("id")

// 所有符合条件的数据会导出到 list 中，如果 list 不够长，会自动添加元素。
list := make([]*User, 0, 10)
stmt.QueryObject(false, &list)
```

### 命名参数


### CreateTable
