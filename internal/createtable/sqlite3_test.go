// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package createtable_test

import (
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/orm/v5/core"
	"github.com/issue9/orm/v5/internal/createtable"
	"github.com/issue9/orm/v5/internal/sqltest"
	"github.com/issue9/orm/v5/internal/test"
)

func TestMain(m *testing.M) {
	test.Main(m)
}

var sqlite3CreateTable = []string{`CREATE TABLE fk_table(
	id integer NOT NULL,
	PRIMARY KEY(id)
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
	CONSTRAINT users_pk PRIMARY KEY (id),
	CONSTRAINT u_user_xx1 UNIQUE (mobile,username),
	CONSTRAINT u_user_email1 UNIQUE (email,username),
	CONSTRAINT unique_id UNIQUE (id),
	CONSTRAINT xxx_fk FOREIGN KEY (id) REFERENCES fk_table (id),
	CONSTRAINT xxx CHECK(created > 0)
	)`,
	`create index index_user_mobile on usr(mobile)`,
	`create unique index index_user_unique_email_id on usr(email,id)`,
}

func TestTable_CreateTableSQL(t *testing.T) {
	a := assert.New(t, false)

	tbl := &createtable.Sqlite3Table{
		Columns: map[string]string{
			"id": "id integer not null",
		},
		Constraints: map[string]*createtable.Sqlite3Constraint{
			"users_pk": {
				Type: core.ConstraintPK,
				SQL:  "constraint users_pk primary key(id)",
			},
		},
	}

	query, err := tbl.CreateTableSQL("test")
	a.NotError(err)
	sqltest.Equal(a, query, `create table test( id integer not null,constraint users_pk primary key(id))`)
}

func TestParseSqlite3CreateTable(t *testing.T) {
	a := assert.New(t, false)

	suite := test.NewSuite(a, test.Sqlite3)

	suite.Run(func(t *test.Driver) {
		db := t.DB

		for _, query := range sqlite3CreateTable {
			_, err := db.Exec(query)
			t.NotError(err)
		}

		defer func() {
			_, err := db.Exec("DROP TABLE `usr`")
			t.NotError(err)

			_, err = db.Exec("DROP TABLE `fk_table`")
			t.NotError(err)
		}()

		table, err := createtable.ParseSqlite3CreateTable("usr", db)
		t.NotError(err).NotNil(table)

		t.Equal(len(table.Columns), 8)
		sqltest.Equal(a, table.Columns["id"], "id integer NOT NULL")
		sqltest.Equal(a, table.Columns["created"], "created integer NOT NULL")
		sqltest.Equal(a, table.Columns["nickname"], "nickname text NOT NULL")
		sqltest.Equal(a, table.Columns["state"], "state integer NOT NULL")
		sqltest.Equal(a, table.Columns["username"], "username text NOT NULL")
		sqltest.Equal(a, table.Columns["mobile"], "mobile text NOT NULL")
		sqltest.Equal(a, table.Columns["email"], "email text NOT NULL")
		sqltest.Equal(a, table.Columns["pwd"], "pwd text NOT NULL")
		t.Equal(len(table.Constraints), 6).
			Equal(table.Constraints["u_user_xx1"], &createtable.Sqlite3Constraint{
				Type: core.ConstraintUnique,
				SQL:  "CONSTRAINT u_user_xx1 UNIQUE (mobile,username)",
			}).
			Equal(table.Constraints["u_user_email1"], &createtable.Sqlite3Constraint{
				Type: core.ConstraintUnique,
				SQL:  "CONSTRAINT u_user_email1 UNIQUE (email,username)",
			}).
			Equal(table.Constraints["unique_id"], &createtable.Sqlite3Constraint{
				Type: core.ConstraintUnique,
				SQL:  "CONSTRAINT unique_id UNIQUE (id)",
			}).
			Equal(table.Constraints["xxx_fk"], &createtable.Sqlite3Constraint{
				Type: core.ConstraintFK,
				SQL:  "CONSTRAINT xxx_fk FOREIGN KEY (id) REFERENCES fk_table (id)",
			}).
			Equal(table.Constraints["xxx"], &createtable.Sqlite3Constraint{
				Type: core.ConstraintCheck,
				SQL:  "CONSTRAINT xxx CHECK(created > 0)",
			}).
			Equal(table.Constraints["users_pk"], &createtable.Sqlite3Constraint{
				Type: core.ConstraintPK,
				SQL:  "CONSTRAINT users_pk PRIMARY KEY (id)",
			}) // 主键约束名为固定值
		t.Equal(len(table.Indexes), 2).
			Equal(table.Indexes["index_user_mobile"], &createtable.Sqlite3Index{
				Type: core.IndexDefault,
				SQL:  "CREATE INDEX index_user_mobile on usr(mobile)",
			}).
			Equal(table.Indexes["index_user_unique_email_id"], &createtable.Sqlite3Index{
				Type: core.IndexDefault,
				SQL:  "CREATE UNIQUE INDEX index_user_unique_email_id on usr(email,id)",
			}) // sqlite 没有 unique
	})
}
