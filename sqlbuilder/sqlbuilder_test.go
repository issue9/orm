// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

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

func TestSQLBuilder_TruncateLast(t *testing.T) {
	a := assert.New(t)

	sql := New(nil)
	a.NotNil(sql)

	sql.WriteString("123").TruncateLast(1)
	a.False(sql.HasError())
	a.Equal(sql.buffer.Len(), 2)

	sql.TruncateLast(1)
	a.Equal(sql.buffer.String(), "1")
}

func TestSQLBuilder_Delete(t *testing.T) {
	a := assert.New(t)

	sql := New(nil)
	a.NotNil(sql)

	query, vals, err := sql.Delete("table1").
		Where("id=?", 1).
		And("name=?", "n").
		String()
	a.NotError(err).
		Equal(vals, []interface{}{1, "n"})
	chkSQLEqual(a, query, "DELETE FROM table1 WHERE id=? AND name=?")
}

func TestSQLBuilder_Insert(t *testing.T) {
	a := assert.New(t)

	sql := New(nil)
	a.NotNil(sql)

	query, vals, err := sql.Insert("table1").
		Keys("col1", "col2").
		Values(1, 1).
		Values(2, 2).
		String()

	a.NotError(err).
		Equal(vals, []interface{}{1, 1, 2, 2})
	chkSQLEqual(a, query, "INSERT INTO table1 (col1,col2) VALUES(?,?),(?,?)")
}

func TestSQLBuilder_Update(t *testing.T) {
	a := assert.New(t)

	sql := New(nil)
	a.NotNil(sql)

	query, vals, err := sql.Update("table1").
		Set("col1", 1).
		Set("col2", "2").
		String()

	a.NotError(err).Equal(vals, []interface{}{1, "2"})
	chkSQLEqual(a, query, "UPDATE table1 set col1=?,col2=?")
}
