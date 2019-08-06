// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package createtable

import (
	"testing"

	"github.com/issue9/assert"
)

func TestLines(t *testing.T) {
	a := assert.New(t)
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
