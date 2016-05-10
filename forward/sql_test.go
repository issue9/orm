// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package forward

import (
	"regexp"
	"strings"
	"testing"

	"github.com/issue9/assert"
)

var replacer = strings.NewReplacer(")", " ) ", "(", " ( ", ",", " , ")

var spaceReplaceRegexp = regexp.MustCompile("\\s+")

// 检测两条SQL语句是否相等，忽略大小写与多余的空格。
func chkSQLEqual(a *assert.Assertion, s1, s2 string) {
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

func TestSQL_TruncateLast(t *testing.T) {
	a := assert.New(t)

	sql := NewSQL(nil)
	a.NotNil(sql)

	sql.WriteString("123").TruncateLast(1)
	a.False(sql.HasError())
	a.Equal(sql.buffer.Len(), 2)

	sql.TruncateLast(1)
	a.Equal(sql.buffer.String(), "1")
}

func TestSQL_Delete(t *testing.T) {
	a := assert.New(t)

	sql := NewSQL(nil)
	a.NotNil(sql)

	str, vals, err := sql.Delete("table1").
		Where("id=?", 1).
		And("name=?", "n").String()
	a.NotError(err).Equal(vals, []interface{}{1, "n"})
	chkSQLEqual(a, str, "DELETE FROM table1 WHERE id=? AND name=?")
}
