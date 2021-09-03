// SPDX-License-Identifier: MIT

package createtable

import (
	"fmt"
	"strings"

	"github.com/issue9/orm/v4/core"
)

// Sqlite3Table 包含从 sqlite_master 中获取的与当前表相关的信息
type Sqlite3Table struct {
	Columns     map[string]string // 列信息，名称 => SQL 语句
	Constraints map[string]*Sqlite3Constraint
	Indexes     map[string]*Sqlite3Index
}

// Sqlite3Index 表的索引信息
//
// 在 sqlite 中，索引是在创建表之后，别外提交的。
// 在修改表结构时，需要保存索引，方便之后重建。
type Sqlite3Index struct {
	Type core.Index
	SQL  string // 创建索引的语句
}

// Sqlite3Constraint 从 create table 语句解析出来的约束信息
type Sqlite3Constraint struct {
	Type core.Constraint
	SQL  string // 在 Create Sqlite3Table 中的语句
}

// CreateTableSQL 生成 create table 语句
func (t Sqlite3Table) CreateTableSQL(name string) (string, error) {
	builder := core.NewBuilder("CREATE TABLE ").
		WString(name).
		WBytes('(')

	for _, col := range t.Columns {
		builder.WString(col).WBytes(',')
	}

	for _, cont := range t.Constraints {
		builder.WString(cont.SQL).WBytes(',')
	}

	builder.TruncateLast(1).WBytes(')')

	return builder.String()
}

// ParseSqlite3CreateTable 从 sqlite_master 中获取 create table 并分析其内容
func ParseSqlite3CreateTable(table string, engine core.Engine) (*Sqlite3Table, error) {
	tbl := &Sqlite3Table{
		Columns:     make(map[string]string, 10),
		Constraints: make(map[string]*Sqlite3Constraint, 5),
		Indexes:     make(map[string]*Sqlite3Index, 2),
	}

	if err := parseSqlite3CreateTable(tbl, table, engine); err != nil {
		return nil, err
	}

	if err := parseSqlite3Indexes(tbl, table, engine); err != nil {
		return nil, err
	}

	return tbl, nil
}

// https://www.sqlite.org/draft/lang_createtable.html
func parseSqlite3CreateTable(table *Sqlite3Table, tableName string, engine core.Engine) error {
	query := "SELECT sql FROM sqlite_master WHERE `type`='table' and tbl_name='" + tableName + "'"
	var sql string
	if err := scanCreateTable(engine, tableName, query, &sql); err != nil {
		return err
	}

	lines := lines(sql)
	for _, line := range lines {
		index := strings.IndexByte(line, ' ')
		if index <= 0 {
			return fmt.Errorf("语法错误:%s", line)
		}
		first := line[:index]

		switch strings.ToUpper(first) {
		case "CONSTRAINT": // 约束
			words := fields(line[index+1:])
			if len(words) < 2 {
				return fmt.Errorf("语法错误:%s", line)
			}

			cont := &Sqlite3Constraint{SQL: line}
			switch words[1] {
			case "PRIMARY":
				cont.Type = core.ConstraintPK
			case "UNIQUE":
				cont.Type = core.ConstraintUnique
			case "CHECK":
				cont.Type = core.ConstraintCheck
			case "FOREIGN":
				cont.Type = core.ConstraintFK
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

func parseSqlite3Indexes(table *Sqlite3Table, tableName string, engine core.Engine) error {
	// 通过 sql IS NOT NULL 过滤掉自动生成的索引值
	query := "SELECT name,sql FROM sqlite_master WHERE `type`='index' AND sql IS NOT NULL AND tbl_name='" + tableName + "'"
	rows, err := engine.Query(query)
	if err != nil {
		return err
	}
	defer func() {
		if err1 := rows.Close(); err1 != nil {
			err = fmt.Errorf("在抛出错误 %s 时再次发生错误 %w", err.Error(), err1)
		}
	}()

	for rows.Next() {
		var name, sql string
		if err := rows.Scan(&name, &sql); err != nil {
			return err
		}
		table.Indexes[name] = &Sqlite3Index{
			SQL:  sql,
			Type: core.IndexDefault,
		}
	}

	return nil
}
