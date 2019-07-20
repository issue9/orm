// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bufio"
	"testing"

	"github.com/issue9/assert"
)

var _ bufio.SplitFunc = splitWithAS

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
