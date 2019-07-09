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

func TestMysqlHooks(t *testing.T) {
	a := assert.New(t)

	_, ok := dialect.Mysql().(sqlbuilder.TruncateTableStmtHooker)
	a.True(ok)

	_, ok = dialect.Mysql().(sqlbuilder.DropIndexStmtHooker)
	a.True(ok)

	_, ok = dialect.Mysql().(sqlbuilder.DropConstraintStmtHooker)
	a.True(ok)
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
		db := t.DB.DB

		for _, query := range mysqlCreateTable {
			_, err := db.Exec(query)
			t.NotError(err)
		}

		defer func() {
			_, err := db.Exec("DROP TABLE usr")
			a.NotError(err)

			_, err = db.Exec("DROP TABLE fk_table")
			a.NotError(err)
		}()

		testDialectDropConstraintStmtHook(t)
	}, "mysql")
}

func TestMysql_DropIndexStmtHook(t *testing.T) {
	a := assert.New(t)
	my := dialect.Mysql()

	stmt := sqlbuilder.DropIndex(nil, my).Table("tbl").Name("index_name")
	a.NotNil(stmt)

	hook, ok := my.(sqlbuilder.DropIndexStmtHooker)
	a.True(ok).NotNil(hook)
	qs, err := hook.DropIndexStmtHook(stmt)
	a.NotError(err).Equal(qs, []string{"ALTER TABLE tbl DROP INDEX index_name"})
}

func TestMysql_TruncateTableStmtHook(t *testing.T) {
	a := assert.New(t)
	my := dialect.Mysql()

	// mysql 不需要 ai 的相关设置
	stmt := sqlbuilder.TruncateTable(nil, my).Table("tbl", "", "")
	a.NotNil(stmt)

	hook, ok := my.(sqlbuilder.TruncateTableStmtHooker)
	a.True(ok).NotNil(hook)
	qs, err := hook.TruncateTableStmtHook(stmt)
	a.NotError(err).Equal(qs, []string{"TRUNCATE TABLE tbl"})
}

func TestMysql_CreateTableOptions(t *testing.T) {
	a := assert.New(t)
	builder := sqlbuilder.New("")
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
			col: &sqlbuilder.Column{GoType: nil},
			err: true,
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(1)},
			SQLType: "BIGINT NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType:  reflect.TypeOf(1),
				Default: 5,
			},
			SQLType: "BIGINT NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType:     reflect.TypeOf(1),
				HasDefault: true,
				Default:    5,
			},
			SQLType: "BIGINT NOT NULL DEFAULT '5'",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(true)},
			SQLType: "BOOLEAN NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(time.Time{})},
			SQLType: "DATETIME NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(uint16(16))},
			SQLType: "MEDIUMINT UNSIGNED NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(int8(1))},
			SQLType: "SMALLINT NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf([]byte{'a', 'b', 'c'})},
			SQLType: "BLOB NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(int(1)),
				Length: []int{5, 6},
			},
			SQLType: "BIGINT(5) NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(""),
				Length: []int{5, 6},
			},
			SQLType: "VARCHAR(5) NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(""),
				Length: []int{-1},
			},
			SQLType: "LONGTEXT NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(1.2),
				Length: []int{5, 6},
			},
			SQLType: "DOUBLE(5,6) NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(1.2),
				Length: []int{5},
			},
			err: true,
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(sql.NullFloat64{}),
				Length: []int{5},
			},
			err: true,
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(sql.NullFloat64{}),
				Length: []int{5, 7},
			},
			SQLType: "DOUBLE(5,7) NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(sql.NullInt64{}),
				Length: []int{5},
			},
			SQLType: "BIGINT(5) NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(sql.NullString{}),
				Length: []int{5},
			},
			SQLType: "VARCHAR(5) NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.NullString{})},
			SQLType: "LONGTEXT NOT NULL",
		},
		{
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.NullBool{})},
			SQLType: "BOOLEAN NOT NULL",
		},
		{ // sql.RawBytes 会被转换成 []byte
			col:     &sqlbuilder.Column{GoType: reflect.TypeOf(sql.RawBytes{})},
			SQLType: "BLOB NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(int64(1)),
				AI:     true,
			},
			SQLType: "BIGINT PRIMARY KEY AUTO_INCREMENT NOT NULL",
		},
		{
			col: &sqlbuilder.Column{
				GoType: reflect.TypeOf(uint64(1)),
				AI:     true,
			},
			SQLType: "BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT NOT NULL",
		},
		{
			col: &sqlbuilder.Column{GoType: reflect.TypeOf(struct{}{})},
			err: true,
		},
	}

	testSQLType(a, dialect.Mysql(), data)
}
