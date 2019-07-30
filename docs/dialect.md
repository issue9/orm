
目前 orm 包本身定义了 Postgres、Sqlite3 和 Mysql 三个类型数据库的支持。
如果用户需要其它类型的数据库操作，可以自己实现 core.Dialect 接口。
