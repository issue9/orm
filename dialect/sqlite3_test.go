// SPDX-License-Identifier: MIT

package dialect_test

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v3/core"
	"github.com/issue9/orm/v3/dialect"
	"github.com/issue9/orm/v3/internal/sqltest"
	"github.com/issue9/orm/v3/internal/test"
	"github.com/issue9/orm/v3/sqlbuilder"
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
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testDialectVersionSQL(t)
	}, test.Sqlite3)
}

func TestSqlite3_AddConstraintStmtHook(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
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

	}, test.Sqlite3)
}

func TestSqlite3_DropConstraintStmtHook(t *testing.T) {
	a := assert.New(t)

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		db := t.DB

		for _, query := range sqlite3CreateTable {
			_, err := db.Exec(query)
			t.NotError(err)
		}

		defer clearSqlite3CreateTable(t, db)

		testDialectDropConstraintStmtHook(t)
	}, test.Sqlite3)
}

func TestSqlite3_DropColumnStmtHook(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
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
	}, test.Sqlite3)
}

func TestSqlite3_CreateTableOptions(t *testing.T) {
	a := assert.New(t)
	builder := core.NewBuilder("")
	a.NotNil(builder)
	var s = dialect.Sqlite3("sqlite3", "sqlite3_driver")

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

func TestSqlite3_TruncateTableStmtHooker(t *testing.T) {
	a := assert.New(t)

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		hook, ok := t.DB.Dialect().(sqlbuilder.TruncateTableStmtHooker)
		a.True(ok).NotNil(hook)

		stmt := sqlbuilder.TruncateTable(t.DB).Table("tbl", "")
		a.NotNil(stmt)
		qs, err := hook.TruncateTableStmtHook(stmt)
		a.NotError(err).Equal(1, len(qs))
		sqltest.Equal(a, qs[0], "DELETE FROM {tbl}")

		stmt = sqlbuilder.TruncateTable(t.DB).Table("tbl", "id")
		a.NotNil(stmt)
		qs, err = hook.TruncateTableStmtHook(stmt)
		a.NotError(err).Equal(2, len(qs))
		sqltest.Equal(a, qs[0], "DELETE FROM {tbl}")
		sqltest.Equal(a, qs[1], "DELETE FROM SQLITE_SEQUENCE WHERE name='tbl'")
	}, test.Sqlite3)
}

func TestSqlite3_SQLType(t *testing.T) {
	a := assert.New(t)

	var data = []*sqlTypeTester{
		{ // col == nil
			err: true,
		},
		{ // col.GoType == nil
			col: &core.Column{GoType: nil},
			err: true,
		},
		{
			col:     &core.Column{GoType: core.IntType},
			SQLType: "INTEGER NOT NULL",
		},
		{
			col:     &core.Column{GoType: core.NullBoolType},
			SQLType: "INTEGER NOT NULL",
		},
		{
			col:     &core.Column{GoType: core.BoolType},
			SQLType: "INTEGER NOT NULL",
		},
		{
			col:     &core.Column{GoType: reflect.TypeOf([]byte{'a', 'b'})},
			SQLType: "BLOB NOT NULL",
		},
		{
			col:     &core.Column{GoType: core.NullInt64Type},
			SQLType: "INTEGER NOT NULL",
		},
		{
			col:     &core.Column{GoType: core.NullFloat64Type},
			SQLType: "REAL NOT NULL",
		},
		{
			col:     &core.Column{GoType: core.NullStringType},
			SQLType: "TEXT NOT NULL",
		},
		{
			col: &core.Column{
				GoType:   core.NullStringType,
				Nullable: true,
			},
			SQLType: "TEXT",
		},
		{
			col: &core.Column{
				GoType:  core.NullStringType,
				Default: "123",
			},
			SQLType: "TEXT NOT NULL",
		},
		{
			col: &core.Column{
				GoType:     core.NullStringType,
				Default:    "123",
				HasDefault: true,
			},
			SQLType: "TEXT NOT NULL DEFAULT '123'",
		},
		{
			col: &core.Column{
				GoType: core.IntType,
				Length: []int{5, 6},
			},
			SQLType: "INTEGER NOT NULL",
		},
		{
			col: &core.Column{
				GoType: core.IntType,
				AI:     true,
			},
			SQLType: "INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL",
		},
		{
			col:     &core.Column{GoType: core.StringType},
			SQLType: "TEXT NOT NULL",
		},
		{
			col:     &core.Column{GoType: core.Float64Type},
			SQLType: "REAL NOT NULL",
		},
		{
			col:     &core.Column{GoType: core.NullInt64Type},
			SQLType: "INTEGER NOT NULL",
		},

		{
			col:     &core.Column{GoType: core.TimeType},
			SQLType: "TIMESTAMP NOT NULL",
		},

		{
			col: &core.Column{GoType: reflect.TypeOf(struct{}{})},
			err: true,
		},
	}

	testSQLType(a, dialect.Sqlite3("sqlite3", "sqlite3_driver"), data)
}

func TestSqlite3_SQLFormat(t *testing.T) {
	a := assert.New(t)
	now := time.Now().In(time.UTC)

	var data = []*sqlFormatTester{
		{
			v:      1,
			format: "1",
		},
		{
			v:      int8(1),
			format: "1",
		},

		// Bool
		{
			v:      true,
			format: "true",
		},
		{
			v:      false,
			format: "false",
		},

		// NullBool
		{
			v:      sql.NullBool{Valid: true, Bool: true},
			format: "true",
		},
		{
			v:      sql.NullBool{Valid: true, Bool: false},
			format: "false",
		},
		{
			v:      sql.NullBool{Valid: false, Bool: true},
			format: "NULL",
		},

		// NullInt64
		{
			v:      sql.NullInt64{Valid: true, Int64: 64},
			format: "64",
		},
		{
			v:      sql.NullInt64{Valid: true, Int64: -1},
			format: "-1",
		},
		{
			v:      sql.NullInt64{Valid: false, Int64: 64},
			format: "NULL",
		},

		// NullFloat64
		{
			v:      sql.NullFloat64{Valid: true, Float64: 6.4},
			format: "6.4",
		},
		{
			v:      sql.NullFloat64{Valid: true, Float64: -1.64},
			format: "-1.64",
		},
		{
			v:      sql.NullFloat64{Valid: false, Float64: 6.4},
			format: "NULL",
		},

		// NullString
		{
			v:      sql.NullString{Valid: true, String: "str"},
			format: "'str'",
		},
		{
			v:      sql.NullString{Valid: true, String: ""},
			format: "''",
		},
		{
			v:      sql.NullString{Valid: false, String: "str"},
			format: "NULL",
		},

		// time
		{ // sqlite3 不考虑长度信息
			v:      now,
			l:      []int{1, 5},
			format: "'" + now.Format("2006-01-02 15:04:05") + "'",
		},
		{ // sqlite3 不考虑长度信息
			v:      now,
			l:      []int{-1},
			format: "'" + now.Format("2006-01-02 15:04:05") + "'",
		},
		{
			v:      now,
			format: "'" + now.Format("2006-01-02 15:04:05") + "'",
		},
		{
			v:      now,
			l:      []int{1},
			format: "'" + now.Format("2006-01-02 15:04:05") + "'",
		},
		{
			v:      now,
			l:      []int{3},
			format: "'" + now.Format("2006-01-02 15:04:05") + "'",
		},
	}

	testSQLFormat(a, dialect.Sqlite3("sqlite3", "sqlite3_driver"), data)
}

func TestSqlite3_Types(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testTypes(t)
	}, test.Sqlite3)
}

func TestSqlite3_TypesDefault(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testTypesDefault(t)
	}, test.Sqlite3)
}
