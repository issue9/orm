// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bufio"
	"database/sql"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/sqltest"
)

var _ bufio.SplitFunc = splitWithAS

func TestFillArgs(t *testing.T) {
	a := assert.New(t)

	var data = []*struct {
		query  string
		args   []interface{}
		output string
		err    bool
	}{
		{
			query:  "select * from tbl",
			args:   []interface{}{},
			output: "select * from tbl",
		},
		{
			query:  "select * from tbl where id=?",
			args:   []interface{}{1},
			output: "select * from tbl where id='1'",
		},
		{
			query:  "select * from tbl where id=? and name=?",
			args:   []interface{}{1, "n"},
			output: "select * from tbl where id='1' and name='n'",
		},
		{
			query:  "select * from tbl where id=? and name=@name",
			args:   []interface{}{1, sql.Named("name", "n")},
			output: "select * from tbl where id='1' and name='n'",
		},
		{
			query:  "select * from tbl where id=? and name=@name and age>?",
			args:   []interface{}{1, sql.Named("name", "n"), 18},
			output: "select * from tbl where id='1' and name='n' and age>'18'",
		},
		{ // 类型不匹配
			query: "select * from tbl where id=? and name=@name",
			args:  []interface{}{1, "n"},
			err:   true,
		},
		{ // 类型不匹配
			query: "select * from tbl where id=? and name=@name and age>@age",
			args:  []interface{}{1, "n", sql.Named("age", 18)},
			err:   true,
		},
		{ // 名称不存在
			query: "select * from tbl where id=? and name=@name",
			args:  []interface{}{1, sql.Named("not-exists", "n")},
			err:   true,
		},
	}

	for index, item := range data {
		output, err := fillArgs(item.query, item.args)
		if item.err {
			a.Error(err, "%s@%d", err, index).
				Empty(output)
			continue
		}

		a.NotError(err, "%s@%d", err, index)
		sqltest.Equal(a, output, item.output)
	}
}

func TestGetColumnName(t *testing.T) {
	a := assert.New(t)

	var data = []*struct {
		input  string
		output string
	}{
		{
			input:  "",
			output: "",
		},
		{
			input:  "table.*",
			output: "*",
		},
		{
			input:  "{table}.*",
			output: "*",
		},
		{
			input:  "{table}.{as}",
			output: "{as}",
		},
		{ // 多个 as
			input:  "table.{as} as {as}",
			output: "{as}",
		},
		{
			input:  "count({table}.*) as cnt",
			output: "{cnt}",
		},
		{ // 别名中包含 AS
			input:  "count({table}.*) as {col as name}",
			output: "{col as name}",
		},
		{
			input:  "count({table}.*) as {count\t  name}",
			output: "{count\t  name}",
		},
		{ // 采用 \t 分隔
			input:  "count({table}.*)\tas\tcnt",
			output: "{cnt}",
		},
		{ // 采用 \t、\n 混合
			input:  "count({table}.*)\tas\ncnt",
			output: "{cnt}",
		},
		{ // 采用 \t 与空格混合
			input:  "count({table}.*) \tas\t cnt",
			output: "{cnt}",
		},
		{
			input:  "sum(count({table}.*)) as cnt",
			output: "{cnt}",
		},
		{ // 整个内容作为列名
			input:  "count({table}.*)",
			output: "{count(table.*)}",
		},
		{
			input:  "sum(count({table}.*)) as cnt",
			output: "{cnt}",
		},
		{
			input:  "sum(count({table}.as)) as {as}",
			output: "{as}",
		},
		{
			input:  "{table}.{as} as {as}",
			output: "{as}",
		},
		{
			input:  "{table}.{as} as 列名1",
			output: "{列名1}",
		},
		{
			input:  "{table}.{as} as {列名1}",
			output: "{列名1}",
		},
	}

	for index, item := range data {
		col := getColumnName(item.input)
		a.Equal(col, item.output, "not equal @%d v1:%v,v2:%v", index, col, item.output)
	}
}
