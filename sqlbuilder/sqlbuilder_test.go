// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"reflect"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/test"
	"github.com/issue9/orm/v2/sqlbuilder"
)

// user 需要与 initDB 中的 users 表中的字段相同
type user struct {
	ID      int64  `orm:"name(id);ai"`
	Name    string `orm:"name(name);len(20)"`
	Age     int    `orm:"name(age)"`
	Version int64  `orm:"name(version);default(0)"`
}

func (u *user) Meta() string {
	return "name(users)"
}

func initDB(t *test.Test) {
	db := t.DB.DB
	dialect := t.DB.Dialect()

	creator := sqlbuilder.CreateTable(db, dialect).
		Table("users").
		AutoIncrement("id", reflect.TypeOf(int64(1))).
		Column("name", reflect.TypeOf(""), false, false, nil, 20).
		Column("age", reflect.TypeOf(1), true, false, nil).
		Column("version", reflect.TypeOf(int64(1)), false, true, 0).
		Unique("unique_users_id", "id")
	err := creator.Exec()
	t.NotError(err, "%s@%s", err, t.DriverName)

	creator.Reset().Table("info").
		Column("uid", reflect.TypeOf(int64(1)), false, false, nil).
		Column("tel", reflect.TypeOf(""), false, false, nil, 11).
		Column("nickname", reflect.TypeOf(""), false, false, nil, 20).
		Column("address", reflect.TypeOf(""), false, false, nil, 1024).
		//Column("birthday", reflect.TypeOf(time.Time{}), false, true, time.Time{}).
		PK("tel", "nickname").
		ForeignKey("info_fk", "uid", "users", "id", "CASCADE", "CASCADE")
	err = creator.Exec()
	t.NotError(err)

	sql := sqlbuilder.Insert(db, dialect).
		Columns("name", "age").
		Table("users").
		Values("1", 1).
		Values("2", 2)
	_, err = sql.Exec()
	t.NotError(err, "%s@%s", err, t.DriverName)

	stmt, err := sql.Prepare()
	t.NotError(err, "%s@%s", err, t.DriverName).
		NotNil(stmt, "not nil @s", t.DriverName)

	_, err = stmt.Exec("3", 3, "4", 4)
	t.NotError(err, "%s@%s", err, t.DriverName)
	_, err = stmt.Exec("5", 6, "6", 6)
	t.NotError(err, "%s@%s", err, t.DriverName)

	sql.Reset()
	sql.Table("users").
		Columns("name").
		Values("7")
	id, err := sql.LastInsertID("users", "id")
	t.NotError(err, "%s@%s", err, t.DriverName).
		Equal(id, 7, "%d != %d @ %s", id, 7, t.DriverName)

	// 多行插入，不能拿到 lastInsertID
	sql.Table("users").
		Columns("name").
		Values("8").
		Values("9")
	id, err = sql.LastInsertID("users", "id")
	t.Error(err, "%s@%s", err, t.DriverName).
		Empty(id, "not empty @%s", t.DriverName)
}

func clearDB(t *test.Test) {
	err := sqlbuilder.DropTable(t.DB.DB, t.DB.Dialect()).
		Table("info"). // 需要先删除 info，info 的外键依赖 users
		Table("users").
		Exec()
	t.NotError(err)
}

func TestSQLBuilder(t *testing.T) {
	a := assert.New(t)

	b := sqlbuilder.New("")
	b.WriteBytes('1')
	b.WriteString("23")

	a.Equal("123", b.String())
	a.Equal(3, b.Len())

	b.Reset()
	a.Equal(b.String(), "")
	a.Equal(b.Len(), 0)

	b.WriteBytes('3').WriteString("21")
	a.Equal(b.String(), "321")

	b.TruncateLast(1)
	a.Equal(b.String(), "32").Equal(2, b.Len())
}
