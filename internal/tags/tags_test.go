// SPDX-License-Identifier: MIT

package tags

import (
	"testing"

	"github.com/issue9/assert"
)

type testData struct { // 测试数据结构
	tag  string // 待分析字符串
	data []*Tag // 分析后数据
}

var tests = []*testData{
	{
		tag: "name,abc;name2,;;name3,n1,n2;name3(n3,n4)",
		data: []*Tag{
			{
				Name: "name",
				Args: []string{"abc"},
			},
			{
				Name: "name2",
				Args: []string{},
			},
			{
				Name: "name3",
				Args: []string{"n1", "n2"},
			},
			{
				Name: "name3",
				Args: []string{"n3", "n4"},
			},
		},
	},
	{
		tag: "name(abc);name2,;;name3(n1,n2);name3(n3,n4)",
		data: []*Tag{
			{
				Name: "name",
				Args: []string{"abc"},
			},
			{
				Name: "name2",
				Args: []string{},
			},
			{
				Name: "name3",
				Args: []string{"n1", "n2"},
			},
			{
				Name: "name3",
				Args: []string{"n3", "n4"},
			},
		},
	},
	{
		tag:  "",
		data: nil,
	},
	{
		tag:  "",
		data: []*Tag{},
	},
}

func TestReplace(t *testing.T) {
	tag1 := "name,abc;name2,;;name3,n1,n2;name3,n1,n2"
	tag2 := "name(abc);name2,;;name3(n1,n2);name3(n1,n2)"
	tag := styleReplace.Replace(tag2)
	assert.Equal(t, tag, tag1)
}

func TestParse(t *testing.T) {
	a := assert.New(t)

	for _, test := range tests {
		m := Parse(test.tag)
		if m != nil {
			for index, item := range m {
				a.Equal(item, test.data[index])
			}
		}
	}
}

func TestGet(t *testing.T) {
	a := assert.New(t)

	for _, test := range tests {
		for _, items := range test.data {
			t.Log(test.tag)
			val, found := Get(test.tag, items.Name)
			a.True(found)
			if items.Name == "name3" {
				a.Equal(val, []string{"n1", "n2"}) // 多个重名的，只返回第一个数据
			} else {
				a.Equal(val, items.Args)
			}

			val, found = Get(test.tag, items.Name+"-temp")
			a.False(found).Nil(val)
		}
	}
}

func TestMustGet(t *testing.T) {
	a := assert.New(t)

	for _, test := range tests {
		for _, items := range test.data {
			val := MustGet(test.tag, items.Name, "default")
			if items.Name == "name3" {
				a.Equal(val, []string{"n1", "n2"}) // 多个重名的，只返回第一个数据
			} else {
				a.Equal(val, items.Args)
			}

			val = MustGet(test.tag, items.Name+"-temp", "def1", "def2")
			a.Equal(val, []string{"def1", "def2"})
		}
	}
}

func TestHas(t *testing.T) {
	a := assert.New(t)

	for _, test := range tests {
		for _, item := range test.data {
			a.True(Has(test.tag, item.Name))

			a.False(Has(test.tag, item.Name+"-temp"))
		}
	}
}
