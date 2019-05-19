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

	Regdate int64 `orm:"-"`
}

type FetchUser struct {
	FetchEmail
	ID       int    `orm:"name(id);ai(1,2);"`
	Username string `orm:"index(index)"`
	Group    int    `orm:"name(group);fk(fk_group,group,id)"`
}

type Log struct {
	ID      int        `orm:"name(id);ai"`
	Content string     `orm:"name(content);len(1024)"`
	Created int        `orm:"name(caeated)"`
	UID     int        `orm:"name(uid)"`
	User    *FetchUser `orm:"name(user)"`
}

// AfterFetcher 接口
func (u *FetchEmail) AfterFetch() error {
	u.Regdate = time.Now().Unix()
	return nil
}

// 初始化一个 sql.DB(sqlite3)，方便后面的测试用例使用。
func initDB(a *assert.Assertion) *sql.DB {
	db, err := sql.Open("sqlite3", testDBFile)
	a.NotError(err).NotNil(db)
	now := time.Now().Unix()

	/* 创建表 */
	sql := `create table user (
        id integer not null primary key,
        Email text,
        Username text,
        [group] interger)`
	_, err = db.Exec(sql)
	a.NotError(err)

	sql = `create table logs (
        id integer not null primary key,
        created text,
        content text,
        uid interger)`
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

	stmt, err = tx.Prepare("insert into logs(id, created,content,uid) values(?, ?, ?, ?)")
	a.NotError(err).NotNil(stmt)
	for i := 1; i < 100; i++ {
		_, err = stmt.Exec(i, now, fmt.Sprintf("content-%d", i), i)
		a.NotError(err)
	}
	tx.Commit()
	stmt.Close()

	return db
}

// 关闭 sql.DB(sqlite3) 的数据库连接。
func closeDB(db *sql.DB, a *assert.Assertion) {
	a.NotError(db.Close()).
		NotError(os.Remove(testDBFile)).
		FileNotExists(testDBFile)
}

func BenchmarkObject(b *testing.B) {
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

		cnt, err := Object(true, rows, &objs)
		a.NotError(err).NotEmpty(cnt)
		rows.Close()
	}
}

func TestParseObject(t *testing.T) {
	a := assert.New(t)
	obj := &Log{ID: 5}
	mapped := map[string]reflect.Value{}

	v := reflect.ValueOf(obj).Elem()
	a.True(v.IsValid())

	a.NotError(parseObject(v, &mapped))
	a.Equal(8, len(mapped), "长度不相等，导出元素为:[%v]", mapped)

	// 忽略的字段
	_, found := mapped["user.Regdate"]
	a.False(found)

	// 判断字段是否存在
	vi, found := mapped["id"]
	a.True(found).True(vi.IsValid())

	// 设置字段的值
	mapped["user.id"].Set(reflect.ValueOf(36))
	a.Equal(36, obj.User.ID)
	mapped["user.Email"].SetString("email")
	a.Equal("email", obj.User.Email)
	mapped["user.Username"].SetString("username")
	a.Equal("username", obj.User.Username)
	mapped["user.group"].SetInt(1)
	a.Equal(1, obj.User.Group)

	type m struct {
		*FetchEmail
		ID int
	}
	o := &m{ID: 5}
	mapped = map[string]reflect.Value{}
	v = reflect.ValueOf(o).Elem()
	a.NotError(parseObject(v, &mapped))
	a.Equal(2, len(mapped), "长度不相等，导出元素为:[%v]", mapped)
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

func TestObject_strict(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer closeDB(db, a)

	sql := `SELECT id,Email FROM user WHERE id<2 ORDER BY id`
	now := time.Now().Unix()

	// test1:objs 的长度与导出的数据长度相等
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)

	objs := []*FetchUser{
		&FetchUser{},
		&FetchUser{},
	}
	cnt, err := Object(true, rows, &objs)
	a.NotError(err).NotEmpty(cnt)
	a.Equal([]*FetchUser{
		&FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0", Regdate: now}},
		&FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}},
	}, objs)
	a.NotError(rows.Close())

	// test2:objs 的长度小于导出数据的长度，objs 应该自动增加长度。
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)
	objs = []*FetchUser{
		&FetchUser{},
	}
	cnt, err = Object(true, rows, &objs)
	a.NotError(err).Equal(len(objs), cnt)
	a.Equal([]*FetchUser{
		&FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0", Regdate: now}},
		&FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}},
	}, objs)
	a.NotError(rows.Close())

	// test3:objs 的长度小于导出数据的长度，objs 不会增加长度。
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)
	objs = []*FetchUser{
		&FetchUser{},
	}
	cnt, err = Object(true, rows, objs) // 非指针传递
	a.NotError(err).Equal(len(objs), cnt)
	a.Equal([]*FetchUser{
		&FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0", Regdate: now}},
	}, objs)
	a.NotError(rows.Close())

	// test4:objs 的长度大于导出数据的长度。
	rows, err = db.Query(sql)
	objs = []*FetchUser{
		&FetchUser{},
		&FetchUser{},
		&FetchUser{},
	}
	cnt, err = Object(true, rows, &objs)
	a.NotError(err).NotEmpty(cnt)
	a.Equal([]*FetchUser{
		&FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0", Regdate: now}},
		&FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}},
		&FetchUser{},
	}, objs)
	a.NotError(rows.Close())

	// test5:非数组指针传递。
	rows, err = db.Query(sql)
	array := [1]*FetchUser{
		&FetchUser{},
	}
	cnt, err = Object(true, rows, array)
	a.Error(err).Equal(cnt, 0) // 非指针传递，出错
	a.NotError(rows.Close())

	// test6:数组指针传递，不会增长数组长度。
	rows, err = db.Query(sql)
	array = [1]*FetchUser{
		&FetchUser{},
	}
	cnt, err = Object(true, rows, &array)
	a.NotError(err).NotEmpty(cnt)
	a.Equal([1]*FetchUser{
		&FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0", Regdate: now}},
	}, array)
	a.NotError(rows.Close())

	// test7:obj 为一个 struct 指针。
	rows, err = db.Query(sql)
	obj := FetchUser{}
	cnt, err = Object(true, rows, &obj)
	a.NotError(err).NotEmpty(cnt)
	a.Equal(FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0"}}, obj)
	a.NotError(rows.Close())

	// test8:obj 为一个 struct。这将返回错误信息
	rows, err = db.Query(sql)
	obj = FetchUser{}
	cnt, err = Object(true, rows, obj)
	a.Error(err).Empty(cnt)
	a.NotError(rows.Close())

	sql = `SELECT * FROM user WHERE id<2 ORDER BY id`

	// test8: objs 的长度与导出的数据长度相等
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	objs = []*FetchUser{
		&FetchUser{},
		&FetchUser{},
	}
	cnt, err = Object(true, rows, &objs)
	a.NotError(err).NotEmpty(cnt)
	a.Equal([]*FetchUser{
		&FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0", Regdate: now}, Username: "username-0", Group: 1},
		&FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}, Username: "username-1", Group: 1},
	}, objs)
	a.NotError(rows.Close())
}

func TestObject_no_strict(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer closeDB(db, a)

	// 导出一条数据有对应的 logs，一条没有对应的 logs
	sql := `SELECT u.id,u.Email,l.id as lid FROM user AS u LEFT JOIN logs AS l ON l.uid=u.id WHERE u.id<2 ORDER BY u.id`
	now := time.Now().Unix()

	type userlog struct {
		*FetchUser
		LID int64 `orm:"name(lid)"`
	}

	// test1:objs 的长度与导出的数据长度相等
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)

	objs := []*userlog{
		&userlog{},
		&userlog{},
	}
	cnt, err := Object(false, rows, &objs)
	a.NotError(err).NotEmpty(cnt)
	a.Equal([]*userlog{
		&userlog{FetchUser: &FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0", Regdate: now}}, LID: 0},
		&userlog{FetchUser: &FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}}, LID: 1},
	}, objs)
	a.NotError(rows.Close())

	// test2:objs 的长度小于导出数据的长度，objs 应该自动增加长度。
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)
	objs = []*userlog{
		&userlog{},
	}
	cnt, err = Object(false, rows, &objs)
	a.NotError(err).Equal(len(objs), cnt)
	a.Equal([]*userlog{
		&userlog{FetchUser: &FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0", Regdate: now}}, LID: 0},
		&userlog{FetchUser: &FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}}, LID: 1},
	}, objs)
	a.NotError(rows.Close())

	// test3:objs 的长度小于导出数据的长度，objs 不会增加长度。
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)
	objs = []*userlog{
		&userlog{},
	}
	cnt, err = Object(false, rows, objs) // 非指针传递
	a.NotError(err).Equal(len(objs), cnt)
	a.Equal([]*userlog{
		&userlog{FetchUser: &FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0", Regdate: now}}, LID: 0},
	}, objs)
	a.NotError(rows.Close())

	// test4:objs 的长度大于导出数据的长度。
	rows, err = db.Query(sql)
	objs = []*userlog{
		&userlog{},
		&userlog{},
		&userlog{},
	}
	cnt, err = Object(false, rows, &objs)
	a.NotError(err).NotEmpty(cnt)
	a.Equal([]*userlog{
		&userlog{FetchUser: &FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0", Regdate: now}}, LID: 0},
		&userlog{FetchUser: &FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}}, LID: 1},
		&userlog{},
	}, objs)
	a.NotError(rows.Close())

	// test5:非数组指针传递。
	rows, err = db.Query(sql)
	array := [1]*userlog{
		&userlog{},
	}
	cnt, err = Object(false, rows, array)
	a.Error(err).Equal(cnt, 0) // 非指针传递，出错
	a.NotError(rows.Close())

	// test6:数组指针传递，不会增长数组长度。
	rows, err = db.Query(sql)
	array = [1]*userlog{
		&userlog{},
	}
	cnt, err = Object(false, rows, &array)
	a.NotError(err).NotEmpty(cnt)
	a.Equal([1]*userlog{
		&userlog{FetchUser: &FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0", Regdate: now}}, LID: 0},
	}, array)
	a.NotError(rows.Close())

	// test7:obj 为一个 struct 指针。
	rows, err = db.Query(sql)
	obj := userlog{}
	cnt, err = Object(false, rows, &obj)
	a.NotError(err).NotEmpty(cnt)
	a.Equal(userlog{FetchUser: &FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0", Regdate: now}}, LID: 0}, obj)
	a.NotError(rows.Close())

	// test8:obj 为一个 struct。这将返回错误信息
	rows, err = db.Query(sql)
	obj = userlog{}
	cnt, err = Object(false, rows, obj)
	a.Error(err).Empty(cnt)
	a.NotError(rows.Close())

	sql = `SELECT u.*,l.id AS lid FROM user AS u LEFT JOIN logs AS l on l.uid=u.id WHERE u.id<2 ORDER BY u.id`

	// test8: objs 的长度与导出的数据长度相等
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	objs = []*userlog{
		&userlog{},
		&userlog{},
	}
	cnt, err = Object(false, rows, &objs)
	a.NotError(err).NotEmpty(cnt)
	a.Equal([]*userlog{
		&userlog{FetchUser: &FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0", Regdate: now}, Username: "username-0", Group: 1}, LID: 0},
		&userlog{FetchUser: &FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}, Username: "username-1", Group: 1}, LID: 1},
	}, objs)
	a.NotError(rows.Close())
}

func TestObjectNest(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer closeDB(db, a)

	sql := `SELECT l.*,u.id as [user.id],u.username as [user.username]  FROM logs AS l LEFT JOIN user as u ON u.id=l.uid WHERE l.id<3 ORDER BY l.id`
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)
	objs := []*Log{
		&Log{User: &FetchUser{}},
	}
	cnt, err := Object(true, rows, &objs)
	a.NotError(err).Equal(cnt, len(objs))
	a.Equal(objs[0].User.ID, objs[0].UID)
	a.Equal(objs[1].User.ID, objs[1].UID)
}

func TestObjectNotFound(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer closeDB(db, a)

	sql := `SELECT id,Email FROM user WHERE id>100 ORDER BY id`

	// test1: 查询条件不满足，返回空数据
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)
	objs := []*FetchUser{
		&FetchUser{},
		&FetchUser{},
	}
	cnt, err := Object(true, rows, &objs)
	a.NotError(err).Equal(cnt, 0)
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
	cnt, err = Object(true, rows, array)
	a.Error(err).Equal(0, cnt) // 非指针传递，出错
	a.NotError(rows.Close())
}
