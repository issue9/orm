// SPDX-License-Identifier: MIT

// Package dialect 提供了部分数据库对 orm.Dialect 接口的实现
package dialect

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/issue9/orm/v3/core"
)

var (
	errColIsNil              = errors.New("参数 col 参数是个空值")
	errGoTypeIsNil           = errors.New("无效的 col.GoType 值")
	errMissLength            = errors.New("缺少长度")
	errTimeFractionalInvalid = errors.New("时间精度只能介于 [0,6] 之间")

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

func errUncovert(dest string) error {
	return fmt.Errorf("不支持的类型: %s", dest)
}

// MysqlLimitSQL mysql 系列数据库分页语法的实现
//
// 支持以下数据库：MySQL, H2, HSQLDB, Postgres, SQLite3
func MysqlLimitSQL(limit interface{}, offset ...interface{}) (string, []interface{}) {
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

// OracleLimitSQL oracle 系列数据库分页语法的实现
//
// 支持以下数据库：Derby, SQL Server 2012, Oracle 12c, the SQL 2008 standard
func OracleLimitSQL(limit interface{}, offset ...interface{}) (string, []interface{}) {
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

type namedArg struct {
	sql.NamedArg
	index int
}

// ReplaceNamedArgs 替换 SQL 语句中的命名参数为标准的 ? 符号
func ReplaceNamedArgs(query string, args []interface{}) string {
	as := make([]namedArg, 0, len(args))

	for index, arg := range args {
		if named, ok := arg.(sql.NamedArg); ok {
			as = append(as, namedArg{
				NamedArg: named,
				index:    index,
			})
			continue
		}

		if named, ok := arg.(*sql.NamedArg); ok {
			as = append(as, namedArg{
				NamedArg: *named,
				index:    index,
			})
		}
	}

	// 将名称长的排到前面，确保可以正确替换
	sort.SliceStable(as, func(i, j int) bool {
		return len(as[i].Name) > len(as[j].Name)
	})

	for _, arg := range as {
		query = strings.Replace(query, "@"+arg.Name, "?", 1)
		args[arg.index] = arg.Value
	}

	return query
}

// PrepareNamedArgs 对命名参数进行预处理
//
// 返回符合 core.Dialect.Prepare 方法的数据。
func PrepareNamedArgs(query string) (string, map[string]int, error) {
	orders := map[string]int{}
	builder := core.NewBuilder("")
	start := -1
	cnt := 0

	write := func(name string, val int) error {
		builder.WString(" ? ")

		if _, found := orders[name]; found {
			return fmt.Errorf("存在相同的参数名：%s", name)
		}
		orders[name] = cnt
		return nil
	}

	var qm bool // 是否存在问号
	for index, c := range query {
		switch {
		case c == '@':
			start = index + 1
		case start != -1 && !unicode.IsLetter(c):
			if err := write(query[start:index], cnt); err != nil {
				return "", nil, err
			}
			builder.WRunes(c) // 当前的字符不能丢
			cnt++
			start = -1
		case start == -1:
			builder.WRunes(c)
			if c == '?' {
				qm = true
			}
		}
	}

	if qm && len(orders) > 0 {
		return "", nil, errors.New("命名参数与 ? 不能同时存在")
	}

	if start > -1 {
		if err := write(query[start:], cnt); err != nil {
			return "", nil, err
		}
	}

	q, err := builder.String()
	if err != nil {
		return "", nil, err
	}
	return q, orders, nil
}
