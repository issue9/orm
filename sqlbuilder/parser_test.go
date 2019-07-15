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

	col, alias := splitWithAS("col as alias")
	a.Equal(col, "col").Equal(alias, "alias")

	col, alias = splitWithAS("col As alias")
	a.Equal(col, "col").Equal(alias, "alias")

	col, alias = splitWithAS("col AS\talias")
	a.Equal(col, "col").Equal(alias, "alias")

	col, alias = splitWithAS("col\taS alias")
	a.Equal(col, "col").Equal(alias, "alias")

	col, alias = splitWithAS("col aS alias name")
	a.Equal(col, "col").Equal(alias, "alias name")

	col, alias = splitWithAS("col tS alias")
	a.Equal(col, "col tS alias").Equal(alias, "")

	col, alias = splitWithAS("col AS alias AS name")
	a.Equal(col, "col").Equal(alias, "alias AS name")
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
