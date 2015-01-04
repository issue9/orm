orm [![Build Status](https://travis-ci.org/issue9/orm.svg?branch=master)](https://travis-ci.org/issue9/orm) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://github.com/issue9/orm/blob/master/LICENSE)
======

###开发中，勿用!


```go
type User struct {
    // 对应表中的id字段，为自增列，从0开始
    Id          int64      `orm:"name(id);ai(0);"`
    // 对应表中的first_name字段，为索引index_name的一部分
    FirstName   string     `orm:"name(first_name);index(index_name)"`
    LastName    string     `orm:"name(first_name);index(index_name)"`
}

// 创建User表
e.Create(&User{})

// 更新id为1的记录
e.Update(&User{Id:1,FirstName:"abc"})
e.Where("id=?", 1).Add("FirstName", "abc").Update()

// 删除id为1的记录
e.Delete(&User{Id:1})
e.Where("id=?").Delete(1)

// 插入数据
e.Insert(&User{FirstName:"abc"})
e.SQL().Columns("FirstName","LastName").Insert("firstName", "lastName")

// 查找数据
maps,err := e.Where("id<?", 5).FetchMaps()
```

### 安装

```shell
go get github.com/issue9/orm
```


### 文档

[![Go Walker](http://gowalker.org/api/v1/badge)](http://gowalker.org/github.com/issue9/orm)
[![GoDoc](https://godoc.org/github.com/issue9/orm?status.svg)](https://godoc.org/github.com/issue9/orm)


### 版权

本项目采用[MIT](http://opensource.org/licenses/MIT)开源授权许可证，完整的授权说明可在LICENSE文件中找到。
