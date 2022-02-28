// SPDX-License-Identifier: MIT

package dialect_test

import (
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/orm/v5/core"
	"github.com/issue9/orm/v5/dialect"
	"github.com/issue9/orm/v5/internal/sqltest"
	"github.com/issue9/orm/v5/internal/test"
	"github.com/issue9/orm/v5/sqlbuilder"
)

func TestMysql_VersionSQL(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, test.Mysql, test.Mariadb)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testDialectVersionSQL(t)
	})
}

func TestMysql_DropConstrainStmtHook(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, test.Mysql, test.Mariadb)
	defer suite.Close()

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
			PK("#info_pk", "uid").
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
			PK("test_pk").
			Exec()
		a.NotError(err)

		err = addStmt.Reset().Table("#info").
			PK("#info_pk", "uid").
			Exec()
		t.NotError(err)
	})
}

func TestMysql_DropIndexSQL(t *testing.T) {
	a := assert.New(t, false)

	suite := test.NewSuite(a, test.Mysql, test.Mariadb)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		qs, err := t.DB.Dialect().DropIndexSQL("tbl", "index_name")
		a.NotError(err).Equal(qs, "ALTER TABLE {tbl} DROP INDEX {index_name}")
	})
}

func TestMysql_TruncateTableSQL(t *testing.T) {
	a := assert.New(t, false)

	suite := test.NewSuite(a, test.Mysql, test.Mariadb)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		qs, err := t.DB.Dialect().TruncateTableSQL("tbl", "")
		a.NotError(err).Equal(qs, []string{"TRUNCATE TABLE {tbl}"})
	})
}

func TestMysql_CreateTableOptions(t *testing.T) {
	a := assert.New(t, false)
	builder := core.NewBuilder("")
	a.NotNil(builder)
	var m = dialect.Mysql("mysql_driver_name", "")

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
	a := assert.New(t, false)

	var data = []*sqlTypeTester{
		{ // col.PrimitiveType = auto
			col: &core.Column{PrimitiveType: core.Auto},
			err: true,
		},
		{
			col:     &core.Column{PrimitiveType: core.Int},
			SQLType: "BIGINT NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Int16,
				Default:       5,
			},
			SQLType: "mediumint NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Int32,
				HasDefault:    true,
				Default:       5,
			},
			SQLType: "INT NOT NULL DEFAULT 5",
		},
		{
			col:     &core.Column{PrimitiveType: core.Bool},
			SQLType: "BOOLEAN NOT NULL",
		},
		{
			col:     &core.Column{PrimitiveType: core.Time},
			SQLType: "DATETIME NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Time,
				Length:        []int{-1},
			},
			err: true,
		},
		{
			col: &core.Column{
				PrimitiveType: core.Time,
				Length:        []int{7},
			},
			err: true,
		},
		{
			col:     &core.Column{PrimitiveType: core.Uint16},
			SQLType: "MEDIUMINT UNSIGNED NOT NULL",
		},
		{
			col:     &core.Column{PrimitiveType: core.Int8},
			SQLType: "SMALLINT NOT NULL",
		},
		{
			col:     &core.Column{PrimitiveType: core.Bytes},
			SQLType: "BLOB NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Int,
				Length:        []int{5, 6},
			},
			SQLType: "BIGINT(5) NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.String,
				Length:        []int{5, 6},
			},
			SQLType: "VARCHAR(5) NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.String,
				Length:        []int{-1},
			},
			SQLType: "LONGTEXT NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Float32,
				Length:        []int{5, 6},
			},
			SQLType: "FLOAT NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Float64,
				Length:        []int{5, 7},
			},
			SQLType: "DOUBLE PRECISION NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Int64,
				Length:        []int{5},
			},
			SQLType: "BIGINT(5) NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.String,
				Length:        []int{5},
			},
			SQLType: "VARCHAR(5) NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Decimal,
				Length:        []int{5, 9},
			},
			SQLType: "decimal(5,9) NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Decimal,
				Length:        []int{5},
			},
			err: true,
		},
		{
			col:     &core.Column{PrimitiveType: core.String},
			SQLType: "LONGTEXT NOT NULL",
		},
		{
			col:     &core.Column{PrimitiveType: core.Bool},
			SQLType: "BOOLEAN NOT NULL",
		},
		{ // sql.RawBytes 会被转换成 []byte
			col:     &core.Column{PrimitiveType: core.Bytes},
			SQLType: "BLOB NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Int64,
				AI:            true,
			},
			SQLType: "BIGINT PRIMARY KEY AUTO_INCREMENT NOT NULL",
		},
		{
			col: &core.Column{
				PrimitiveType: core.Uint64,
				AI:            true,
			},
			SQLType: "BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT NOT NULL",
		},
	}

	testSQLType(a, dialect.Mysql("mysql_driver_name", ""), data)
}

func TestMysql_Types(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, test.Mysql, test.Mariadb)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testTypes(t)
	})
}

func TestMysql_TypesDefault(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a, test.Mysql, test.Mariadb)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		testTypesDefault(t)
	})
}
