// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package parser

import (
	"database/sql"
	"os"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/v2/sqlbuilder"

	"github.com/issue9/orm/v2/internal/sqltest"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
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
	CONSTRAINT xxx CHECK ((created > 0))
	)`,
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
	CONSTRAINT xxx CHECK ((created > 0))
	)`,
	`create index index_user_mobile on usr(mobile)`,
	`create unique index index_user_unique_email_id on usr(email,id)`,
}

func TestParseMysqlCreateTable(t *testing.T) {
	a := assert.New(t)

	db, err := sql.Open("mysql", "root@/orm_test")
	a.NotError(err).NotNil(db)
	defer func() {
		a.NotError(db.Exec("DROP TABLE users"))
		a.NotError(db.Exec("DROP TABLE fk_table"))
		a.NotError(db.Close())
	}()
	for _, query := range mysqlCreateTable {
		_, err = db.Exec(query)
		a.NotError(err)
	}

	table, err := ParseCreateTable("mysql", "users", db)
	a.NotError(err).NotNil(table)

	a.Equal(len(table.Columns), 8)
	sqltest.Equal(a, table.Columns["id"], "bigint(20) NOT NULL")
	sqltest.Equal(a, table.Columns["created"], "bigint(20) NOT NULL")
	sqltest.Equal(a, table.Columns["nickname"], "varchar(50) NOT NULL")
	sqltest.Equal(a, table.Columns["state"], "smallint(6) NOT NULL")
	sqltest.Equal(a, table.Columns["username"], "varchar(50) NOT NULL")
	sqltest.Equal(a, table.Columns["mobile"], "varchar(18) NOT NULL")
	sqltest.Equal(a, table.Columns["email"], "varchar(200) NOT NULL")
	sqltest.Equal(a, table.Columns["password"], "varchar(128) NOT NULL")
	a.Equal(len(table.Constraints), 6).
		Equal(table.Constraints["u_user_xx1"], sqlbuilder.ConstraintUnique).
		Equal(table.Constraints["u_user_email1"], sqlbuilder.ConstraintUnique).
		Equal(table.Constraints["unique_id"], sqlbuilder.ConstraintUnique).
		Equal(table.Constraints["xxx_fk"], sqlbuilder.ConstraintFK).
		Equal(table.Constraints["xxx"], sqlbuilder.ConstraintCheck).
		Equal(table.Constraints["users_pk"], sqlbuilder.ConstraintPK) // 主键约束名为固定值
	a.Equal(len(table.Indexes), 1).
		Equal(table.Indexes["index_user_mobile"], sqlbuilder.IndexDefault)
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

	table, err := ParseCreateTable("sqlite3", "usr", db)
	a.NotError(err).NotNil(table)

	a.Equal(len(table.Columns), 8)
	sqltest.Equal(a, table.Columns["id"], "integer NOT NULL")
	sqltest.Equal(a, table.Columns["created"], "integer NOT NULL")
	sqltest.Equal(a, table.Columns["nickname"], "text NOT NULL")
	sqltest.Equal(a, table.Columns["state"], "integer NOT NULL")
	sqltest.Equal(a, table.Columns["username"], "text NOT NULL")
	sqltest.Equal(a, table.Columns["mobile"], "text NOT NULL")
	sqltest.Equal(a, table.Columns["email"], "text NOT NULL")
	sqltest.Equal(a, table.Columns["pwd"], "text NOT NULL")
	a.Equal(len(table.Constraints), 6).
		Equal(table.Constraints["u_user_xx1"], sqlbuilder.ConstraintUnique).
		Equal(table.Constraints["u_user_email1"], sqlbuilder.ConstraintUnique).
		Equal(table.Constraints["unique_id"], sqlbuilder.ConstraintUnique).
		Equal(table.Constraints["xxx_fk"], sqlbuilder.ConstraintFK).
		Equal(table.Constraints["xxx"], sqlbuilder.ConstraintCheck).
		Equal(table.Constraints["users_pk"], sqlbuilder.ConstraintPK) // 主键约束名为固定值
	a.Equal(len(table.Indexes), 2).
		Equal(table.Indexes["index_user_mobile"], sqlbuilder.IndexDefault).
		Equal(table.Indexes["index_user_unique_email_id"], sqlbuilder.IndexDefault) // sqlite 没有 unique
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
