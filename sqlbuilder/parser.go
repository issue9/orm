// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bufio"
	"strings"
	"unicode"
)

var quoteReplacer = strings.NewReplacer("{", "", "}", "")

// 从表达式中获取列的名称。
//
// 如果不存在别名，则取其列名或是整个表达式作为别名。
//  *  => *
//  table.*  => *
//  table.col  => {col}
//  table.col as col  => {col}
//  sum(table.count) as cnt  ==> {cnt}
//  func1(func2(table.col1),table.col2) as fn1  ==> {fn1}
//  count({table.*})  => {count(table.*)}
func getColumnName(expr string) string {
	if expr == "*" {
		return expr
	}

	s := bufio.NewScanner(strings.NewReader(expr))
	s.Split(splitWithAS)

	var name string
	for s.Scan() {
		name = s.Text()
	}

	if len(name) == 0 || name == "*" {
		return name
	}

	// 尽量取列名部分作为别名，如果包含了函数信息，
	// 则将整个表达式作为别名。
	var deep, start int
	for i, b := range name {
		switch {
		case b == '{':
			deep++
		case b == '}':
			deep--
		case b == '.' && deep == 0:
			start = i
		case b == '(': // 包含函数信息，则将整个表达式作为别名
			return "{" + quoteReplacer.Replace(name) + "}"
		}
	}

	if start > 0 {
		name = name[start+1:]
	}

	if name == "*" || name[0] == '{' {
		return name
	}

	return "{" + name + "}"
}

func splitWithAS(data []byte, atEOF bool) (advance int, token []byte, err error) {
	var start, deep int
	var b byte

	// 去掉行首的空格
	for start, b = range data {
		if !unicode.IsSpace(rune(b)) {
			break
		}
	}

	// 找到第一个 AS 字符串
	for i, b := range data {
		if b == '{' {
			deep++
			continue
		}

		if b == '}' {
			deep--
			continue
		}

		if deep != 0 {
			continue
		}

		if !unicode.IsSpace(rune(b)) {
			continue
		}

		if len(data) <= i+3 {
			break
		}

		b1 := data[i+1]
		b2 := data[i+2]
		b3 := data[i+3]
		if (b1 == 'a' || b1 == 'A') &&
			(b2 == 's' || b2 == 'S') &&
			unicode.IsSpace(rune(b3)) {
			return i + 4, data[start:i], nil
		}
	}

	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}

	return start, nil, nil
}
