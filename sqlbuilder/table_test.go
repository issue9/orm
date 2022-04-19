// SPDX-License-Identifier: MIT

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/orm/v5/core"
	"github.com/issue9/orm/v5/internal/test"
	"github.com/issue9/orm/v5/sqlbuilder"
)

var (
	_ sqlbuilder.DDLStmt = &sqlbuilder.CreateTableStmt{}
	_ sqlbuilder.DDLStmt = &sqlbuilder.DropTableStmt{}
	_ sqlbuilder.DDLStmt = &sqlbuilder.TruncateTableStmt{}
)

func TestCreateTableStmt(t *testing.T) {
	a := assert.New(t, false)
	table := "create_table_test"
	suite := test.NewSuite(a)

	suite.Run(func(t *test.Driver) {
		stmt := sqlbuilder.CreateTable(t.DB).
			Table(table).
			AutoIncrement("id", core.Int).
			Column("age", core.Int, false, false, false, nil).
			Column("name", core.String, false, true, true, "", 100).
			Column("address", core.String, false, false, false, nil, 100).
			Index(core.IndexDefault, "index_index", "name", "address").
			Unique("u_age", "name", "address").
			Check("age_gt_0", "age>0")
		err := stmt.Exec()
		a.NotError(err)

		defer func() {
			err = sqlbuilder.DropTable(t.DB).
				Table(table).
				Exec()
			a.NotError(err)
		}()

		// AI 和 PK 同时指定为 ID
		err = stmt.Reset().
			Table("users").
			AutoIncrement("id", core.Int).
			PK("users_pk", "id").
			Err()
		t.Error(err)

		// 约束名重和昨
		err = stmt.Reset().Table("users").
			Column("name", core.String, false, false, false, nil).
			Unique("c1", "name").
			Check("c1", "name IS NOT NULL").
			Exec()
		a.Error(err)

		a.Error(stmt.Reset().Exec(), sqlbuilder.ErrTableIsEmpty)

		a.Error(stmt.Reset().Table("users").Exec(), sqlbuilder.ErrColumnsIsEmpty)

		insert := sqlbuilder.Insert(t.DB).
			Table(table).
			KeyValue("age", 1).
			KeyValue("name", "name1").
			KeyValue("address", "address1")
		rslt, err := insert.Exec()
		a.NotError(err).NotNil(rslt)

		prepare, err := insert.Prepare()
		a.NotError(err).NotNil(prepare)
		rslt, err = prepare.Exec(2, "name2", "address2")
		a.NotError(err).NotNil(rslt)
		rslt, err = prepare.Exec(3, "name3", "address3")
		a.NotError(err).NotNil(rslt)

		cnt, err := sqlbuilder.Select(t.DB).
			Count("count(*) as cnt").
			From(table).
			QueryInt("cnt")
		a.NotError(err).Equal(cnt, 3)
	})
}

func TestTruncateTable(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a)

	suite.Run(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		_, err := sqlbuilder.Insert(t.DB).
			Table("info").
			KeyValue("uid", 1).
			KeyValue("tel", "18011112222").
			KeyValue("nickname", "nickname1").
			KeyValue("address", "address1").
			Exec()
		a.NotError(err)

		truncate := sqlbuilder.TruncateTable(t.DB)
		err = truncate.Table("info", "").Exec()
		t.NotError(err)

		// 可重复调用
		err = truncate.Reset().Table("info", "").Exec()
		t.NotError(err)

		sel := sqlbuilder.Select(t.DB).
			Count("count(*) as cnt").
			From("info")
		rows, err := sel.Query()
		t.NotError(err).NotNil(rows)
		t.True(rows.Next())
		var val int
		t.NotError(rows.Scan(&val))
		t.NotError(rows.Close())
		t.Equal(val, 0)
	})
}

func TestDropTable(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a)

	suite.Run(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		drop := sqlbuilder.DropTable(t.DB)
		a.Error(drop.Exec())

		a.NotError(drop.Reset().Table("info").Exec())

		// 删除不存在的表
		a.NotError(drop.Reset().Table("not-exists").Exec())
	})
}
