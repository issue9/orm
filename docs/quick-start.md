### 安装

```shell
go get github.com/issue9/orm
```

或是通过 go.mod 指定，推荐采用 go.mod 的方式自动安装。


### 快速开始

```go
import (
    "github.com/issue9/orm/v2"
    "github.com/issue9/orm/v2/dialect"

    _ "github.com/mattn/go-sqlite3"
)

type User struct {
    ID   int64  `orm:"name(id);ai" json:"id"`
    Name string `orm:"name(name);len(500)" json:"name"`
    Age  int    `orm:"name(age)" json:"age"`
}

// 指定了表名，以及其它一些表属性
func (u *User) Meta() string {
    return `name(users);mysql_charset(utf8)`
}

func main() {
    db,err := orm.New("sqlite3", "./test.db", "test_", dialect.Sqlite3())
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


