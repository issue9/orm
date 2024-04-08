### 快速开始

```go
import (
    "github.com/issue9/orm/v6"
    "github.com/issue9/orm/v6/dialect"

    _ "github.com/mattn/go-sqlite3"
)

type User struct {
    ID   int64  `orm:"name(id);ai" json:"id"`
    Name string `orm:"name(name);len(500)" json:"name"`
    Age  int    `orm:"name(age)" json:"age"`
}

func (u *User) TableName() string { return "users" }

// 指定了表名，以及其它一些表属性
func (u *User) ApplyModel(m*core.Model) error {
    m.Options["mysql_charset"] = []string{"utf8"}
    return nil
}

func main() {
    db, err := orm.NewDB("./test.db", dialect.Sqlite3("sqlite3", "test_"))
    if err !=nil {
        panic(err)
    }
    defer db.Close()

    // 创建表
    if err = db.Create(&User{});err != nil {
        panic(err)
    }

    // 插入一条数据，ID 自增为 1
    rslt, err := db.Insert(&User{
        Name: "test",
        Age: 18,
    })

    // 读取 ID 值为 1 的数据到 u 中
    u := &User{ID: 1}
    err = db.Select(u)

    // 更新，根据自增列 ID 查找需要更新列
    u = &User{ID: 1, Name: "name", Age: 100}
    rslt, err = db.Update(u)

    // 删除，根据自增 ID 查找唯一数据删除
    u = &User{ ID: 1}
    rslt, err = db.Delete(u)

    // 删除表
    db.Drop(&User{})
}
```


### 安装

在项目的 go.mod 中引用项目即可，当前版本为 v5：

```go.mod
require (
    github.com/issue9/orm/v6 v6.x.x
)

go 1.18
```

### 数据库

目前支持以下数据库以及对应的驱动:

 1. sqlite3:  github.com/mattn/go-sqlite3
 1. sqlite:   modernc.org/sqlite 纯 go 写的驱动
 1. mysql:    github.com/go-sql-driver/mysql
 1. mariadb:  github.com/go-sql-driver/mysql
 1. postgres: github.com/lib/pq

在初始化时，需要用到什么数据库，只需要引入该驱动即可。
如果用到了 check 约束，则需要 mysql > 8.0.16、mariadb > 10.2.1。

```go
import (
    "github.com/issue9/orm/v6"
    "github.com/issue9/orm/v6/dialect"

    _ "github.com/mattn/go-sql-driver/mysql"
    _ "github.com/mattn/mattn/go-sqlite3"
)
```

之后就可以直接使用 `orm.NewDB` 初始化实例。

或者在已经初始化 `sql.DB` 的情况下，直接使用 `sql.DB` 实例初始化：

```go
// 初始化 sqlite3 的实例
sqlite, err := orm.NewDB("./orm.db", dialect.Sqlite3())

// 初始化 mysql 的实例
db, err := sql.Open("mysql", "root@/orm")
my, err := orm.NewDBWithStdDB(db, dialect.Mysql())
```

后续代码中可以同时使用 my 和 sqlite 两个实例操纵不同的数据库数据。

### 调试

系统提供了 `DB.Debug(l log.Logger)` 用于调试，调用时，如果指定了非空参数，
那么在下次 SQL 调用时，会将 SQL 输出到 l 指定的日志上；如果指定了空值，
则表示关闭调试输出。

```go
l := log.New(os.Stdout, "[SQL]", 0)
db.Debug(l)
db.Query("select * from users") // 输出到日志
db.Debug(nil)
db.Query("select * from groups") // 不会输出到日志
```
