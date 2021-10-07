// SPDX-License-Identifier: MIT

// Package dialect 提供了部分数据库对 orm.Dialect 接口的实现
package dialect

import (
	"database/sql"
	"errors"
	"fmt"
	"unicode"

	"github.com/issue9/orm/v4/core"
	"github.com/issue9/orm/v4/sqlbuilder"
)

var (
	errColIsNil = errors.New("参数 col 参数是个空值")

	datetimeLayouts = []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.9",
		"2006-01-02 15:04:05.99",
		"2006-01-02 15:04:05.999",
		"2006-01-02 15:04:05.9999",
		"2006-01-02 15:04:05.99999",
		"2006-01-02 15:04:05.999999",
	}
)

func missLength(col *core.Column) error {
	return fmt.Errorf("列 %s 缺少长度数据", col.Name)
}

func invalidTimeFractional(col *core.Column) error {
	return fmt.Errorf("列 %s 时间精度只能介于 [0,6] 之间", col.Name)
}

func errUncovert(col *core.Column) error {
	return fmt.Errorf("不支持的列类型: %s", col.Name)
}

// mysqlLimitSQL mysql 系列数据库分页语法的实现
//
// 支持以下数据库：MySQL, H2, HSQLDB, Postgres, SQLite3
func mysqlLimitSQL(limit interface{}, offset ...interface{}) (string, []interface{}) {
	query := " LIMIT "

	if named, ok := limit.(sql.NamedArg); ok && named.Name != "" {
		query += "@" + named.Name
	} else {
		query += "?"
	}

	if len(offset) == 0 {
		return query + " ", []interface{}{limit}
	}

	query += " OFFSET "
	o := offset[0]
	if named, ok := o.(sql.NamedArg); ok && named.Name != "" {
		query += "@" + named.Name
	} else {
		query += "?"
	}

	return query + " ", []interface{}{limit, offset[0]}
}

// oracleLimitSQL oracle 系列数据库分页语法的实现
//
// 支持以下数据库：Derby, SQL Server 2012, Oracle 12c, the SQL 2008 standard
func oracleLimitSQL(limit interface{}, offset ...interface{}) (string, []interface{}) {
	query := "FETCH NEXT "

	if named, ok := limit.(sql.NamedArg); ok && named.Name != "" {
		query += "@" + named.Name
	} else {
		query += "?"
	}
	query += " ROWS ONLY "

	if len(offset) == 0 {
		return query, []interface{}{limit}
	}

	o := offset[0]
	if named, ok := o.(sql.NamedArg); ok && named.Name != "" {
		query = "OFFSET @" + named.Name + " ROWS " + query
	} else {
		query = "OFFSET ? ROWS " + query
	}

	return query, []interface{}{offset[0], limit}
}

// PrepareNamedArgs 对命名参数进行预处理
//
// 命名参数替换成 ?，并返回参数名称对应在语句的位置，包括 ? 在内。
func PrepareNamedArgs(query string) (string, map[string]int, error) {
	orders := map[string]int{}
	builder := core.NewBuilder("")
	start := -1
	cnt := 0

	write := func(name string) error {
		builder.WString(" ? ")

		if _, found := orders[name]; found {
			panic(fmt.Sprintf("存在相同的参数名：%s", name))
		}
		orders[name] = cnt
		return nil
	}

	for index, c := range query {
		switch {
		case c == '@':
			start = index + 1
		case start != -1 && !(unicode.IsLetter(c) || unicode.IsDigit(c)):
			if err := write(query[start:index]); err != nil {
				return "", nil, err
			}
			builder.WRunes(c) // 当前的字符不能丢
			cnt++
			start = -1
		case start == -1:
			builder.WRunes(c)
			if c == '?' {
				cnt++
			}
		}
	}

	if start > -1 {
		if err := write(query[start:]); err != nil {
			return "", nil, err
		}
	}

	q, err := builder.String()
	if err != nil {
		return "", nil, err
	}
	return q, orders, nil
}

func stdDropIndex(index string) (string, error) {
	if index == "" {
		return "", sqlbuilder.ErrColumnsIsEmpty
	}

	return core.NewBuilder("DROP INDEX ").
		QuoteKey(index).
		String()
}

func appendViewBody(builder *core.Builder, name, selectQuery string, cols []string) (string, error) {
	builder.WString(" VIEW ").QuoteKey(name)

	if len(cols) > 0 {
		builder.WBytes('(')
		for _, col := range cols {
			builder.QuoteKey(col).
				WBytes(',')
		}
		builder.TruncateLast(1).WBytes(')')
	}

	return builder.WString(" AS ").
		WString(selectQuery).
		String()
}

// 修正查询语句和查询参数的位置
func fixQueryAndArgs(query string, args []interface{}) (string, []interface{}, error) {
	query, orders, err := PrepareNamedArgs(query)
	if err != nil {
		return "", nil, err
	}

	// 整理返回参数
	named := make(map[int]interface{}, len(orders))
	for _, arg := range args {
		if n, ok := arg.(*sql.NamedArg); ok {
			i, found := orders[n.Name]
			if !found {
				panic(fmt.Sprintf("不存在指定名称的参数 %s", n.Name))
			}
			delete(orders, n.Name)
			named[i] = n.Value
			continue
		}
		if n, ok := arg.(sql.NamedArg); ok {
			i, found := orders[n.Name]
			if !found {
				panic(fmt.Sprintf("不存在指定名称的参数 %s", n.Name))
			}
			delete(orders, n.Name)
			named[i] = n.Value
			continue
		}
	}

	if len(orders) > 0 {
		panic("占位符与命名参数的数量不相同")
	}

	for index, val := range named {
		args[index] = val
	}

	return query, args, nil
}
