// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package mysql dialect 中用到的 mysql 专有功能
package mysql

import (
	"errors"
	"fmt"
	"strings"

	"github.com/issue9/orm/v2/sqlbuilder"
)

var quoteReplacer = strings.NewReplacer("`", "")

// Table 表信息
type Table struct {
	Columns     map[string]string                // 列信息，名称=>类型
	Constraints map[string]sqlbuilder.Constraint // 约束信息，名称=>约束类型
	Indexes     map[string]sqlbuilder.Index      // 索引信息，名称=>索引类型
}

// ParseCreateTable 分析 create table 的语法
func ParseCreateTable(table string, engine sqlbuilder.Engine) (*Table, error) {
	// show index 语句无法获取 check 约束的相关信息
	rows, err := engine.Query("SHOW CREATE TABLE `" + table + "`")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err1 := rows.Close(); err1 != nil {
			err = errors.New(err1.Error() + err.Error())
		}
	}()

	if !rows.Next() {
		return nil, errors.New("未找到 CREATE TABLE 数据")
	}

	var tableName, sql string
	if err = rows.Scan(&tableName, &sql); err != nil {
		return nil, err
	}

	return parseMysqlCreateTable(table, filterCreateTableSQL(sql))
}

// show create table 产生的格式比较统一，不像 create table 那样多样化。
// https://dev.mysql.com/doc/refman/8.0/en/show-create-table.html
func parseMysqlCreateTable(tableName string, lines []string) (*Table, error) {
	table := &Table{
		Columns:     make(map[string]string, len(lines)),
		Constraints: make(map[string]sqlbuilder.Constraint, len(lines)),
		Indexes:     make(map[string]sqlbuilder.Index, len(lines)),
	}

	for _, line := range lines {
		index := strings.IndexByte(line, ' ')
		if index <= 0 {
			return nil, fmt.Errorf("语法错误：%s", line)
		}
		first := line[:index]
		line = line[index+1:]

		switch strings.ToUpper(first) {
		case "INDEX", "KEY": // 索引
			index = strings.IndexByte(line, ' ')
			if index <= 0 {
				return nil, fmt.Errorf("语法错误:%s", line)
			}
			table.Indexes[line[:index]] = sqlbuilder.IndexDefault
		case "PRIMARY": // 主键约束，没有约束名
			table.Constraints[tableName+"_pk"] = sqlbuilder.ConstraintPK
		case "UNIQUE":
			words := strings.Fields(line)
			table.Constraints[words[1]] = sqlbuilder.ConstraintUnique
		case "CONSTRAINT": // check 或是 fk 约束
			words := strings.Fields(line)
			switch strings.ToUpper(words[1]) {
			case "FOREIGN":
				table.Constraints[words[0]] = sqlbuilder.ConstraintFK
			case "CHECK":
				table.Constraints[words[0]] = sqlbuilder.ConstraintCheck
			default:
				return nil, fmt.Errorf("未知的约束类型:%s", words[1])
			}
		default: // 普通列定义，第一个字符串即为列名
			table.Columns[first] = line
		}
	}

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
				break
			}
		} // end switch
	} // end for

	return lines
}
