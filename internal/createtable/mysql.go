// SPDX-License-Identifier: MIT

package createtable

import (
	"fmt"
	"strings"

	"github.com/issue9/orm/v4/core"
)

// MysqlTable 从 create table 中获取的表信息
type MysqlTable struct {
	Columns     map[string]string          // 列信息，名称 => 类型定义
	Constraints map[string]core.Constraint // 约束信息，名称 => 约束类型
	Indexes     map[string]core.Index      // 索引信息，名称 => 索引类型
}

// ParseMysqlCreateTable 分析 create table 的语法
func ParseMysqlCreateTable(table string, engine core.Engine) (*MysqlTable, error) {
	// show index 语句无法获取 check 约束的相关信息
	query := "SHOW CREATE TABLE {" + table + "}"
	var tableName, sql string
	if err := scanCreateTable(engine, table, query, &tableName, &sql); err != nil {
		return nil, err
	}

	// table 与 tableName 值可能是不相同的。
	// table 的值可能是 #tbl 而 tableName 的值可能是 prefix_tbl，
	// 两者之间一个表示未替换表名前缀之前的值，一个为替换之后的值。
	return parseMysqlCreateTable(tableName, lines(sql))
}

// show create table 产生的格式比较统一，不像 create table 那样多样化。
// https://dev.mysql.com/doc/refman/8.0/en/show-create-table.html
func parseMysqlCreateTable(tableName string, lines []string) (*MysqlTable, error) {
	table := &MysqlTable{
		Columns:     make(map[string]string, len(lines)),
		Constraints: make(map[string]core.Constraint, len(lines)),
		Indexes:     make(map[string]core.Index, len(lines)),
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
			table.Indexes[line[:index]] = core.IndexDefault
		case "PRIMARY": // 主键约束，没有约束名
			table.Constraints[core.PKName(tableName)] = core.ConstraintPK
		case "UNIQUE":
			words := fields(line)
			table.Constraints[words[1]] = core.ConstraintUnique
		case "CONSTRAINT": // check 或是 fk 约束
			words := fields(line)
			switch strings.ToUpper(words[1]) {
			case "FOREIGN":
				table.Constraints[words[0]] = core.ConstraintFK
			case "CHECK":
				table.Constraints[words[0]] = core.ConstraintCheck
			default:
				return nil, fmt.Errorf("未知的约束类型:%s", words[1])
			}
		default: // 普通列定义，第一个字符串即为列名
			table.Columns[first] = line
		}
	}

	return table, nil
}
