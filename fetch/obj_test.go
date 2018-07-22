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
	"time"

	"github.com/issue9/assert"

	_ "github.com/mattn/go-sqlite3"
)

const testDBFile = "./test.db"

type FetchEmail struct {
	Email string `orm:"unique(unique_index);nullable;pk"`
}

type FetchUser struct {
	FetchEmail
	ID       int    `orm:"name(id);ai(1,2);"`
	Username string `orm:"index(index)"`
	Group    int    `orm:"name(group);fk(fk_group,group,id)"`

	Regdate int64 `orm:"-"`
}

// AfterFetcher 接口
func (u *FetchUser) AfterFetch() error {
	u.Regdate = time.Now().Unix()
	return nil
}

func BenchmarkObj(b *testing.B) {
	a := assert.New(b)
	db := initDB(a)
	defer closeDB(db, a)

	sql := `SELECT id,Email FROM user WHERE id<2 ORDER BY id`
	objs := []*FetchUser{
		&FetchUser{},
		&FetchUser{},
	}

	for i := 0; i < b.N; i++ {
		rows, err := db.Query(sql)
		a.NotError(err)

		a.NotError(Obj(&objs, rows))
		rows.Close()
	}
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

func TestGetColumns(t *testing.T) {
	a := assert.New(t)
	obj := &FetchUser{}

	cols, err := getColumns(reflect.ValueOf(obj), []string{"id"})
	a.NotError(err).NotNil(cols)
	a.Equal(len(cols), 1)

	// 当列不存在数据模型时
	cols, err = getColumns(reflect.ValueOf(obj), []string{"id", "not-exists"})
	a.NotError(err).NotNil(cols)
	a.Equal(len(cols), 2)
}

func TestObj(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer closeDB(db, a)

	sql := `SELECT id,Email FROM user WHERE id<2 ORDER BY id`
	now := time.Now().Unix()

	// test1:objs的长度与导出的数据长度相等
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)

	objs := []*FetchUser{
		&FetchUser{},
		&FetchUser{},
	}
	a.NotError(Obj(&objs, rows))
	a.Equal([]*FetchUser{
		&FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0"}, Regdate: now},
		&FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1"}, Regdate: now},
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
		&FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0"}, Regdate: now},
		&FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1"}, Regdate: now},
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
		&FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0"}, Regdate: now},
	}, objs)
	a.NotError(rows.Close())

	// test4:objs 的长度大于导出数据的长度。
	rows, err = db.Query(sql)
	objs = []*FetchUser{
		&FetchUser{},
		&FetchUser{},
		&FetchUser{},
	}
	a.NotError(Obj(&objs, rows))
	a.Equal([]*FetchUser{
		&FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0"}, Regdate: now},
		&FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1"}, Regdate: now},
		&FetchUser{},
	}, objs)
	a.NotError(rows.Close())

	// test5:非数组指针传递。
	rows, err = db.Query(sql)
	array := [1]*FetchUser{
		&FetchUser{},
	}
	cnt, err := Obj(array, rows)
	a.Error(err).Equal(cnt, 0) // 非指针传递，出错
	a.NotError(rows.Close())

	// test6:数组指针传递，不会增长数组长度。
	rows, err = db.Query(sql)
	array = [1]*FetchUser{
		&FetchUser{},
	}
	a.NotError(Obj(&array, rows))
	a.Equal([1]*FetchUser{
		&FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0"}, Regdate: now},
	}, array)
	a.NotError(rows.Close())

	// test7:obj 为一个 struct 指针。
	rows, err = db.Query(sql)
	obj := FetchUser{}
	a.NotError(Obj(&obj, rows))
	a.Equal(FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0"}}, obj)
	a.NotError(rows.Close())

	// test8:obj 为一个 struct。这将返回错误信息
	rows, err = db.Query(sql)
	obj = FetchUser{}
	cnt, err = Obj(obj, rows)
	a.Error(err).Equal(0, cnt)
	a.NotError(rows.Close())

	sql = `SELECT * FROM user WHERE id<2 ORDER BY id`

	// test8: objs 的长度与导出的数据长度相等
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	objs = []*FetchUser{
		&FetchUser{},
		&FetchUser{},
	}
	a.NotError(Obj(&objs, rows))
	a.Equal([]*FetchUser{
		&FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0"}, Regdate: now, Username: "username-0", Group: 1},
		&FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1"}, Regdate: now, Username: "username-1", Group: 1},
	}, objs)
	a.NotError(rows.Close())
}

func TestObjNotFound(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer closeDB(db, a)

	sql := `SELECT id,Email FROM user WHERE id>100 ORDER BY id`

	// test1:
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)

	objs := []*FetchUser{
		&FetchUser{},
		&FetchUser{},
	}

	a.NotError(Obj(&objs, rows))

	a.Equal([]*FetchUser{
		&FetchUser{},
		&FetchUser{},
	}, objs)
	a.NotError(rows.Close())

	// test2:非数组指针传递。
	rows, err = db.Query(sql)
	array := [1]*FetchUser{
		&FetchUser{},
	}
	cnt, err := Obj(array, rows)
	a.Error(err).Equal(0, cnt) // 非指针传递，出错
	a.NotError(rows.Close())

	// test3:数组指针传递。
	rows, err = db.Query(sql)
	array = [1]*FetchUser{
		&FetchUser{},
	}
	a.NotError(Obj(&array, rows))
	a.Equal([1]*FetchUser{
		&FetchUser{},
	}, array)
	a.NotError(rows.Close())

	// test4:obj为一个struct指针。
	rows, err = db.Query(sql)
	obj := FetchUser{}
	a.NotError(Obj(&obj, rows))
	a.Equal(FetchUser{}, obj)
	a.NotError(rows.Close())
}
