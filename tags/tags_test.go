// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tags

import (
	"testing"

	"github.com/issue9/assert"
)

type testData struct { // 测试数据结构
	tag  string              // 待分析字符串
	data map[string][]string // 分析后数据
}

var tests = []*testData{
	&testData{
		tag: "name,abc;name2,;;name3,n1,n2",
		data: map[string][]string{
			"name":  []string{"abc"},
			"name2": []string{},
			"name3": []string{"n1", "n2"},
		},
	},
	&testData{
		tag: "name(abc);name2,;;name3(n1,n2)",
		data: map[string][]string{
			"name":  []string{"abc"},
			"name2": []string{},
			"name3": []string{"n1", "n2"},
		},
	},
	&testData{
		tag:  "",
		data: nil,
	},
	&testData{
		tag:  "",
		data: map[string][]string{},
	},
}

func TestReplace(t *testing.T) {
	tag1 := "name,abc;name2,;;name3,n1,n2"
	tag2 := "name(abc);name2,;;name3(n1,n2)"
	tag := styleReplace.Replace(tag2)
	assert.Equal(t, tag, tag1)
}

func TestParse(t *testing.T) {
	a := assert.New(t)

	for _, test := range tests {
		m := Parse(test.tag)
		if m != nil { // m == nil或是m == map[string][]string{}
			a.Equal(m, test.data)
		}
	}
}

func TestGet(t *testing.T) {
	a := assert.New(t)

	for _, test := range tests {
		for name, items := range test.data {
			val, found := Get(test.tag, name)
			println(test.tag)
			a.True(found).Equal(val, items)

			val, found = Get(test.tag, name+"-temp")
			a.False(found).Nil(val)
		}
	}
}

func TestMustGet(t *testing.T) {
	a := assert.New(t)

	for _, test := range tests {
		for name, items := range test.data {
			val := MustGet(test.tag, name, "default")
			a.Equal(val, items)

			val = MustGet(test.tag, name+"-temp", "def1", "def2")
			a.Equal(val, []string{"def1", "def2"})
		}
	}
}

func TestHas(t *testing.T) {
	a := assert.New(t)

	for _, test := range tests {
		for name, _ := range test.data {
			a.True(Has(test.tag, name))

			a.False(Has(test.tag, name+"-temp"))
		}
	}
}
