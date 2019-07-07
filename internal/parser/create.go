// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package parser

import (
	"fmt"
	"strings"

	"github.com/issue9/orm/v2/sqlbuilder"
)

var (
	// mysql 和 sqlite3 使用相同的字符
	quoteReplacer = strings.NewReplacer("`", "")
)

// Table 表信息
type Table struct {
	Columns     map[string]string                // 列信息，名称=>类型
	Constraints map[string]sqlbuilder.Constraint // 约束信息，名称=>约束类型
	Indexes     map[string]sqlbuilder.Index      // 索引信息，名称=>索引类型
}

// ParseCreateTable 分析 create table 的语法
func ParseCreateTable(sql, driverName string) (*Table, error) {
	switch driverName {
	case "mysql":
		return parseMysqlCreateTable(filterCreateTableSQL(sql))
	case "sqlite3":
		return parseSqlite3CreateTable(filterCreateTableSQL(sql))
	}
	panic("未实现的数据库：" + driverName)
}

// https://dev.mysql.com/doc/refman/8.0/en/create-table.html
func parseMysqlCreateTable(lines []string) (*Table, error) {
	table := &Table{
		Columns:     make(map[string]string, len(lines)),
		Constraints: make(map[string]sqlbuilder.Constraint, len(lines)),
		Indexes:     make(map[string]sqlbuilder.Index, len(lines)),
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		index := strings.IndexByte(line, ' ')
		if index <= 0 {
			continue
		}
		first := line[:index]
		line = line[index+1:]

		switch strings.ToUpper(first) {
		case "FULLTEXT", "SPATIAL": // FULLTEXT|SPATIAL INDEX
			index = strings.IndexByte(line, ' ')
			if index <= 0 {
				continue
			}
			first = line[:index]
			line = line[index+1:]
			fallthrough
		case "INDEX", "KEY": // 索引
			index = strings.IndexByte(line, ' ')
			if index <= 0 {
				continue
			}
			table.Indexes[line[:index]] = sqlbuilder.IndexDefault
		case "CONSTRAINT":
			index = strings.IndexByte(line, ' ')
			if index <= 0 {
				continue
			}
			first = line[:index]
			line = line[index+1:]
			fallthrough
		case "UNIQUE", "PRIMARY", "FOREIGN": // 约束
		// TODO
		default: // 普通列定义，第一个字符串即为列名
			table.Columns[line[:index]] = line[index+1:]
		}
	}

	return table, nil
}

// https://www.sqlite.org/draft/lang_createtable.html
func parseSqlite3CreateTable(lines []string) (*Table, error) {
	table := &Table{
		Columns:     make(map[string]string, len(lines)),
		Constraints: make(map[string]sqlbuilder.Constraint, len(lines)),
		Indexes:     make(map[string]sqlbuilder.Index, len(lines)),
	}

LOOP:
	for _, line := range lines {
		line = strings.TrimSpace(line)
		index := strings.IndexByte(line, ' ')
		if index <= 0 {
			continue
		}
		first := line[:index]

		switch strings.ToUpper(first) {
		case "CONSTRAINT": // 约束
			line = strings.TrimSpace(line[index+1:])
			index = strings.IndexByte(line, ' ')
			if index <= 0 {
				continue LOOP
			}
			name := line[:index]

			line = strings.TrimSpace(line[index+1:])
			index = strings.IndexByte(line, ' ')
			if index <= 0 {
				continue LOOP
			}
			switch line[:index] {
			case "PRIMARY":
				table.Constraints[name] = sqlbuilder.ConstraintPK
			case "UNIQUE":
				table.Constraints[name] = sqlbuilder.ConstraintUnique
			case "CHECK":
				table.Constraints[name] = sqlbuilder.ConstraintCheck
			case "FOREIGN":
				table.Constraints[name] = sqlbuilder.ConstraintFK
			default:
				return nil, fmt.Errorf("未知的约束名：%s", line[:index])
			}
		default: // 普通列定义，第一个字符串即为列名
			table.Columns[line[:index]] = line[index+1:]
		}
	}

	// TODO 如何获取索引信息

	return table, nil
}

func filterCreateTableSQL(sql string) []string {
	sql = quoteReplacer.Replace(sql)
	var deep, start int
	var lines []string

	for index, c := range sql {
		switch c {
		case ',':
			if deep == 1 && index > start {
				lines = append(lines, sql[start:index])
				start = index
			}
		case '(':
			deep++
			if deep == 1 {
				start = index
			}
		case ')':
			deep--
			if deep == 0 { // 不需要 create table xx() 之后的内容
				break
			}
		} // end switch
	} // end for

	return lines
}
