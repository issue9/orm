// SPDX-License-Identifier: MIT

package createtable_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v4/core"
	"github.com/issue9/orm/v4/internal/createtable"
	"github.com/issue9/orm/v4/internal/sqltest"
	"github.com/issue9/orm/v4/internal/test"
)

var mysqlCreateTable = []string{`CREATE TABLE fk_table (
	id BIGINT(20) NOT NULL,
	PRIMARY KEY(id)
	)`,

	`CREATE TABLE users (
	id BIGINT(20) NOT NULL,
	created BIGINT NOT NULL,
	nickname VARCHAR(50) NOT NULL,
	state SMALLINT(6) NOT NULL,
	username VARCHAR(50) NOT NULL,
	mobile VARCHAR(18) NOT NULL,
	email VARCHAR(50) NOT NULL,
	password VARCHAR(128) NOT NULL,
	PRIMARY KEY (id,username),
	UNIQUE KEY u_user_xx1 (mobile,username),
	UNIQUE KEY u_user_email1 (email,username),
	UNIQUE KEY unique_id (id),
	KEY index_user_mobile (mobile),
	CONSTRAINT xxx_fk FOREIGN KEY (id) REFERENCES fk_table (id),
	CONSTRAINT xxx CHECK(created > 0)
	)`,
}

func TestParseMysqlCreateTable(t *testing.T) {
	a := assert.New(t)

	suite := test.NewSuite(a, test.Mysql)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		defer func() {
			_, err := t.DB.Exec("DROP TABLE users")
			a.NotError(err)

			_, err = t.DB.Exec("DROP TABLE fk_table")
			a.NotError(err)
		}()

		for _, query := range mysqlCreateTable {
			_, err := t.DB.Exec(query)
			t.NotError(err)
		}

		table, err := createtable.ParseMysqlCreateTable("users", t.DB)
		t.NotError(err).NotNil(table)

		t.Equal(len(table.Columns), 8)
		sqltest.Equal(a, table.Columns["id"], "bigint NOT NULL")
		sqltest.Equal(a, table.Columns["created"], "bigint NOT NULL")
		sqltest.Equal(a, table.Columns["nickname"], "varchar(50) NOT NULL")
		sqltest.Equal(a, table.Columns["state"], "smallint NOT NULL")
		sqltest.Equal(a, table.Columns["username"], "varchar(50) NOT NULL")
		sqltest.Equal(a, table.Columns["mobile"], "varchar(18) NOT NULL")
		sqltest.Equal(a, table.Columns["email"], "varchar(50) NOT NULL")
		sqltest.Equal(a, table.Columns["password"], "varchar(128) NOT NULL")
		t.Equal(len(table.Constraints), 6).
			Equal(table.Constraints["u_user_xx1"], core.ConstraintUnique).
			Equal(table.Constraints["u_user_email1"], core.ConstraintUnique).
			Equal(table.Constraints["unique_id"], core.ConstraintUnique).
			Equal(table.Constraints["xxx_fk"], core.ConstraintFK).
			Equal(table.Constraints["xxx"], core.ConstraintCheck).
			Equal(table.Constraints["users_pk"], core.ConstraintPK) // 主键约束名为固定值
		t.Equal(len(table.Indexes), 1).
			Equal(table.Indexes["index_user_mobile"], core.IndexDefault)
	})
}

// mariadb 稍微和 mysql 有点不一样
func TestParseMysqlCreateTable_mariadb(t *testing.T) {
	a := assert.New(t)

	suite := test.NewSuite(a, test.Mariadb)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		defer func() {
			_, err := t.DB.Exec("DROP TABLE users")
			a.NotError(err)

			_, err = t.DB.Exec("DROP TABLE fk_table")
			a.NotError(err)
		}()

		for _, query := range mysqlCreateTable {
			_, err := t.DB.Exec(query)
			t.NotError(err)
		}

		table, err := createtable.ParseMysqlCreateTable("users", t.DB)
		t.NotError(err).NotNil(table)

		t.Equal(len(table.Columns), 8)
		sqltest.Equal(a, table.Columns["id"], "bigint(20) NOT NULL")
		sqltest.Equal(a, table.Columns["created"], "bigint(20) NOT NULL")
		sqltest.Equal(a, table.Columns["nickname"], "varchar(50) NOT NULL")
		sqltest.Equal(a, table.Columns["state"], "smallint(6) NOT NULL")
		sqltest.Equal(a, table.Columns["username"], "varchar(50) NOT NULL")
		sqltest.Equal(a, table.Columns["mobile"], "varchar(18) NOT NULL")
		sqltest.Equal(a, table.Columns["email"], "varchar(50) NOT NULL")
		sqltest.Equal(a, table.Columns["password"], "varchar(128) NOT NULL")
		t.Equal(len(table.Constraints), 6).
			Equal(table.Constraints["u_user_xx1"], core.ConstraintUnique).
			Equal(table.Constraints["u_user_email1"], core.ConstraintUnique).
			Equal(table.Constraints["unique_id"], core.ConstraintUnique).
			Equal(table.Constraints["xxx_fk"], core.ConstraintFK).
			Equal(table.Constraints["xxx"], core.ConstraintCheck).
			Equal(table.Constraints["users_pk"], core.ConstraintPK) // 主键约束名为固定值
		t.Equal(len(table.Indexes), 1).
			Equal(table.Indexes["index_user_mobile"], core.IndexDefault)
	})
}
