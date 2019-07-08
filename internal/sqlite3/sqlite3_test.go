// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlite3

import (
	"database/sql"
	"os"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/sqltest"
	"github.com/issue9/orm/v2/sqlbuilder"

	_ "github.com/mattn/go-sqlite3"
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

	tbl := &Table{
		Columns: map[string]string{
			"id": "id integer not null",
		},
		Constraints: map[string]*Constraint{
			"users_pk": &Constraint{
				Type: sqlbuilder.ConstraintPK,
				SQL:  "constraint users_pk primary key(id)",
			},
		},
	}

	query := tbl.CreateTableSQL("test")
	sqltest.Equal(a, query, `create table test( id integer not null,constraint users_pk primary key(id))`)
}

func TestParseSqlite3CreateTable(t *testing.T) {
	a := assert.New(t)
	dbFile := "./orm_test.db"

	db, err := sql.Open("sqlite3", dbFile)
	a.NotError(err).NotNil(db)
	defer func() {
		a.NotError(db.Close())
		a.NotError(os.Remove(dbFile))
	}()
	for _, query := range sqlite3CreateTable {
		_, err = db.Exec(query)
		a.NotError(err)
	}

	table, err := ParseCreateTable("usr", db)
	a.NotError(err).NotNil(table)

	a.Equal(len(table.Columns), 8)
	sqltest.Equal(a, table.Columns["id"], "id integer NOT NULL")
	sqltest.Equal(a, table.Columns["created"], "created integer NOT NULL")
	sqltest.Equal(a, table.Columns["nickname"], "nickname text NOT NULL")
	sqltest.Equal(a, table.Columns["state"], "state integer NOT NULL")
	sqltest.Equal(a, table.Columns["username"], "username text NOT NULL")
	sqltest.Equal(a, table.Columns["mobile"], "mobile text NOT NULL")
	sqltest.Equal(a, table.Columns["email"], "email text NOT NULL")
	sqltest.Equal(a, table.Columns["pwd"], "pwd text NOT NULL")
	a.Equal(len(table.Constraints), 6).
		Equal(table.Constraints["u_user_xx1"], &Constraint{
			Type: sqlbuilder.ConstraintUnique,
			SQL:  "CONSTRAINT u_user_xx1 UNIQUE (mobile,username)",
		}).
		Equal(table.Constraints["u_user_email1"], &Constraint{
			Type: sqlbuilder.ConstraintUnique,
			SQL:  "CONSTRAINT u_user_email1 UNIQUE (email,username)",
		}).
		Equal(table.Constraints["unique_id"], &Constraint{
			Type: sqlbuilder.ConstraintUnique,
			SQL:  "CONSTRAINT unique_id UNIQUE (id)",
		}).
		Equal(table.Constraints["xxx_fk"], &Constraint{
			Type: sqlbuilder.ConstraintFK,
			SQL:  "CONSTRAINT xxx_fk FOREIGN KEY (id) REFERENCES fk_table (id)",
		}).
		Equal(table.Constraints["xxx"], &Constraint{
			Type: sqlbuilder.ConstraintCheck,
			SQL:  "CONSTRAINT xxx CHECK(created > 0)",
		}).
		Equal(table.Constraints["users_pk"], &Constraint{
			Type: sqlbuilder.ConstraintPK,
			SQL:  "CONSTRAINT users_pk PRIMARY KEY (id)",
		}) // 主键约束名为固定值
	a.Equal(len(table.Indexes), 2).
		Equal(table.Indexes["index_user_mobile"], &Index{
			Type: sqlbuilder.IndexDefault,
			SQL:  "CREATE INDEX index_user_mobile on usr(mobile)",
		}).
		Equal(table.Indexes["index_user_unique_email_id"], &Index{
			Type: sqlbuilder.IndexDefault,
			SQL:  "CREATE UNIQUE INDEX index_user_unique_email_id on usr(email,id)",
		}) // sqlite 没有 unique
}

func TestFilterCreateTableSQL(t *testing.T) {
	a := assert.New(t)
	query := `create table tb1(
	id int not null primary key,
	name string not null,
	constraint fk foreign key (name) references tab2(col1)
);charset=utf-8`
	a.Equal(filterCreateTableSQL(query), []string{
		"id int not null primary key",
		"name string not null",
		"constraint fk foreign key (name) references tab2(col1)",
	})

	query = "create table `tb1`(`id` int,`name` string,unique `fk`(`id`,`name`))"
	a.Equal(filterCreateTableSQL(query), []string{
		"id int",
		"name string",
		"unique fk(id,name)",
	})
}
