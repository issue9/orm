// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package createtable

import (
	"testing"

	"github.com/issue9/assert/v4"
)

func TestLines(t *testing.T) {
	a := assert.New(t, false)
	query := `create table tb1(
	id int not null primary key,
	name string not null,
	constraint fk foreign key (name) references tab2(col1)
);charset=utf-8`
	a.Equal(lines(query), []string{
		"id int not null primary key",
		"name string not null",
		"constraint fk foreign key (name) references tab2(col1)",
	})

	query = "create table `tb1`(`id` int,`name` string,unique `fk`(`id`,`name`))"
	a.Equal(lines(query), []string{
		"id int",
		"name string",
		"unique fk(id,name)",
	})
}
