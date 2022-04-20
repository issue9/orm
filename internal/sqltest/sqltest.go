// SPDX-License-Identifier: MIT

// Package sqltest 提供对 SQL 内容测试的工具
package sqltest

import (
	"regexp"
	"strings"

	"github.com/issue9/assert/v2"
)

var replacer = strings.NewReplacer(
	")", " ) ",
	"(", " ( ",
	",", " , ",
	"=", " = ",
)

var spaceReplaceRegexp = regexp.MustCompile("\\s+")

// Equal 检测两条 SQL 语句是否相等
//
// 忽略大小写与多余的空格。
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

	a.TB().Helper()

	a.Equal(s1, s2)
}
