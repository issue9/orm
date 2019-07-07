// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package parser

import (
	"errors"
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
func ParseCreateTable(driverName, table string, engine sqlbuilder.Engine) (*Table, error) {
	switch driverName {
	case "mysql":
		return getMysqlTableInfo(table, engine)
	case "sqlite3":
		return getSqlite3TableInfo(table, engine)
	}
	panic("未实现的数据库：" + driverName)
}

func getMysqlTableInfo(table string, engine sqlbuilder.Engine) (*Table, error) {
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

	var tableName, sql string
	if err = rows.Scan(&tableName, &sql); err != nil {
		return nil, err
	}

	return parseMysqlCreateTable(filterCreateTableSQL(sql))
}

func getSqlite3TableInfo(tableName string, engine sqlbuilder.Engine) (*Table, error) {
	table := &Table{
		Columns:     make(map[string]string, 10),
		Constraints: make(map[string]sqlbuilder.Constraint, 5),
		Indexes:     make(map[string]sqlbuilder.Index, 2),
	}

	if err := parseSqlite3CreateTable(table, tableName, engine); err != nil {
		return nil, err
	}

	if err := parseSqlite3Indexes(table, tableName, engine); err != nil {
		return nil, err
	}

	return table, nil
}

// show create table 产生的格式比较统一，不像 create table 那样多样化。
// https://dev.mysql.com/doc/refman/8.0/en/show-create-table.html
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
		case "INDEX", "KEY": // 索引
			index = strings.IndexByte(line, ' ')
			if index <= 0 {
				continue
			}
			table.Indexes[line[:index]] = sqlbuilder.IndexDefault
		case "PRIMARY":
			words := strings.Fields(line)
			table.Constraints[words[1]] = sqlbuilder.ConstraintPK
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
			table.Columns[line[:index]] = line
		}
	}

	return table, nil
}

// https://www.sqlite.org/draft/lang_createtable.html
func parseSqlite3CreateTable(table *Table, tableName string, engine sqlbuilder.Engine) error {
	rows, err := engine.Query("SELECT sql FROM sqlite_master WHERE type='table' and name='" + tableName + "'")
	if err != nil {
		return err
	}
	defer func() {
		if err1 := rows.Close(); err1 != nil {
			err = errors.New(err1.Error() + err.Error())
		}
	}()

	var sql string
	if err = rows.Scan(&sql); err != nil {
		return err
	}

	lines := filterCreateTableSQL(sql)

LOOP:
	for _, line := range lines {
		line = strings.TrimSpace(line)
		index := strings.IndexByte(line, ' ')
		if index <= 0 {
			continue
		}
		first := line[:index]
		line = line[index+1:]

		switch strings.ToUpper(first) {
		case "CONSTRAINT": // 约束
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
				return fmt.Errorf("未知的约束名：%s", line[:index])
			}
		default: // 普通列定义，第一个字符串即为列名
			table.Columns[line[:index]] = line
		}
	}

	return nil
}
func parseSqlite3Indexes(table *Table, tableName string, engine sqlbuilder.Engine) error {
	rows, err := engine.Query("SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='" + tableName + "'")
	if err != nil {
		return err
	}
	defer func() {
		if err1 := rows.Close(); err1 != nil {
			err = errors.New(err1.Error() + err.Error())
		}
	}()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		table.Indexes[name] = sqlbuilder.IndexDefault
	}

	return nil
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
