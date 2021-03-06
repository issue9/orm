// SPDX-License-Identifier: MIT

package fetch_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v3/core"
	"github.com/issue9/orm/v3/fetch"
	"github.com/issue9/orm/v3/internal/test"
)

type FetchEmail struct {
	Email string `orm:"name(email);unique(unique_index);nullable;len(100)"`

	Regdate int64 `orm:"-"`
}

type FetchUser struct {
	FetchEmail
	ID       int    `orm:"name(id);ai;"`
	Username string `orm:"name(username);index(username_index);len(20)"`
	Group    int    `orm:"name(group)"`
}

func (u *FetchUser) Meta() string {
	return "name(user)"
}

type Log struct {
	ID      int       `orm:"name(id);ai"`
	Content string    `orm:"name(content);len(1024)"`
	Created core.Unix `orm:"name(created)"`
	UID     int       `orm:"name(uid)"`
}

func (l *Log) Meta() string {
	return "name(logs)"
}

// AfterFetcher 接口
func (u *FetchEmail) AfterFetch() error {
	u.Regdate = time.Now().Unix()
	return nil
}

// 初始化一个 sql.DB，方便后面的测试用例使用。
func initDB(t *test.Driver) {
	t.NotError(t.DB.MultCreate(&FetchUser{}, &Log{}))

	/* 插入数据 */
	tx, err := t.DB.Begin()
	t.NotError(err).NotNil(tx)

	stmt, err := tx.Prepare("INSERT INTO #user(id,email,username,{group}) values(?, ?, ?, ?)")
	t.NotError(err).NotNil(stmt)
	for i := 1; i < 100; i++ { // 自增 ID 部分数据库不能为 0
		_, err = stmt.Exec(i, fmt.Sprintf("email-%d", i), fmt.Sprintf("username-%d", i), 1)
		t.NotError(err)
	}
	t.NotError(stmt.Close())

	stmt, err = tx.Prepare("INSERT INTO #logs(id, created,content,uid) values(?, ?, ?, ?)")
	t.NotError(err).NotNil(stmt)
	for i := 1; i < 100; i++ {
		_, err = stmt.Exec(i, core.Unix{Time: time.Now()}, fmt.Sprintf("content-%d", i), i)
		t.NotError(err)
	}
	t.NotError(stmt.Close())
	t.NotError(tx.Commit())
}

func clearDB(t *test.Driver) {
	t.NotError(t.DB.MultDrop(&FetchUser{}, &Log{}))
}

func TestObject_strict(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		db := t.DB

		sql := `SELECT id,email FROM #user WHERE id<3 ORDER BY id`
		now := time.Now().Unix()

		// test1:objs 的长度与导出的数据长度相等
		rows, err := db.Query(sql)
		t.NotError(err).NotNil(rows)

		objs := []*FetchUser{
			{},
			{},
		}
		cnt, err := fetch.Object(true, rows, &objs)
		t.NotError(err).NotEmpty(cnt)
		t.Equal([]*FetchUser{
			{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}},
			{ID: 2, FetchEmail: FetchEmail{Email: "email-2", Regdate: now}},
		}, objs)
		t.NotError(rows.Close())

		// test2:objs 的长度小于导出数据的长度，objs 应该自动增加长度。
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)
		objs = []*FetchUser{
			{},
		}
		cnt, err = fetch.Object(true, rows, &objs)
		t.NotError(err).Equal(len(objs), cnt)
		t.Equal([]*FetchUser{
			{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}},
			{ID: 2, FetchEmail: FetchEmail{Email: "email-2", Regdate: now}},
		}, objs)
		t.NotError(rows.Close())

		// test3:objs 的长度小于导出数据的长度，objs 不会增加长度。
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)
		objs = []*FetchUser{
			{},
		}
		cnt, err = fetch.Object(true, rows, objs) // 非指针传递
		t.NotError(err).Equal(len(objs), cnt)
		t.Equal([]*FetchUser{
			{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}},
		}, objs)
		t.NotError(rows.Close())

		// test4:objs 的长度大于导出数据的长度。
		rows, err = db.Query(sql)
		objs = []*FetchUser{
			{},
			{},
			{},
		}
		cnt, err = fetch.Object(true, rows, &objs)
		t.NotError(err).NotEmpty(cnt)
		t.Equal([]*FetchUser{
			{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}},
			{ID: 2, FetchEmail: FetchEmail{Email: "email-2", Regdate: now}},
			{},
		}, objs)
		t.NotError(rows.Close())

		// test5:非数组指针传递。
		rows, err = db.Query(sql)
		array := [1]*FetchUser{
			{},
		}
		cnt, err = fetch.Object(true, rows, array)
		t.Error(err).Equal(cnt, 0) // 非指针传递，出错
		t.NotError(rows.Close())

		// test6:数组指针传递，不会增长数组长度。
		rows, err = db.Query(sql)
		array = [1]*FetchUser{
			{},
		}
		cnt, err = fetch.Object(true, rows, &array)
		t.NotError(err).NotEmpty(cnt)
		t.Equal([1]*FetchUser{
			{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}},
		}, array)
		t.NotError(rows.Close())

		// test7:obj 为一个 struct 指针。
		rows, err = db.Query(sql)
		obj := FetchUser{}
		cnt, err = fetch.Object(true, rows, &obj)
		t.NotError(err).NotEmpty(cnt)
		t.Equal(FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1"}}, obj)
		t.NotError(rows.Close())

		// test8:obj 为一个 struct。这将返回错误信息
		rows, err = db.Query(sql)
		obj = FetchUser{}
		cnt, err = fetch.Object(true, rows, obj)
		t.Error(err).Empty(cnt)
		t.NotError(rows.Close())

		sql = `SELECT * FROM #user WHERE id<3 ORDER BY id`

		// test8: objs 的长度与导出的数据长度相等
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		objs = []*FetchUser{
			{},
			{},
		}
		cnt, err = fetch.Object(true, rows, &objs)
		t.NotError(err).NotEmpty(cnt)
		t.Equal([]*FetchUser{
			{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}, Username: "username-1", Group: 1},
			{ID: 2, FetchEmail: FetchEmail{Email: "email-2", Regdate: now}, Username: "username-2", Group: 1},
		}, objs)
		t.NotError(rows.Close())
	})
}

func TestObject_no_strict(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)
		db := t.DB

		// 导出一条数据有对应的 logs，一条没有对应的 logs
		sql := `SELECT u.id,u.email,l.id as lid FROM #user AS u LEFT JOIN #logs AS l ON l.uid=u.id WHERE u.id<3 ORDER BY u.id`
		now := time.Now().Unix()

		type userlog struct {
			*FetchUser
			LID int64 `orm:"name(lid)"`
		}

		// test1:objs 的长度与导出的数据长度相等
		rows, err := db.Query(sql)
		t.NotError(err).NotNil(rows)

		objs := []*userlog{
			{},
			{},
		}
		cnt, err := fetch.Object(false, rows, &objs)
		t.NotError(err).NotEmpty(cnt)
		t.Equal([]*userlog{
			{FetchUser: &FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}}, LID: 1},
			{FetchUser: &FetchUser{ID: 2, FetchEmail: FetchEmail{Email: "email-2", Regdate: now}}, LID: 2},
		}, objs)
		t.NotError(rows.Close())

		// 严格模式将出错，有一条记录部分数据为 NULL
		cnt, err = fetch.Object(true, rows, &objs)
		t.Error(err).Equal(cnt, 0)

		// test2:objs 的长度小于导出数据的长度，objs 应该自动增加长度。
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)
		objs = []*userlog{
			{},
		}
		cnt, err = fetch.Object(false, rows, &objs)
		t.NotError(err).Equal(len(objs), cnt)
		t.Equal([]*userlog{
			{FetchUser: &FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}}, LID: 1},
			{FetchUser: &FetchUser{ID: 2, FetchEmail: FetchEmail{Email: "email-2", Regdate: now}}, LID: 2},
		}, objs)
		t.NotError(rows.Close())

		// test3:objs 的长度小于导出数据的长度，objs 不会增加长度。
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)
		objs = []*userlog{
			{},
		}
		cnt, err = fetch.Object(false, rows, objs) // 非指针传递
		t.NotError(err).Equal(len(objs), cnt)
		t.Equal([]*userlog{
			{FetchUser: &FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}}, LID: 1},
		}, objs)
		t.NotError(rows.Close())

		// test4:objs 的长度大于导出数据的长度。
		rows, err = db.Query(sql)
		objs = []*userlog{
			{},
			{},
			{},
		}
		cnt, err = fetch.Object(false, rows, &objs)
		t.NotError(err).NotEmpty(cnt)
		t.Equal([]*userlog{
			{FetchUser: &FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}}, LID: 1},
			{FetchUser: &FetchUser{ID: 2, FetchEmail: FetchEmail{Email: "email-2", Regdate: now}}, LID: 2},
			{},
		}, objs)
		t.NotError(rows.Close())

		// test5:非数组指针传递。
		rows, err = db.Query(sql)
		array := [1]*userlog{
			{},
		}
		cnt, err = fetch.Object(false, rows, array)
		t.Error(err).Equal(cnt, 0) // 非指针传递，出错
		t.NotError(rows.Close())

		// test6:数组指针传递，不会增长数组长度。
		rows, err = db.Query(sql)
		array = [1]*userlog{
			{},
		}
		cnt, err = fetch.Object(false, rows, &array)
		t.NotError(err).NotEmpty(cnt)
		t.Equal([1]*userlog{
			{FetchUser: &FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}}, LID: 1},
		}, array)
		t.NotError(rows.Close())

		// test7:obj 为一个 struct 指针。
		rows, err = db.Query(sql)
		obj := userlog{}
		cnt, err = fetch.Object(false, rows, &obj)
		t.NotError(err).NotEmpty(cnt)
		t.Equal(userlog{FetchUser: &FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}}, LID: 1}, obj)
		t.NotError(rows.Close())

		// test8:obj 为一个 struct。这将返回错误信息
		rows, err = db.Query(sql)
		obj = userlog{}
		cnt, err = fetch.Object(false, rows, obj)
		t.Error(err).Empty(cnt)
		t.NotError(rows.Close())

		sql = `SELECT u.*,l.id AS lid FROM #user AS u LEFT JOIN #logs AS l on l.uid=u.id WHERE u.id<3 ORDER BY u.id`

		// test8: objs 的长度与导出的数据长度相等
		rows, err = db.Query(sql)
		t.NotError(err).NotNil(rows)

		objs = []*userlog{
			{},
			{},
		}
		cnt, err = fetch.Object(false, rows, &objs)
		t.NotError(err).NotEmpty(cnt)
		t.Equal([]*userlog{
			{FetchUser: &FetchUser{ID: 1, FetchEmail: FetchEmail{Email: "email-1", Regdate: now}, Username: "username-1", Group: 1}, LID: 1},
			{FetchUser: &FetchUser{ID: 2, FetchEmail: FetchEmail{Email: "email-2", Regdate: now}, Username: "username-2", Group: 1}, LID: 2},
		}, objs)
		t.NotError(rows.Close())
	})
}

func TestObjectNest(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		type log struct {
			Log
			User *FetchUser `orm:"name(user)"`
		}

		sql := `SELECT l.*,u.id as {user.id},u.username as {user.username}  FROM #logs AS l LEFT JOIN #user as u ON u.id=l.uid WHERE l.id<3 ORDER BY l.id`
		rows, err := t.DB.Query(sql)
		t.NotError(err).NotNil(rows)
		objs := []*log{
			{},
		}
		cnt, err := fetch.Object(true, rows, &objs)
		t.NotError(err).Equal(cnt, len(objs))
		yestday := time.Now().Add(-24 * time.Hour)
		o0 := objs[0]
		o1 := objs[1]
		t.Equal(o0.User.ID, o0.UID).
			False(o0.Created.IsNull).True(o0.Created.After(yestday)) // Created 肯定是一个晚于 24 小时之前值
		t.Equal(o1.User.ID, o1.UID).
			False(o1.Created.IsNull).True(o1.Created.After(yestday))
	})
}

func TestObjectNotFound(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		sql := `SELECT id,email FROM #user WHERE id>100 ORDER BY id`

		// test1: 查询条件不满足，返回空数据
		rows, err := t.DB.Query(sql)
		t.NotError(err).NotNil(rows)
		objs := []*FetchUser{
			{},
			{},
		}
		cnt, err := fetch.Object(true, rows, &objs)
		t.NotError(err).Equal(cnt, 0)
		t.Equal([]*FetchUser{
			{},
			{},
		}, objs)
		t.NotError(rows.Close())

		// test2:非数组指针传递。
		rows, err = t.DB.Query(sql)
		array := [1]*FetchUser{
			{},
		}
		cnt, err = fetch.Object(true, rows, array)
		t.Error(err).Equal(0, cnt) // 非指针传递，出错
		t.NotError(rows.Close())
	})
}
