
目前 orm 包本身定义了 Postgres、Sqlite3 和 Mysql 三个类型数据库的支持。
如果用户需要其它类型的数据库操作，可以自己实现 `core.Dialect` 接口。

Dialect 需要实现两个部分的内容：其中 core.Dialect 是必须要实现的接口，
另外，在 sqlbuilder 包中，还提供了一部分 **Hooker 的接定义，
如果你当前的数据库实现与 sqlbuilder 中的默认实现不一样，可需要自行实现该接口。

比如 `sqlbuilder.InsertDefaultValueHooker` 接口，默认实现，采用了比较常用的方法：

```sql
INSERT INTO table DEFAULT VALUES;
```

但是 mysql 没有对应的实现，需要自定义该口，而 postgres 和 sqlite3 不需要。
