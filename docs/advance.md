### 时区

sqlite3 可以通过 _loc 的方式指定时区；
mysql 可以通过 loc 参数指定时区；
postgres 无法指定时间，都直接当作时区 0 进行了处理；

如果你的代码后期需要在不同的数据库之间迁移，那么建议将时区统一设置为 UTC。

### 时间精度

各个数据库驱动对精度处理方式并不相同，mysql 将未设置精度等同于精度为 0，而
postgres 和 sqlite3 则会将其赞同于最大精度 6。

### 自定义类型

ORM 支持对自定义类型的存储和读取，需要实现以下几个接口：

- sql.Scan/driver.Valuer 这两个接口为标准库本身要求必须实现的；
- core.PrimitiveTyper 指定了底层的 Go 类型，该值会在创建表时用于判断应该创建的数据库类型；
- core.DefaultParser 表示在 struct tag 中的 Default 标签中值该如果解析，可以不实现，如果未实现，则采用 `sql.Scanner` 接口；
- core.TableNamer 指定表名；

可以查找 core.Unix 的实现方式。
