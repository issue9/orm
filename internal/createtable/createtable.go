// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package createtable 分析 create table 语句的内容
package createtable

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/issue9/orm/v6/core"
)

var backQuoteReplacer = strings.NewReplacer("`", "")

func lines(sql string) []string {
	sql = backQuoteReplacer.Replace(sql)
	var deep, start int
	var lines []string

LOOP:
	for index, c := range sql {
		switch c {
		case ',':
			if deep == 1 && index > start {
				lines = append(lines, strings.TrimSpace(sql[start:index]))
				start = index + 1 // 不包含 ( 本身
			}
		case '(':
			deep++
			if deep == 1 {
				start = index + 1 // 不包含 ( 本身
			}
		case ')':
			deep--
			if deep == 0 { // 不需要 create table xx() 之后的内容
				if start != index {
					lines = append(lines, strings.TrimSpace(sql[start:index]))
				}
				break LOOP
			}
		} // end switch
	} // end for

	return lines
}

func fields(line string) []string {
	return strings.FieldsFunc(line, func(r rune) bool {
		return unicode.IsSpace(r) || r == '(' || r == ')'
	})
}

// 获取 create table 的内容
//
// query 查询 create table 的语句；
// val 从查询语句中获取的值。
func scanCreateTable(engine core.Engine, table, query string, val ...any) error {
	rows, err := engine.Query(query)
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, rows.Close())
	}()

	if !rows.Next() {
		return fmt.Errorf("未找到任何与 %s 相关的 CREATE TABLE 数据", table)
	}

	return rows.Scan(val...)
}
