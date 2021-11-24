// SPDX-License-Identifier: MIT

package dialect_test

import (
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/orm/v4/core"
	"github.com/issue9/orm/v4/dialect"
	"github.com/issue9/orm/v4/internal/sqltest"
	"github.com/issue9/orm/v4/internal/test"
	"github.com/issue9/orm/v4/sqlbuilder"
)

// 创建测试数据表的脚本
var sqlite3CreateTable = []string{`CREATE TABLE fk_table(
	id integer NOT NULL,
	name text not null,
	address text not null,
	constraint fk_table_pk PRIMARY KEY(id)
	)`,
	`CREATE TABLE usr (
	id integer NOT NULL,
	created integer NOT NULL,
	nickname text NOT NULL,
	state integer NOT NULL,
	username text NOT NULL,
	mobile text NOT NULL,
	email text NOT NULL,
	pwd text NOT NULL,
	CONSTRAINT usr_pk PRIMARY KEY (id),
	CONSTRAINT u_user_xx1 UNIQUE (mobile,username),
	CONSTRAINT u_user_email1 UNIQUE (email,username),
	CONSTRAINT unique_id UNIQUE (id),
	CONSTRAINT xxx_fk FOREIGN KEY (id) REFERENCES fk_table (id),
	CONSTRAINT xxx CHECK (created > 0)
	)`,
	`create index index_user_mobile on usr(mobile)`,
	`create unique index index_user_unique_email_id on usr(email,id)`,
}

func clearSqlite3CreateTable(t *test.Driver, db core.Engine) {
	_, err := db.Exec("DROP TABLE `usr`")
	t.NotError(err)

	_, err = db.Exec("DROP TABLE `fk_table`")
	t.NotError(err)
}

func TestSqlite3_VersionSQL(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, test.Sqlite3)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testDialectVersionSQL(t)
	})
}

func TestSqlite3_AddConstraintStmtHook(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, test.Sqlite3)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		db := t.DB

		for _, query := range sqlite3CreateTable {
			_, err := db.Exec(query)
			t.NotError(err)
		}

		defer clearSqlite3CreateTable(t, db)

		// check 约束
		err := sqlbuilder.AddConstraint(db).
			Table("fk_table").
			Check("id_great_zero", "id>0").
			Exec()
		t.NotError(err)

	})
}

func TestSqlite3_DropConstraintStmtHook(t *testing.T) {
	a := assert.New(t, false)

	suite := test.NewSuite(a, test.Sqlite3)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		db := t.DB

		for _, query := range sqlite3CreateTable {
			_, err := db.Exec(query)
			t.NotError(err)
		}

		defer clearSqlite3CreateTable(t, db)

		testDialectDropConstraintStmtHook(t)
	})
}

func TestSqlite3_DropColumnStmtHook(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, test.Sqlite3)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		db := t.DB

		for _, query := range sqlite3CreateTable {
			_, err := db.Exec(query)
			t.NotError(err)
		}

		defer clearSqlite3CreateTable(t, db)

		err := sqlbuilder.DropColumn(db).
			Table("usr").
			Column("state").
			Exec()
		t.NotError(err)

		// 查询删除的列会出错
		_, err = db.Query("SELECT state FROM usr")
		t.Error(err)
	})
}

func TestSqlite3_CreateTableOptions(t *testing.T) {
	a := assert.New(t, false)
	builder := core.NewBuilder("")
	a.NotNil(builder)
	var s = dialect.Sqlite3("sqlite3_driver", "")

	// 空的 meta
	a.NotError(s.CreateTableOptionsSQL(builder, nil))
	a.Equal(builder.Len(), 0)

	// engine
	builder.Reset()
	a.NotError(s.CreateTableOptionsSQL(builder, map[string][]string{
		"sqlite3_rowid": {"false"},
	}))
	a.True(builder.Len() > 0)
	query, err := builder.String()
	a.NotError(err)
	sqltest.Equal(a, query, "without rowid")

	builder.Reset()
	a.Error(s.CreateTableOptionsSQL(builder, map[string][]string{
		"sqlite3_rowid": {"false", "false"},
	}))
}

func TestSqlite3_TruncateTableSQL(t *testing.T) {
	a := assert.New(t, false)

	suite := test.NewSuite(a, test.Sqlite3)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		qs, err := t.DB.Dialect().TruncateTableSQL("tbl", "")
		a.NotError(err).Equal(1, len(qs))
		sqltest.Equal(a, qs[0], "DELETE FROM {tbl}")

		qs, err = t.DB.Dialect().TruncateTableSQL("tbl", "id")
		a.NotError(err).Equal(2, len(qs))
		sqltest.Equal(a, qs[0], "DELETE FROM {tbl}")
		sqltest.Equal(a, qs[1], "DELETE FROM SQLITE_SEQUENCE WHERE name='tbl'")
	})
}

func TestSqlite3_SQLType(t *testing.T) {
	a := assert.New(t, false)

	var data = []*sqlTypeTester{
		{ // col == nil
			err: true,
		},
		{ // col.PrimitiveType == nil
			col: &core.Column{PrimitiveType: core.Auto},
			err: true,
		},
		{
			col:     &core.Column{PrimitiveType: core.Int},
			SQLType: "INTEGER NOT NULL",
		},
		{
			col:     &core.Column{PrimitiveType: core.Bool},
			SQLType: "INTEGER NOT NULL",
		},
		{
			col:     &core.Column{PrimitiveType: core.Bool},
			SQLType: "INTEGER NOT NULL",
		},
		{
			col:     &core.Column{PrimitiveType: core.Bytes},
			SQLType: "BLOB NOT NULL",
		},
		{
			col:     &core.Column{PrimitiveType: core.Int64},
			SQLType: "INTEGER NOT NULL",
		},
		{
			col:     &core.Column{PrimitiveType: core.Float64},
			SQLType: "REAL NOT NULL",
		},
		{
			col:     &core.Column{PrimitiveType: core.String},
			SQLType: "TEXT NOT NULL",
		},
		{
			col:     &core.Column{PrimitiveType: core.Decimal},
			SQLType: "NUMERIC NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.String,
				Nullable:      true,
			},
			SQLType: "TEXT",
		},
		{
			col: &core.Column{
				PrimitiveType: core.String,
				Default:       "123",
			},
			SQLType: "TEXT NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.String,
				Default:       "123",
				HasDefault:    true,
			},
			SQLType: "TEXT NOT NULL DEFAULT '123'",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Int,
				Length:        []int{5, 6},
			},
			SQLType: "INTEGER NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Int,
				AI:            true,
			},
			SQLType: "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL",
		},
		{
			col:     &core.Column{PrimitiveType: core.String},
			SQLType: "TEXT NOT NULL",
		},
		{
			col:     &core.Column{PrimitiveType: core.Float64},
			SQLType: "REAL NOT NULL",
		},
		{
			col:     &core.Column{PrimitiveType: core.Int64},
			SQLType: "INTEGER NOT NULL",
		},

		{
			col:     &core.Column{PrimitiveType: core.Time},
			SQLType: "TIMESTAMP NOT NULL",
		},
	}

	testSQLType(a, dialect.Sqlite3("sqlite3_driver", ""), data)
}

func TestSqlite3_Types(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, test.Sqlite3)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testTypes(t)
	})
}

func TestSqlite3_TypesDefault(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, test.Sqlite3)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testTypesDefault(t)
	})
}
