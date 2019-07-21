// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package sqltest 提供对 SQL 内容测试的工具
package sqltest

import (
	"regexp"
	"strings"

	"github.com/issue9/assert"
)

var replacer = strings.NewReplacer(
	")", " ) ",
	"(", " ( ",
	",", " , ",
	"=", " = ",
)

var spaceReplaceRegexp = regexp.MustCompile("\\s+")

// Equal 检测两条 SQL 语句是否相等，忽略大小写与多余的空格。
func Equal(a *assert.Assertion, s1, s2 string) {
	// 将'(', ')', ',' 等字符的前后空格标准化
	s1 = replacer.Replace(s1)
	s2 = replacer.Replace(s2)

	// 转换成小写，去掉首尾空格
	s1 = strings.TrimSpace(strings.ToLower(s1))
	s2 = strings.TrimSpace(strings.ToLower(s2))

	// 去掉多余的空格。
	s1 = spaceReplaceRegexp.ReplaceAllString(s1, " ")
	s2 = spaceReplaceRegexp.ReplaceAllString(s2, " ")

	a.Equal(s1, s2)
}
