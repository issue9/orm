// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"testing"

	"github.com/issue9/assert"
)

func TestSplitWithAS(t *testing.T) {
	a := assert.New(t)

	var data = []*struct {
		input  string
		output []string // 第一个元素为列名，第二个元素为别名
	}{
		{
			input:  "col as alias",
			output: []string{"col", "alias"},
		},
		{
			input:  "col As alias",
			output: []string{"col", "alias"},
		},
		{
			input:  "col AS\talias",
			output: []string{"col", "alias"},
		},
		{
			input:  "col\tAS\talias",
			output: []string{"col", "alias"},
		},
		{
			input:  "col AS alias name",
			output: []string{"col", "alias name"},
		},
		{
			input:  "col tS alias",
			output: []string{"col tS alias", ""},
		},
		{
			input:  "col AS alias AS name",
			output: []string{"col", "alias AS name"},
		},
	}

	for index, item := range data {
		col, alias := splitWithAS(item.input)
		a.Equal(col, item.output[0], "not equal @%d v1:%v,v2:%v", index, col, item.output[0])
		a.Equal(alias, item.output[1], "not equal @%d v1:%v,v2:%v", index, col, item.output[1])
	}
}

func TestQuoteColumn(t *testing.T) {
	a := assert.New(t)

	var data = []*struct {
		input  string
		output string
	}{
		{
			input:  "column",
			output: "{column}",
		},
		{
			input:  "column_name",
			output: "{column_name}",
		},
		{
			input:  "table.column_name",
			output: "{table}.{column_name}",
		},
	}

	b := New("")
	for index, item := range data {
		b.Reset()
		quoteColumn(b, item.input)
		output := b.String()
		a.Equal(output, item.output, "在第 %d 个元素出错，v1: %v，v2: %v", index, output, item.output)
	}
}
