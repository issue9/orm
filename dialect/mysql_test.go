// SPDX-License-Identifier: MIT

package dialect_test

import (
	"reflect"
	"testing"

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
	}, test.Mysql, test.Mariadb)
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
	}, test.Mysql, test.Mariadb)

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
	}, test.Mysql, test.Mariadb)
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
	}, test.Mysql, test.Mariadb)
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
	}, test.Mysql, test.Mariadb)
}

func TestMysql_CreateTableOptions(t *testing.T) {
	a := assert.New(t)
	builder := core.NewBuilder("")
	a.NotNil(builder)
	var m = dialect.Mysql("mysql_driver_name")

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

	testSQLType(a, dialect.Mysql("mysql_driver_name"), data)
}

func TestMysql_Types(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testTypes(t)
	}, test.Mysql, test.Mariadb)
}

func TestMysql_TypesDefault(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testTypesDefault(t)
	}, test.Mysql, test.Mariadb)
}
