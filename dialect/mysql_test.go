// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

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
var mysqlCreateTable = []string{`CREATE TABLE fk_table(
	id bigint NOT NULL,
	name varchar(20) not null,
	address varchar(200) not null,
	CONSTRAINT fk_table_pk PRIMARY KEY(id)
	)`,
	`CREATE TABLE usr (
	id bigint NOT NULL,
	created bigint NOT NULL,
	nickname varchar(20) NOT NULL,
	state bigint NOT NULL,
	username varchar(20) NOT NULL,
	mobile varchar(18) NOT NULL,
	email varchar(200) NOT NULL,
	pwd varchar(36) NOT NULL,
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

func TestMysql_VersionSQL(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testDialectVersionSQL(t)
	}, "mysql")
}

func TestMysql_DropConstrainStmtHook(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		db := t.DB

		for _, query := range mysqlCreateTable {
			_, err := db.Exec(query)
			t.NotError(err)
		}

		defer func() {
			_, err := db.Exec("DROP TABLE `usr`")
			t.NotError(err)

			_, err = db.Exec("DROP TABLE `fk_table`")
			t.NotError(err)
		}()

		testDialectDropConstraintStmtHook(t)
	}, "mysql")

	// 约束名不是根据 core.PKName() 生成的
	suite.ForEach(func(t *test.Driver) {
		db := t.DB

		query := "CREATE TABLE #info(uid BIGINT NOT NULL,CONSTRAINT test_pk PRIMARY KEY(uid))"
		_, err := db.Exec(query)
		t.NotError(err)

		defer func() {
			_, err = db.Exec("DROP TABLE #info")
			t.NotError(err)
		}()

		// 已经存在主键，出错
		addStmt := sqlbuilder.AddConstraint(t.DB)
		err = addStmt.Table("#info").
			PK("uid").
			Exec()
		t.Error(err)

		// 未指定 PK 属性，无法找到相同的约束名。
		err = sqlbuilder.DropConstraint(t.DB).
			Table("#info").
			Constraint("test_pk").
			Exec()
		a.Error(err)

		err = sqlbuilder.DropConstraint(t.DB).
			Table("#info").
			Constraint("test_pk").
			PK().
			Exec()
		a.NotError(err)

		err = addStmt.Reset().Table("#info").
			PK("uid").
			Exec()
		t.NotError(err)
	}, "mysql")
}

func TestMysql_DropIndexStmtHook(t *testing.T) {
	a := assert.New(t)

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		stmt := sqlbuilder.DropIndex(t.DB).Table("tbl").Name("index_name")
		a.NotNil(stmt)

		hook, ok := t.DB.Dialect().(sqlbuilder.DropIndexStmtHooker)
		a.True(ok).NotNil(hook)
		qs, err := hook.DropIndexStmtHook(stmt)
		a.NotError(err).Equal(qs, []string{"ALTER TABLE {tbl} DROP INDEX {index_name}"})
	}, "mysql")
}

func TestMysql_TruncateTableStmtHook(t *testing.T) {
	a := assert.New(t)

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		// mysql 不需要 ai 的相关设置
		stmt := sqlbuilder.TruncateTable(t.DB).Table("tbl", "")
		a.NotNil(stmt)

		hook, ok := t.DB.Dialect().(sqlbuilder.TruncateTableStmtHooker)
		a.True(ok).NotNil(hook)
		qs, err := hook.TruncateTableStmtHook(stmt)
		a.NotError(err).Equal(qs, []string{"TRUNCATE TABLE {tbl}"})
	}, "mysql")
}

func TestMysql_CreateTableOptions(t *testing.T) {
	a := assert.New(t)
	builder := core.NewBuilder("")
	a.NotNil(builder)
	var m = dialect.Mysql()

	// 空的 meta
	a.NotError(m.CreateTableOptionsSQL(builder, nil))
	a.Equal(builder.Len(), 0)

	// engine
	builder.Reset()
	a.NotError(m.CreateTableOptionsSQL(builder, map[string][]string{
		"mysql_engine":  {"innodb"},
		"mysql_charset": {"utf8"},
	}))
	a.True(builder.Len() > 0)
	query, err := builder.String()
	a.NotError(err)
	sqltest.Equal(a, query, "engine=innodb character set=utf8")
}

func TestMysql_SQLType(t *testing.T) {
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
			SQLType: "BIGINT NOT NULL",
		},
		{
			col: &core.Column{
				GoType:  core.Int16Type,
				Default: 5,
			},
			SQLType: "mediumint NOT NULL",
		},
		{
			col: &core.Column{
				GoType:     core.Int32Type,
				HasDefault: true,
				Default:    5,
			},
			SQLType: "INT NOT NULL DEFAULT 5",
		},
		{
			col:     &core.Column{GoType: core.BoolType},
			SQLType: "BOOLEAN NOT NULL",
		},
		{
			col:     &core.Column{GoType: core.TimeType},
			SQLType: "DATETIME NOT NULL",
		},
		{
			col: &core.Column{
				GoType: core.TimeType,
				Length: []int{-1},
			},
			err: true,
		},
		{
			col: &core.Column{
				GoType: core.TimeType,
				Length: []int{7},
			},
			err: true,
		},
		{
			col:     &core.Column{GoType: core.Uint16Type},
			SQLType: "MEDIUMINT UNSIGNED NOT NULL",
		},
		{
			col:     &core.Column{GoType: core.Int8Type},
			SQLType: "SMALLINT NOT NULL",
		},
		{
			col:     &core.Column{GoType: reflect.TypeOf([]byte{'a', 'b', 'c'})},
			SQLType: "BLOB NOT NULL",
		},
		{
			col: &core.Column{
				GoType: core.IntType,
				Length: []int{5, 6},
			},
			SQLType: "BIGINT(5) NOT NULL",
		},
		{
			col: &core.Column{
				GoType: core.StringType,
				Length: []int{5, 6},
			},
			SQLType: "VARCHAR(5) NOT NULL",
		},
		{
			col: &core.Column{
				GoType: core.StringType,
				Length: []int{-1},
			},
			SQLType: "LONGTEXT NOT NULL",
		},
		{
			col: &core.Column{
				GoType: core.Float32Type,
				Length: []int{5, 6},
			},
			SQLType: "DOUBLE(5,6) NOT NULL",
		},
		{
			col: &core.Column{
				GoType: core.Float64Type,
				Length: []int{5},
			},
			err: true,
		},
		{
			col: &core.Column{
				GoType: core.NullFloat64Type,
				Length: []int{5},
			},
			err: true,
		},
		{
			col: &core.Column{
				GoType: core.NullFloat64Type,
				Length: []int{5, 7},
			},
			SQLType: "DOUBLE(5,7) NOT NULL",
		},
		{
			col: &core.Column{
				GoType: core.NullInt64Type,
				Length: []int{5},
			},
			SQLType: "BIGINT(5) NOT NULL",
		},
		{
			col: &core.Column{
				GoType: core.NullStringType,
				Length: []int{5},
			},
			SQLType: "VARCHAR(5) NOT NULL",
		},
		{
			col:     &core.Column{GoType: core.NullStringType},
			SQLType: "LONGTEXT NOT NULL",
		},
		{
			col:     &core.Column{GoType: core.NullBoolType},
			SQLType: "BOOLEAN NOT NULL",
		},
		{ // sql.RawBytes 会被转换成 []byte
			col:     &core.Column{GoType: core.RawBytesType},
			SQLType: "BLOB NOT NULL",
		},
		{
			col: &core.Column{
				GoType: core.Int64Type,
				AI:     true,
			},
			SQLType: "BIGINT PRIMARY KEY AUTO_INCREMENT NOT NULL",
		},
		{
			col: &core.Column{
				GoType: core.Uint64Type,
				AI:     true,
			},
			SQLType: "BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT NOT NULL",
		},
		{
			col: &core.Column{GoType: reflect.TypeOf(struct{}{})},
			err: true,
		},
	}

	testSQLType(a, dialect.Mysql(), data)
}

func TestMysql_SQLFormat(t *testing.T) {
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
			format: "1",
		},
		{
			v:      false,
			format: "0",
		},

		// NullBool
		{
			v:      sql.NullBool{Valid: true, Bool: true},
			format: "1",
		},
		{
			v:      sql.NullBool{Valid: true, Bool: false},
			format: "0",
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
		{ // 长度错误
			v:   now,
			l:   []int{1, 2},
			err: true,
		},
		{ // 长度错误
			v:   now,
			l:   []int{600},
			err: true,
		},
		{
			v:      now,
			format: "'" + now.Format("2006-01-02 15:04:05") + "'",
		},
		{
			v:      now,
			l:      []int{1},
			format: "'" + now.Format("2006-01-02 15:04:05.9") + "'",
		},
		{
			v:      now,
			l:      []int{3},
			format: "'" + now.Format("2006-01-02 15:04:05.999") + "'",
		},
	}

	testSQLFormat(a, dialect.Mysql(), data)
}

func TestMysql_Types(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testTypes(t)
	}, "mysql")
}

func TestMysql_TypesDefault(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testTypesDefault(t)
	}, "mysql")
}
