// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package fetch

import (
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/issue9/assert"

	_ "github.com/mattn/go-sqlite3"
)

const testDBFile = "./test.db"

type FetchEmail struct {
	Email string `orm:"unique(unique_index);nullable;pk(pk_name)"`
}

type FetchUser struct {
	FetchEmail
	Id       int    `orm:"name(id);ai(1,2);"`
	Username string `orm:"index(index)"`
	Group    int    `orm:"name(group);fk(fk_group,group,id)"`

	Regdate int `orm:"-"`
}

func TestParseObj(t *testing.T) {
	a := assert.New(t)
	obj := &FetchUser{Id: 5}
	mapped := map[string]reflect.Value{}

	v := reflect.ValueOf(obj).Elem()
	a.True(v.IsValid())

	err := parseObj(v, &mapped)
	a.NotError(err).Equal(4, len(mapped), "长度不相等，导出元素为:[%v]", mapped)

	// 忽略的字段
	_, found := mapped["Regdate"]
	a.False(found)

	// 判断字段是否存在
	vi, found := mapped["id"]
	a.True(found).True(vi.IsValid())

	// 设置字段的值
	mapped["id"].Set(reflect.ValueOf(36))
	a.Equal(36, obj.Id)
	mapped["Email"].SetString("email")
	a.Equal("email", obj.Email)
	mapped["Username"].SetString("username")
	a.Equal("username", obj.Username)
	mapped["group"].SetInt(1)
	a.Equal(1, obj.Group)
}

// 初始化一个sql.DB(sqlite3)，方便后面的测试用例使用。
func initDB(a *assert.Assertion) *sql.DB {
	db, err := sql.Open("sqlite3", testDBFile)
	a.NotError(err).NotNil(db)

	/* 创建表 */
	sql := `create table user (
        id integer not null primary key, 
        Email text,
        Username text,
        [group] interger)`
	_, err = db.Exec(sql)
	a.NotError(err)

	/* 插入数据 */
	tx, err := db.Begin()
	a.NotError(err).NotNil(tx)

	stmt, err := tx.Prepare("insert into user(id, Email,Username,[group]) values(?, ?, ?, ?)")
	a.NotError(err).NotNil(stmt)

	for i := 0; i < 100; i++ {
		_, err = stmt.Exec(i, fmt.Sprintf("email-%d", i), fmt.Sprintf("username-%d", i), 1)
		a.NotError(err)
	}
	tx.Commit()
	stmt.Close()

	return db
}

// 关闭sql.DB(sqlite3)的数据库连结。
func closeDB(db *sql.DB, a *assert.Assertion) {
	a.NotError(db.Close()).
		NotError(os.Remove(testDBFile)).
		FileNotExists(testDBFile)
}

func TestObj(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer closeDB(db, a)

	sql := `SELECT id,Email FROM user WHERE id<2 ORDER BY id`

	// test1:objs的长度与导出的数据长度相等
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)

	objs := []*FetchUser{
		&FetchUser{},
		&FetchUser{},
	}

	a.NotError(Obj(&objs, rows))

	a.Equal([]*FetchUser{
		&FetchUser{Id: 0, FetchEmail: FetchEmail{Email: "email-0"}},
		&FetchUser{Id: 1, FetchEmail: FetchEmail{Email: "email-1"}},
	}, objs)
	a.NotError(rows.Close())

	// test2:objs的长度小于导出数据的长度，objs应该自动增加长度。
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)
	objs = []*FetchUser{
		&FetchUser{},
	}
	a.NotError(Obj(&objs, rows))
	a.Equal(len(objs), 2)
	a.Equal([]*FetchUser{
		&FetchUser{Id: 0, FetchEmail: FetchEmail{Email: "email-0"}},
		&FetchUser{Id: 1, FetchEmail: FetchEmail{Email: "email-1"}},
	}, objs)
	a.NotError(rows.Close())

	// test3:objs的长度小于导出数据的长度，objs不会增加长度。
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)
	objs = []*FetchUser{
		&FetchUser{},
	}
	a.NotError(Obj(objs, rows)) // 非指针传递
	a.Equal(len(objs), 1)
	a.Equal([]*FetchUser{
		&FetchUser{Id: 0, FetchEmail: FetchEmail{Email: "email-0"}},
	}, objs)
	a.NotError(rows.Close())

	// test4:objs的长度大于导出数据的长度。
	rows, err = db.Query(sql)
	objs = []*FetchUser{
		&FetchUser{},
		&FetchUser{},
		&FetchUser{},
	}
	a.NotError(Obj(&objs, rows))
	a.Equal([]*FetchUser{
		&FetchUser{Id: 0, FetchEmail: FetchEmail{Email: "email-0"}},
		&FetchUser{Id: 1, FetchEmail: FetchEmail{Email: "email-1"}},
		&FetchUser{},
	}, objs)
	a.NotError(rows.Close())

	// test5:非数组指针传递。
	rows, err = db.Query(sql)
	array := [1]*FetchUser{
		&FetchUser{},
	}
	a.Error(Obj(array, rows)) // 非指针传递，出错
	a.NotError(rows.Close())

	// test6:数组指针传递，不会增长数组长度。
	rows, err = db.Query(sql)
	array = [1]*FetchUser{
		&FetchUser{},
	}
	a.NotError(Obj(&array, rows))
	a.Equal([1]*FetchUser{
		&FetchUser{Id: 0, FetchEmail: FetchEmail{Email: "email-0"}},
	}, array)
	a.NotError(rows.Close())

	// test7:obj为一个struct指针。
	rows, err = db.Query(sql)
	obj := FetchUser{}
	a.NotError(Obj(&obj, rows))
	a.Equal(FetchUser{Id: 0, FetchEmail: FetchEmail{Email: "email-0"}}, obj)
	a.NotError(rows.Close())

	// test8:obj为一个struct。这将返回错误信息
	rows, err = db.Query(sql)
	obj = FetchUser{}
	a.Error(Obj(obj, rows))
	a.NotError(rows.Close())
}
