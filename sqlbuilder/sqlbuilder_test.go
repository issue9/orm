// SPDX-License-Identifier: MIT

package sqlbuilder_test

import (
	"testing"
	"time"

	"github.com/issue9/orm/v5/core"
	"github.com/issue9/orm/v5/internal/test"
	"github.com/issue9/orm/v5/sqlbuilder"
)

func TestMain(m *testing.M) {
	test.Main(m)
}

// user 需要与 initDB 中的 users 表中的字段相同
type user struct {
	ID      int64  `orm:"name(id);ai"`
	Name    string `orm:"name(name);len(20)"`
	Age     int    `orm:"name(age)"`
	Version int64  `orm:"name(version);default(0)"`
}

func (u *user) ApplyModel(m *core.Model) error {
	m.Name = "users"
	return nil
}

func initDB(t *test.Driver) {
	t.Assertion.TB().Helper()

	creator := sqlbuilder.CreateTable(t.DB).
		Table("users").
		AutoIncrement("id", core.Int64).
		Column("name", core.String, false, false, false, nil, 20).
		Column("age", core.Int, false, true, false, nil).
		Column("version", core.Int64, false, false, true, 0).
		Unique("unique_users_id", "id")
	err := creator.Exec()
	t.NotError(err, "%s@%s", err, t.DriverName)

	creator.Reset().Table("info").
		Column("uid", core.Int64, false, false, false, nil).
		Column("tel", core.String, false, false, false, nil, 11).
		Column("nickname", core.String, false, false, false, nil, 20).
		Column("address", core.String, false, false, false, nil, 1024).
		Column("birthday", core.Time, false, false, true, time.Time{}).
		PK("info_pk", "tel", "nickname").
		ForeignKey("info_fk", "uid", "users", "id", "CASCADE", "CASCADE")
	err = creator.Exec()
	t.NotError(err)

	sql := sqlbuilder.Insert(t.DB).
		Columns("name", "age").
		Table("users").
		Values("1", 1).
		Values("2", 2)
	_, err = sql.Exec()
	t.NotError(err, "%s@%s", err, t.DriverName)

	stmt, err := sql.Prepare()
	t.NotError(err, "%s@%s", err, t.DriverName).
		NotNil(stmt, "not nil @%s", t.DriverName)

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

func clearDB(t *test.Driver) {
	err := sqlbuilder.DropTable(t.DB).
		Table("info"). // 需要先删除 info，info 的外键依赖 users
		Table("users").
		Exec()
	t.NotError(err)
}
