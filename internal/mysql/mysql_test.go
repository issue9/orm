// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mysql_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/mysql"
	"github.com/issue9/orm/v2/internal/sqltest"
	"github.com/issue9/orm/v2/internal/test"
	"github.com/issue9/orm/v2/sqlbuilder"
)

var mysqlCreateTable = []string{`CREATE TABLE fk_table(
	id BIGINT(20) NOT NULL,
	PRIMARY KEY(id)
	)`,
	`CREATE TABLE users (
	id BIGINT(20) NOT NULL,
	created BIGINT(20) NOT NULL,
	nickname VARCHAR(50) NOT NULL,
	state SMALLINT(6) NOT NULL,
	username VARCHAR(50) NOT NULL,
	mobile VARCHAR(18) NOT NULL,
	email VARCHAR(200) NOT NULL,
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

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		defer func() {
			a.NotError(t.DB.Exec("DROP TABLE users"))
			a.NotError(t.DB.Exec("DROP TABLE fk_table"))
			a.NotError(t.DB.Close())
		}()
		for _, query := range mysqlCreateTable {
			_, err := t.DB.Exec(query)
			t.NotError(err)
		}

		table, err := mysql.ParseCreateTable("users", t.DB)
		t.NotError(err).NotNil(table)

		t.Equal(len(table.Columns), 8)
		sqltest.Equal(a, table.Columns["id"], "bigint(20) NOT NULL")
		sqltest.Equal(a, table.Columns["created"], "bigint(20) NOT NULL")
		sqltest.Equal(a, table.Columns["nickname"], "varchar(50) NOT NULL")
		sqltest.Equal(a, table.Columns["state"], "smallint(6) NOT NULL")
		sqltest.Equal(a, table.Columns["username"], "varchar(50) NOT NULL")
		sqltest.Equal(a, table.Columns["mobile"], "varchar(18) NOT NULL")
		sqltest.Equal(a, table.Columns["email"], "varchar(200) NOT NULL")
		sqltest.Equal(a, table.Columns["password"], "varchar(128) NOT NULL")
		t.Equal(len(table.Constraints), 6).
			Equal(table.Constraints["u_user_xx1"], sqlbuilder.ConstraintUnique).
			Equal(table.Constraints["u_user_email1"], sqlbuilder.ConstraintUnique).
			Equal(table.Constraints["unique_id"], sqlbuilder.ConstraintUnique).
			Equal(table.Constraints["xxx_fk"], sqlbuilder.ConstraintFK).
			Equal(table.Constraints["xxx"], sqlbuilder.ConstraintCheck).
			Equal(table.Constraints["users_pk"], sqlbuilder.ConstraintPK) // 主键约束名为固定值
		t.Equal(len(table.Indexes), 1).
			Equal(table.Indexes["index_user_mobile"], sqlbuilder.IndexDefault)
	}, "mysql")
}
