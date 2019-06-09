// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package fetch_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2"
	"github.com/issue9/orm/v2/fetch"
	"github.com/issue9/orm/v2/internal/testconfig"
)

type FetchEmail struct {
	Email string `orm:"unique(unique_index);nullable;pk;len(100)"`

	Regdate int64 `orm:"-"`
}

type FetchUser struct {
	FetchEmail
	ID       int    `orm:"name(id);ai;"`
	Username string `orm:"index(username_index);len(20)"`
	Group    int    `orm:"name(group);fk(fk_group,{group},id)"`
}

func (u *FetchUser) Meta() string {
	return "name(user)"
}

type Log struct {
	ID      int        `orm:"name(id);ai"`
	Content string     `orm:"name(content);len(1024)"`
	Created int        `orm:"name(created)"`
	UID     int        `orm:"name(uid)"`
	User    *FetchUser `orm:"name(user);nullable"`
}

func (l *Log) Meta() string {
	return "name(logs)"
}

// AfterFetcher 接口
func (u *FetchEmail) AfterFetch() error {
	u.Regdate = time.Now().Unix()
	return nil
}

// 初始化一个 sql.DB(sqlite3)，方便后面的测试用例使用。
func initDB(a *assert.Assertion) *orm.DB {
	db := testconfig.NewDB(a)
	now := time.Now().Unix()

	a.NotError(db.MultCreate(&FetchUser{}, &Log{}))

	/* 插入数据 */
	tx, err := db.Begin()
	a.NotError(err).NotNil(tx)

	stmt, err := tx.Prepare("INSERT INTO #user(id, Email,Username,{group}) values(?, ?, ?, ?)")
	a.NotError(err).NotNil(stmt)
	for i := 0; i < 100; i++ {
		_, err = stmt.Exec(i, fmt.Sprintf("email-%d", i), fmt.Sprintf("username-%d", i), 1)
		a.NotError(err)
	}

	stmt, err = tx.Prepare("INSERT INTO #logs(id, created,content,uid) values(?, ?, ?, ?)")
	a.NotError(err).NotNil(stmt)
	for i := 1; i < 100; i++ {
		_, err = stmt.Exec(i, now, fmt.Sprintf("content-%d", i), i)
		a.NotError(err)
	}
	a.NotError(tx.Commit())
	a.NotError(stmt.Close())

	return db
}

func TestObject_strict(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer testconfig.CloseDB(db, a)

	sql := `SELECT id,Email FROM #user WHERE id<2 ORDER BY id`
	now := time.Now().Unix()

	// test1:objs 的长度与导出的数据长度相等
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)

	objs := []*FetchUser{
		&FetchUser{},
		&FetchUser{},
	}
	cnt, err := fetch.Object(true, rows, &objs)
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
	cnt, err = fetch.Object(true, rows, &objs)
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
	cnt, err = fetch.Object(true, rows, objs) // 非指针传递
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
	cnt, err = fetch.Object(true, rows, &objs)
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
	cnt, err = fetch.Object(true, rows, array)
	a.Error(err).Equal(cnt, 0) // 非指针传递，出错
	a.NotError(rows.Close())

	// test6:数组指针传递，不会增长数组长度。
	rows, err = db.Query(sql)
	array = [1]*FetchUser{
		&FetchUser{},
	}
	cnt, err = fetch.Object(true, rows, &array)
	a.NotError(err).NotEmpty(cnt)
	a.Equal([1]*FetchUser{
		&FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0", Regdate: now}},
	}, array)
	a.NotError(rows.Close())

	// test7:obj 为一个 struct 指针。
	rows, err = db.Query(sql)
	obj := FetchUser{}
	cnt, err = fetch.Object(true, rows, &obj)
	a.NotError(err).NotEmpty(cnt)
	a.Equal(FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0"}}, obj)
	a.NotError(rows.Close())

	// test8:obj 为一个 struct。这将返回错误信息
	rows, err = db.Query(sql)
	obj = FetchUser{}
	cnt, err = fetch.Object(true, rows, obj)
	a.Error(err).Empty(cnt)
	a.NotError(rows.Close())

	sql = `SELECT * FROM #user WHERE id<2 ORDER BY id`

	// test8: objs 的长度与导出的数据长度相等
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	objs = []*FetchUser{
		&FetchUser{},
		&FetchUser{},
	}
	cnt, err = fetch.Object(true, rows, &objs)
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
	defer testconfig.CloseDB(db, a)

	// 导出一条数据有对应的 logs，一条没有对应的 logs
	sql := `SELECT u.id,u.Email,l.id as lid FROM #user AS u LEFT JOIN #logs AS l ON l.uid=u.id WHERE u.id<2 ORDER BY u.id`
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
	cnt, err := fetch.Object(false, rows, &objs)
	a.NotError(err).NotEmpty(cnt)
	a.Equal([]*userlog{
		&userlog{FetchUser: &FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0", Regdate: now}}, LID: 0},
		&userlog{FetchUser: &FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}}, LID: 1},
	}, objs)
	a.NotError(rows.Close())

	// 严格模式将出错，有一条记录部分数据为 NULL
	cnt, err = fetch.Object(true, rows, &objs)
	a.Error(err).Equal(cnt, 0)

	// test2:objs 的长度小于导出数据的长度，objs 应该自动增加长度。
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)
	objs = []*userlog{
		&userlog{},
	}
	cnt, err = fetch.Object(false, rows, &objs)
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
	cnt, err = fetch.Object(false, rows, objs) // 非指针传递
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
	cnt, err = fetch.Object(false, rows, &objs)
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
	cnt, err = fetch.Object(false, rows, array)
	a.Error(err).Equal(cnt, 0) // 非指针传递，出错
	a.NotError(rows.Close())

	// test6:数组指针传递，不会增长数组长度。
	rows, err = db.Query(sql)
	array = [1]*userlog{
		&userlog{},
	}
	cnt, err = fetch.Object(false, rows, &array)
	a.NotError(err).NotEmpty(cnt)
	a.Equal([1]*userlog{
		&userlog{FetchUser: &FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0", Regdate: now}}, LID: 0},
	}, array)
	a.NotError(rows.Close())

	// test7:obj 为一个 struct 指针。
	rows, err = db.Query(sql)
	obj := userlog{}
	cnt, err = fetch.Object(false, rows, &obj)
	a.NotError(err).NotEmpty(cnt)
	a.Equal(userlog{FetchUser: &FetchUser{ID: 0, FetchEmail: FetchEmail{Email: "email-0", Regdate: now}}, LID: 0}, obj)
	a.NotError(rows.Close())

	// test8:obj 为一个 struct。这将返回错误信息
	rows, err = db.Query(sql)
	obj = userlog{}
	cnt, err = fetch.Object(false, rows, obj)
	a.Error(err).Empty(cnt)
	a.NotError(rows.Close())

	sql = `SELECT u.*,l.id AS lid FROM #user AS u LEFT JOIN #logs AS l on l.uid=u.id WHERE u.id<2 ORDER BY u.id`

	// test8: objs 的长度与导出的数据长度相等
	rows, err = db.Query(sql)
	a.NotError(err).NotNil(rows)

	objs = []*userlog{
		&userlog{},
		&userlog{},
	}
	cnt, err = fetch.Object(false, rows, &objs)
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
	defer testconfig.CloseDB(db, a)

	sql := `SELECT l.*,u.id as [user.id],u.username as [user.username]  FROM #logs AS l LEFT JOIN #user as u ON u.id=l.uid WHERE l.id<3 ORDER BY l.id`
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)
	objs := []*Log{
		&Log{User: &FetchUser{}},
	}
	cnt, err := fetch.Object(true, rows, &objs)
	a.NotError(err).Equal(cnt, len(objs))
	a.Equal(objs[0].User.ID, objs[0].UID)
	a.Equal(objs[1].User.ID, objs[1].UID)
}

func TestObjectNotFound(t *testing.T) {
	a := assert.New(t)
	db := initDB(a)
	defer testconfig.CloseDB(db, a)

	sql := `SELECT id,Email FROM #user WHERE id>100 ORDER BY id`

	// test1: 查询条件不满足，返回空数据
	rows, err := db.Query(sql)
	a.NotError(err).NotNil(rows)
	objs := []*FetchUser{
		&FetchUser{},
		&FetchUser{},
	}
	cnt, err := fetch.Object(true, rows, &objs)
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
	cnt, err = fetch.Object(true, rows, array)
	a.Error(err).Equal(0, cnt) // 非指针传递，出错
	a.NotError(rows.Close())
}
