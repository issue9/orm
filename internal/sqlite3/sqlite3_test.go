// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlite3_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/core"
	"github.com/issue9/orm/v2/internal/sqlite3"
	"github.com/issue9/orm/v2/internal/sqltest"
	"github.com/issue9/orm/v2/internal/test"
)

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
	a := assert.New(t)

	tbl := &sqlite3.Table{
		Columns: map[string]string{
			"id": "id integer not null",
		},
		Constraints: map[string]*sqlite3.Constraint{
			"users_pk": {
				Type: core.ConstraintPK,
				SQL:  "constraint users_pk primary key(id)",
			},
		},
	}

	query := tbl.CreateTableSQL("test")
	sqltest.Equal(a, query, `create table test( id integer not null,constraint users_pk primary key(id))`)
}

func TestParseSqlite3CreateTable(t *testing.T) {
	a := assert.New(t)

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		db := t.DB

		for _, query := range sqlite3CreateTable {
			_, err := db.Exec(query)
			t.NotError(err)
		}

		table, err := sqlite3.ParseCreateTable("usr", db)
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
			Equal(table.Constraints["u_user_xx1"], &sqlite3.Constraint{
				Type: core.ConstraintUnique,
				SQL:  "CONSTRAINT u_user_xx1 UNIQUE (mobile,username)",
			}).
			Equal(table.Constraints["u_user_email1"], &sqlite3.Constraint{
				Type: core.ConstraintUnique,
				SQL:  "CONSTRAINT u_user_email1 UNIQUE (email,username)",
			}).
			Equal(table.Constraints["unique_id"], &sqlite3.Constraint{
				Type: core.ConstraintUnique,
				SQL:  "CONSTRAINT unique_id UNIQUE (id)",
			}).
			Equal(table.Constraints["xxx_fk"], &sqlite3.Constraint{
				Type: core.ConstraintFK,
				SQL:  "CONSTRAINT xxx_fk FOREIGN KEY (id) REFERENCES fk_table (id)",
			}).
			Equal(table.Constraints["xxx"], &sqlite3.Constraint{
				Type: core.ConstraintCheck,
				SQL:  "CONSTRAINT xxx CHECK(created > 0)",
			}).
			Equal(table.Constraints["users_pk"], &sqlite3.Constraint{
				Type: core.ConstraintPK,
				SQL:  "CONSTRAINT users_pk PRIMARY KEY (id)",
			}) // 主键约束名为固定值
		t.Equal(len(table.Indexes), 2).
			Equal(table.Indexes["index_user_mobile"], &sqlite3.Index{
				Type: core.IndexDefault,
				SQL:  "CREATE INDEX index_user_mobile on usr(mobile)",
			}).
			Equal(table.Indexes["index_user_unique_email_id"], &sqlite3.Index{
				Type: core.IndexDefault,
				SQL:  "CREATE UNIQUE INDEX index_user_unique_email_id on usr(email,id)",
			}) // sqlite 没有 unique
	}, "sqlite3")
}
