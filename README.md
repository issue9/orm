orm [![Build Status](https://travis-ci.org/issue9/orm.svg?branch=master)](https://travis-ci.org/issue9/orm)
======

```go
type User struct {
    // 对应表中的 id 字段，为自增列，从 0 开始
    Id          int64      `orm:"name(id);ai(0);"`
    // 对应表中的 first_name 字段，为索引 index_name 的一部分
    FirstName   string     `orm:"name(first_name);index(index_name)"`
    LastName    string     `orm:"name(first_name);index(index_name)"`
}

// 创建 User 表
e.Create(&User{})

// 更新 id 为 1 的记录
e.Update(&User{Id:1,FirstName:"abc"})
e.Where("id=?", 1).Table("#tbl_name").Update(true, "FirstName", "abc")

// 删除 id 为 1 的记录
e.Delete(&User{Id:1})
e.Where("id=?", 1).Table("#tbl_name").Delete(true, []interface{}{"id":1})

// 插入数据
e.Insert(&User{FirstName:"abc"})

// 查找数据
maps,err := e.Where("id<?", 5).Table("#tbl_name").SelectMap(true, "*")
```

### 安装

```shell
go get github.com/issue9/orm
```


### 文档

[![Go Walker](http://gowalker.org/api/v1/badge)](http://gowalker.org/github.com/issue9/orm)
[![GoDoc](https://godoc.org/github.com/issue9/orm?status.svg)](https://godoc.org/github.com/issue9/orm)


### 版权

本项目采用[MIT](http://opensource.org/licenses/MIT)开源授权许可证，完整的授权说明可在[LICENSE](LICENSE)文件中找到。
