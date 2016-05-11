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

func TestSQLBuilder_Reset(t *testing.T) {
	a := assert.New(t)
	sql := New(nil)
	a.NotNil(sql)

	sql.Delete("t1").Where("col1=?", 1)
	a.True(len(sql.args) > 0).True(sql.flag > 0)
	sql.Reset()
	a.Equal(len(sql.args), 0).Equal(sql.flag, 0)
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
	sql := New(&engine{dialect: &dialect{}})
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

func TestSQLBuilder_Select(t *testing.T) {
	a := assert.New(t)
	e := &engine{dialect: &dialect{}}
	sql := New(e)
	a.NotNil(sql)

	query, vals, err := sql.Select("col1,col2").
		Select("col3 AS c3", "col4").
		From("table1 AS t1").
		Join("left", "table2 as t2", "t1.id=t2.uid").
		Where("t1.id>?", 5).
		Or("t1.id<?", 3).
		Asc("t1.id").
		Desc("t1.type").
		Asc("t1.name").
		Limit(20, 5).
		String()

	a.NotError(err).Equal(vals, []interface{}{5, 3, 20, 5})
	chkSQLEqual(a, query, `SELECT col1,col2,col3 AS c3,col4 FROM
	table1 AS t1
	LEFT JOIN table2 as t2 ON t1.id=t2.uid
	where t1.id>? or t1.id<?
	order by t1.id asc,t1.type desc,t1.name asc
	limit ? offset ?`)
}
