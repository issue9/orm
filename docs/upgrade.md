
upgrade 可以当作是数据库升级的小助手，在大部分情况下，都能胜任。
如果数据库本身是支持事务中执行 DDL 的，那么在执行失败时，会回滚。

```go
type User {
    ID int64 `orm:"name(id)"`
    Name string `orm:"name(name);len(20)"` // 新加的列
}
err := db.Upgrade(&User{}).
    AddColumn("name"). // 将该列添加到数据库
    DropColumn("username"). // 删除数据库中的 username 列
    Do() // 执行以上操作
```
