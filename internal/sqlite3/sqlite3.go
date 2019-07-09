// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package sqlite3 针对 sqlite3 的一些额外处理，比如对 create table 的分析等。
package sqlite3

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/issue9/orm/v2/sqlbuilder"
)

var (
	quoteReplacer = strings.NewReplacer("`", "")

	// 从 sqlite_master 中查询 SQL 语句的代码
	//
	// 其中 queryIndex 加上了针对 sql IS NOT NULL 的判断，
	// 在 sqlite3 中，会自动生成一些索值，并不是我们需要的内容，
	// 我们只需要拿到 Create Index 的相关 SQL 语句。
	queryCreateTable = "SELECT sql FROM sqlite_master WHERE `type`='table' and tbl_name='%s'"
	queryIndex       = "SELECT name,sql FROM sqlite_master WHERE `type`='index' AND sql IS NOT NULL AND tbl_name='%s'"
)

// Table 包含从 sqlite_master 中获取的与当前表相关的信息
//
// 方便 dialect 从表信息中重新构建表内容。
type Table struct {
	Columns     map[string]string // 列信息，名称 => 类型
	Constraints map[string]*Constraint
	Indexes     map[string]*Index
}

// Index 表的索引信息
//
// 在 sqlite 中，索引是在创建表之后，别外提交的。
// 在修改表结构时，需要保存索引，方便之后重建。
type Index struct {
	Type sqlbuilder.Index
	SQL  string // 创建索引的语句
}

// Constraint 从 create table 语句解析出来的约束信息
type Constraint struct {
	Type sqlbuilder.Constraint
	SQL  string // 在 Create Table 中的语句
}

// CreateTableSQL 生成 create table 语句
func (t Table) CreateTableSQL(name string) string {
	builder := sqlbuilder.New("CREATE TABLE ").
		WriteString(name).
		WriteBytes('(')

	for _, col := range t.Columns {
		builder.WriteString(col).WriteBytes(',')
	}

	for _, cont := range t.Constraints {
		builder.WriteString(cont.SQL).WriteBytes(',')
	}

	builder.TruncateLast(1).WriteBytes(')')

	return builder.String()
}

// ParseCreateTable 从 sqlite_master 中获取 create table 并分析其内容
func ParseCreateTable(table string, engine sqlbuilder.Engine) (*Table, error) {
	tbl := &Table{
		Columns:     make(map[string]string, 10),
		Constraints: make(map[string]*Constraint, 5),
		Indexes:     make(map[string]*Index, 2),
	}

	if err := parseCreateTable(tbl, table, engine); err != nil {
		return nil, err
	}

	if err := parseIndexes(tbl, table, engine); err != nil {
		return nil, err
	}

	return tbl, nil
}

// https://www.sqlite.org/draft/lang_createtable.html
func parseCreateTable(table *Table, tableName string, engine sqlbuilder.Engine) error {
	rows, err := engine.Query(fmt.Sprintf(queryCreateTable, tableName))
	if err != nil {
		return err
	}
	defer func() {
		if err1 := rows.Close(); err1 != nil {
			err = errors.New(err1.Error() + err.Error())
		}
	}()

	if !rows.Next() {
		return errors.New("未找到任务 CREATE TABLE 数据")
	}

	var sql string
	if err = rows.Scan(&sql); err != nil {
		return err
	}

	lines := filterCreateTableSQL(sql)
	for _, line := range lines {
		index := strings.IndexByte(line, ' ')
		if index <= 0 {
			return fmt.Errorf("语法错误:%s", line)
		}
		first := line[:index]

		switch strings.ToUpper(first) {
		case "CONSTRAINT": // 约束
			words := strings.FieldsFunc(line[index+1:], func(r rune) bool { return unicode.IsSpace(r) || r == '(' })
			if len(words) < 2 {
				return fmt.Errorf("语法错误:%s", line)
			}

			cont := &Constraint{SQL: line}
			switch words[1] {
			case "PRIMARY":
				cont.Type = sqlbuilder.ConstraintPK
			case "UNIQUE":
				cont.Type = sqlbuilder.ConstraintUnique
			case "CHECK":
				cont.Type = sqlbuilder.ConstraintCheck
			case "FOREIGN":
				cont.Type = sqlbuilder.ConstraintFK
			default:
				return fmt.Errorf("未知的约束名：%s", line)
			}

			table.Constraints[words[0]] = cont
		default: // 普通列定义，第一个字符串即为列名
			table.Columns[first] = line
		}
	}

	return nil
}

func parseIndexes(table *Table, tableName string, engine sqlbuilder.Engine) error {
	// 通过 sql IS NOT NULL 过滤掉自动生成的索引值
	rows, err := engine.Query(fmt.Sprintf(queryIndex, tableName))
	if err != nil {
		return err
	}
	defer func() {
		if err1 := rows.Close(); err1 != nil {
			err = errors.New(err1.Error() + err.Error())
		}
	}()

	for rows.Next() {
		var name, sql string
		if err := rows.Scan(&name, &sql); err != nil {
			return err
		}
		table.Indexes[name] = &Index{
			SQL:  sql,
			Type: sqlbuilder.IndexDefault,
		}
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
