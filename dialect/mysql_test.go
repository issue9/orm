// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect_test

import (
	"reflect"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/core"
	"github.com/issue9/orm/v2/dialect"
	"github.com/issue9/orm/v2/internal/sqltest"
	"github.com/issue9/orm/v2/internal/test"
	"github.com/issue9/orm/v2/sqlbuilder"
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

	suite.ForEach(func(t *test.Test) {
		testDialectVersionSQL(t)
	}, "mysql")
}

func TestMysql_DropConstrainStmtHook(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		db := t.DB

		for _, query := range mysqlCreateTable {
			_, err := db.Exec(query)
			t.NotError(err)
		}

		defer func() {
			_, err := db.Exec("DROP TABLE `usr`")
			a.NotError(err)

			_, err = db.Exec("DROP TABLE `fk_table`")
			a.NotError(err)
		}()

		testDialectDropConstraintStmtHook(t)
	}, "mysql")
}

func TestMysql_DropIndexStmtHook(t *testing.T) {
	a := assert.New(t)

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
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

	suite.ForEach(func(t *test.Test) {
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
	sqltest.Equal(a, builder.String(), "engine=innodb character set=utf8")
}

func TestMysql_SQLType(t *testing.T) {
	a := assert.New(t)

	var data = []*sqltypeTester{
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
			SQLType: "INT NOT NULL DEFAULT '5'",
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

func TestMysql_Types(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		testTypes(t)
	}, "mysql")
}
