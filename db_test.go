// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package orm_test

import (
	"bytes"
	"database/sql"
	"strings"
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/orm/v6"
	"github.com/issue9/orm/v6/internal/test"
)

func TestMain(m *testing.M) {
	test.Main(m)
}

func TestDB_LastInsertID(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		t.NotError(t.DB.Create(&User{}))
		defer func() {
			t.NotError(t.DB.Drop(&User{}))
		}()

		id, err := t.DB.LastInsertID(&User{Username: "1"})
		t.NotError(err).Equal(id, 1)

		id, err = t.DB.LastInsertID(&User{Username: "2"})
		t.NotError(err).Equal(id, 2)
	})
}

type defvalues struct {
	Name string `orm:"name(name);nullable"`
	Age  int    `orm:"name(age);default(-1)"`
}

func (v *defvalues) TableName() string { return "defvalues" }

func TestDB_InsertDefaultValues(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(d *test.Driver) {
		d.NotError(d.DB.Create(&defvalues{}))
		defer func() {
			d.NotError(d.DB.Drop(&defvalues{}))
		}()

		_, err := d.DB.Insert(&defvalues{})
		a.NotError(err)
		hasCount(d.DB, a, "defvalues", 1)

		_, err = d.DB.Insert(&defvalues{})
		a.NotError(err)
		hasCount(d.DB, a, "defvalues", 2)
	})
}

func TestDB_Update(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		initData(t)
		defer clearData(t)

		// update
		r, err := t.DB.Update(&UserInfo{
			UID:       1,
			FirstName: "firstName1",
			LastName:  "lastName1",
			Sex:       "sex1",
		})
		a.NotError(err)
		cnt, err := r.RowsAffected()
		t.NotError(err).
			Equal(1, cnt)

		r, err = t.DB.Update(&UserInfo{
			UID:       2,
			FirstName: "firstName2",
			LastName:  "lastName2",
			Sex:       "sex2",
		})
		t.NotError(err)
		cnt, err = r.RowsAffected()
		a.NotError(err).
			Equal(1, cnt)

		u1 := &UserInfo{UID: 1}
		found, err := t.DB.Select(u1)
		t.NotError(err).True(found)
		t.Equal(1, u1.UID).
			Equal("firstName1", u1.FirstName).
			Equal("lastName1", u1.LastName).
			Equal("sex1", u1.Sex)
		if t.DriverName == "mysql" {
			t.Equal(u1.Any, []byte("55"))
		} else {
			t.Equal(u1.Any, "55")
		}

		u2 := &UserInfo{LastName: "lastName2", FirstName: "firstName2"}
		found, err = t.DB.Select(u2)
		t.NotError(err).True(found)
		t.Equal(u2, &UserInfo{UID: 2, FirstName: "firstName2", LastName: "lastName2", Sex: "sex2"})
	})
}

func TestDB_Update_occ(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		initData(t)
		defer clearData(t)

		r, err := t.DB.Insert(&Account{
			UID:     1,
			Account: 1,
		})
		t.NotError(err).NotNil(r)

		// 正常更新
		r, err = t.DB.Update(&Account{
			UID:     1,
			Account: 2,
			Version: 0, // 如果线上数据为 0，则能正常更新
		})
		t.NotError(err).NotNil(r)
		cnt, err := r.RowsAffected()
		t.NotError(err).Equal(1, cnt)

		r, err = t.DB.Update(&Account{
			UID:     1,
			Account: 2,
			Version: 1, // 更新一次之后，应该变为 1，则值为 1 时能正常更新。
		})
		t.NotError(err).NotNil(r)
		cnt, err = r.RowsAffected()
		t.NotError(err).Equal(1, cnt)

		r, err = t.DB.Update(&Account{
			UID:     1,
			Account: 2,
			Version: 1, // 多次更新之后，肯定不为 1，则此次更新失败
		})
		t.NotError(err).NotNil(r)
		cnt, err = r.RowsAffected()
		t.NotError(err).Equal(0, cnt)

		obj := &Account{UID: 1}
		obj.Account = 100
		found, err := t.DB.Select(obj)
		t.NotError(err).True(found)
		r, err = t.DB.Update(obj)
		t.NotError(err).NotNil(r)
		cnt, err = r.RowsAffected()
		t.NotError(err).Equal(1, cnt)
	})
}

func TestDB_Update_error(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	// 多个唯一约束符合查询条件
	suite.Run(func(t *test.Driver) {
		initData(t)
		defer clearData(t)

		admin := &Admin{
			User: User{
				Username: "username1",
			},
			Email: "email1",
		}
		rslt, err := t.DB.Update(admin)
		t.ErrorString(err, "多个唯一约束 unique_admin_username").Nil(rslt)
	})
}

func TestDB_Delete(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		initData(t)
		defer clearData(t)

		// delete
		r, err := t.DB.Delete(&UserInfo{UID: 1})
		t.NotError(err)
		cnt, err := r.RowsAffected()
		t.NotError(err).
			Equal(cnt, 1)

		r, err = t.DB.Delete(
			&UserInfo{
				LastName:  "l2",
				FirstName: "f2",
			})
		t.NotError(err)
		cnt, err = r.RowsAffected()
		t.NotError(err).
			Equal(cnt, 1)

		r, err = t.DB.Delete(&Admin{Email: "email1"})
		t.NotError(err)
		cnt, err = r.RowsAffected()
		t.NotError(err).
			Equal(cnt, 1)

		hasCount(t.DB, t.Assertion, "user_info", 0)
		hasCount(t.DB, t.Assertion, "administrators", 0)

		// delete 并不会重置 ai 计数
		_, err = t.DB.Insert(&Admin{Group: 1, Email: "email1"})
		t.NotError(err)
		a1 := &Admin{Email: "email1"}
		found, err := t.DB.Select(a1)
		t.NotError(err).True(found)
		t.Equal(a1.ID, 2) // a1.ID 为一个自增列,不会在 delete 中被重置
	})
}

func TestDB_Truncate(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		initData(t)
		defer clearData(t)

		hasCount(t.DB, t.Assertion, "administrators", 1)

		// truncate 之后，会重置 AI
		t.NotError(t.DB.Truncate(&Admin{}))
		hasCount(t.DB, t.Assertion, "administrators", 0)

		_, err := t.DB.Insert(&Admin{Group: 1, Email: "email1", User: User{Username: "u1"}})
		t.NotError(err)
		_, err = t.DB.Insert(&Admin{Group: 1, Email: "email2", User: User{Username: "u2"}})
		t.NotError(err)

		a1 := &Admin{Email: "email1"}
		found, err := t.DB.Select(a1)
		t.NotError(err).True(found)
		t.Equal(1, a1.ID)

		a2 := &Admin{Email: "email2"}
		found, err = t.DB.Select(a2)
		t.NotError(err).True(found)
		t.Equal(2, a2.ID)
	})
}

func TestDB_Drop(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		initData(t)
		defer clearData(t)

		t.NotError(t.DB.Drop(&UserInfo{}))
		t.NotError(t.DB.Drop(&Admin{}))
		r, err := t.DB.Insert(&Admin{})
		t.Error(err).Nil(r)
	})
}

func TestDB_Version(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		t.NotEmpty(t.DB.Version())
	})
}

func TestDB_Debug(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		buf := new(bytes.Buffer)

		t.DB.Debug(func(v string) { buf.WriteString(v) })
		_, err := t.DB.Query("select 1+1")
		t.NotError(err)
		t.DB.Debug(nil)
		_, err = t.DB.Query("select 2+2")
		t.NotError(err)

		t.True(strings.Contains(buf.String(), "select 1+1")).
			False(strings.Contains(buf.String(), "select 2+2"))
	})
}

func TestDB_Save(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		initData(t)
		defer clearData(t)

		// 指定的行存在，采用 update
		cnt, err := t.DB.SQLBuilder().Select().Column("count(*) AS cnt").From(orm.TableName(&Group{})).QueryInt("cnt")
		a.NotError(err).Equal(cnt, 1)
		lastid, isnew, err := t.DB.Save(&Group{ID: sql.NullInt64{Valid: true, Int64: 1}, Name: "save"})
		a.NotError(err).False(isnew).Zero(lastid)
		cnt, err = t.DB.SQLBuilder().Select().Column("count(*) AS cnt").From(orm.TableName(&Group{})).QueryInt("cnt")
		a.NotError(err).Equal(cnt, 1) // 没有增加行数

		// insert 不存在的行
		lastid, isnew, err = t.DB.Save(&Group{ID: sql.NullInt64{Valid: true, Int64: 2}, Name: "save"})
		a.Error(err)

		// insert 未指定自增列
		lastid, isnew, err = t.DB.Save(&Group{Name: "save"})
		a.NotError(err).True(isnew).Equal(lastid, 2)
		cnt, err = t.DB.SQLBuilder().Select().Column("count(*) AS cnt").From(orm.TableName(&Group{})).QueryInt("cnt")
		a.NotError(err).Equal(cnt, 2)
	})
}

func TestDB_New(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, "")

	suite.Run(func(t *test.Driver) {
		p1 := t.DB.New("p1_")
		p11 := p1.New("p11_")
		p2 := t.DB.New("p2_")

		a.Equal(p2, p2.New("p2_"))

		a.NotError(p1.Create(&User{}))
		a.NotError(p11.Create(&User{}))
		a.NotError(p2.Create(&User{}))
		defer func() {
			a.NotError(p1.Drop(&User{}))
			a.NotError(p11.Drop(&User{}))
			a.NotError(p2.Drop(&User{}))
		}()

		ids1, err := p1.LastInsertID(&User{Username: "1"})
		t.NotError(err).Equal(ids1, 1)

		ids2, err := p2.LastInsertID(&User{Username: "1"})
		t.NotError(err).Equal(ids2, 1)

		ids1, err = p1.LastInsertID(&User{Username: "2"})
		t.NotError(err).Equal(ids1, 2)

		ids2, err = p2.LastInsertID(&User{Username: "2"})
		t.NotError(err).Equal(ids2, 2)
	})
}
