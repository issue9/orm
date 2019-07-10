// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"errors"
	"strconv"
	"strings"
	"unicode"

	"github.com/issue9/orm/v2/fetch"
)

// ErrNoData 在 Select.QueryInt 等函数中，
// 如果没有符合条件的数据，则返回此错误。
var ErrNoData = errors.New("不存在符合和条件的数据")

// SelectStmt 查询语句
type SelectStmt struct {
	*queryStmt

	table     string
	where     *WhereStmt
	cols      []string
	distinct  bool
	forUpdate bool

	// COUNT 查询的列内容
	countExpr string

	joins  []*join
	orders *SQLBuilder
	group  string

	havingQuery string
	havingVals  []interface{}

	limitQuery string
	limitVals  []interface{}
}

type join struct {
	typ   string
	on    string
	table string
}

// Select 声明一条 Select 语句
func Select(e Engine, d Dialect) *SelectStmt {
	stmt := &SelectStmt{where: Where()}
	stmt.queryStmt = newQueryStmt(e, d, stmt)

	return stmt
}

// Distinct 声明一条 Select 语句的 Distinct
//
// 若指定了此值，则 Select() 所指定的列，均为 Distinct 之后的列。
func (stmt *SelectStmt) Distinct() *SelectStmt {
	stmt.distinct = true
	return stmt
}

// Reset 重置语句
func (stmt *SelectStmt) Reset() *SelectStmt {
	stmt.table = ""
	stmt.where.Reset()
	stmt.cols = stmt.cols[:0]
	stmt.distinct = false
	stmt.forUpdate = false

	stmt.countExpr = ""

	stmt.joins = stmt.joins[:0]
	if stmt.orders != nil {
		stmt.orders.Reset()
	}
	stmt.group = ""

	stmt.havingQuery = ""
	stmt.havingVals = nil

	stmt.limitQuery = ""
	stmt.limitVals = nil

	return stmt
}

// SQL 获取 SQL 语句及对应的参数
func (stmt *SelectStmt) SQL() (string, []interface{}, error) {
	if stmt.table == "" {
		return "", nil, ErrTableIsEmpty
	}

	if len(stmt.cols) == 0 && stmt.countExpr == "" {
		return "", nil, ErrColumnsIsEmpty
	}

	buf := New("SELECT ")
	args := make([]interface{}, 0, 10)

	if stmt.countExpr == "" {
		if stmt.distinct {
			buf.WriteString("DISTINCT ")
		}
		for _, c := range stmt.cols {
			switch {
			case c == "*",
				// 简单地通过是否包含空格判断是否为多列
				strings.IndexFunc(c, func(r rune) bool { return unicode.IsSpace(r) }) > 0:
				buf.WriteString(c).WriteBytes(',')
			default:
				buf.WriteBytes(stmt.l).
					WriteString(c).
					WriteBytes(stmt.r, ',')
			}
		}
		buf.TruncateLast(1)
	} else {
		buf.WriteString(stmt.countExpr)
	}

	buf.WriteString(" FROM ").
		WriteBytes(stmt.l).
		WriteString(stmt.table).
		WriteBytes(stmt.r)

	// join
	if len(stmt.joins) > 0 {
		buf.WriteBytes(' ')
		for _, join := range stmt.joins {
			buf.WriteString(join.typ).
				WriteString(" JOIN ").
				WriteBytes(stmt.l).
				WriteString(join.table).
				WriteBytes(stmt.r).
				WriteString(" ON ").
				WriteString(join.on).
				WriteBytes(' ')
		}
		buf.TruncateLast(1)
	}

	// where
	wq, wa, err := stmt.where.SQL()
	if err != nil {
		return "", nil, err
	}
	if wq != "" {
		buf.WriteString(" WHERE ")
		buf.WriteString(wq)
		args = append(args, wa...)
	}

	// group by
	if stmt.group != "" {
		buf.WriteString(stmt.group)
	}

	// having
	if stmt.havingQuery != "" {
		buf.WriteString(stmt.havingQuery)
		args = append(args, stmt.havingVals...)
	}

	if stmt.countExpr == "" {
		// order by
		if stmt.orders != nil && stmt.orders.Len() > 0 {
			buf.WriteString(stmt.orders.String())
		}

		// limit
		if stmt.limitQuery != "" {
			buf.WriteString(stmt.limitQuery)
			args = append(args, stmt.limitVals...)
		}
	}

	// for update
	if stmt.forUpdate {
		buf.WriteString(" FOR UPDATE")
	}

	return buf.String(), args, nil
}

// Select 指定列名
func (stmt *SelectStmt) Select(cols ...string) *SelectStmt {
	if stmt.cols == nil {
		stmt.cols = make([]string, 0, len(cols))
	}

	stmt.cols = append(stmt.cols, cols...)
	return stmt
}

// From 指定表名
func (stmt *SelectStmt) From(table string) *SelectStmt {
	stmt.table = table

	return stmt
}

// Having 指定 having 语句
func (stmt *SelectStmt) Having(expr string, args ...interface{}) *SelectStmt {
	stmt.havingQuery = expr
	stmt.havingVals = args

	return stmt
}

// WhereStmt 实现 WhereStmter 接口
func (stmt *SelectStmt) WhereStmt() *WhereStmt {
	return stmt.where
}

// Where 指定 where 语句
func (stmt *SelectStmt) Where(cond string, args ...interface{}) *SelectStmt {
	return stmt.And(cond, args...)
}

// And 指定 where ... AND ... 语句
func (stmt *SelectStmt) And(cond string, args ...interface{}) *SelectStmt {
	stmt.where.And(cond, args...)
	return stmt
}

// Or 指定 where ... OR ... 语句
func (stmt *SelectStmt) Or(cond string, args ...interface{}) *SelectStmt {
	stmt.where.Or(cond, args...)
	return stmt
}

// AndIsNull 指定 WHERE ... AND col IS NULL
func (stmt *SelectStmt) AndIsNull(col string) *SelectStmt {
	stmt.where.And(col + " IS NULL")
	return stmt
}

// OrIsNull 指定 WHERE ... OR col IS NULL
func (stmt *SelectStmt) OrIsNull(col string) *SelectStmt {
	stmt.where.Or(col + " IS NULL")
	return stmt
}

// AndIsNotNull 指定 WHERE ... AND col IS NOT NULL
func (stmt *SelectStmt) AndIsNotNull(col string) *SelectStmt {
	stmt.where.And(col + " IS NOT NULL")
	return stmt
}

// OrIsNotNull 指定 WHERE ... OR col IS NOT NULL
func (stmt *SelectStmt) OrIsNotNull(col string) *SelectStmt {
	stmt.where.Or(col + " IS NOT NULL")
	return stmt
}

// Join 添加一条 Join 语句
func (stmt *SelectStmt) Join(typ, table, on string) *SelectStmt {
	if stmt.joins == nil {
		stmt.joins = make([]*join, 0, 5)
	}

	stmt.joins = append(stmt.joins, &join{typ: typ, table: table, on: on})
	return stmt
}

// Desc 倒序查询
func (stmt *SelectStmt) Desc(col ...string) *SelectStmt {
	return stmt.orderBy(false, col...)
}

// Asc 正序查询
func (stmt *SelectStmt) Asc(col ...string) *SelectStmt {
	return stmt.orderBy(true, col...)
}

func (stmt *SelectStmt) orderBy(asc bool, col ...string) *SelectStmt {
	if stmt.orders == nil {
		stmt.orders = New("")
	}

	if stmt.orders.Len() == 0 {
		stmt.orders.WriteString(" ORDER BY ")
	} else {
		stmt.orders.WriteBytes(',')
	}

	for _, c := range col {
		stmt.orders.WriteBytes(stmt.l).
			WriteString(c).
			WriteBytes(stmt.r, ',')
	}
	stmt.orders.TruncateLast(1)

	if asc {
		stmt.orders.WriteString(" ASC ")
	} else {
		stmt.orders.WriteString(" DESC ")
	}

	return stmt
}

// ForUpdate 添加 FOR UPDATE 语句部分
func (stmt *SelectStmt) ForUpdate() *SelectStmt {
	stmt.forUpdate = true
	return stmt
}

// Group 添加 GROUP BY 语句
func (stmt *SelectStmt) Group(col string) *SelectStmt {
	b := New(" GROUP BY ").WriteBytes(stmt.l).WriteString(col).WriteBytes(stmt.r, ' ')
	stmt.group = b.String()
	return stmt
}

// Limit 生成 SQL 的 Limit 语句
func (stmt *SelectStmt) Limit(limit interface{}, offset ...interface{}) *SelectStmt {
	query, vals := stmt.dialect.LimitSQL(limit, offset...)
	stmt.limitQuery = query
	stmt.limitVals = vals
	return stmt
}

// Count 指定 Count 表示式，如果指定了 count 表达式，则会造成 limit 失效。
//
// 传递空的 expr 参数，表示去除 count 表达式。
// 格式为： count(*) AS cnt
func (stmt *SelectStmt) Count(expr string) *SelectStmt {
	stmt.countExpr = expr
	return stmt
}

// QueryObject 将符合当前条件的所有记录依次写入 objs 中。
//
// 关于 objs 的值类型，可以参考 github.com/issue9/orm/fetch.Object 函数的相关介绍。
func (stmt *SelectStmt) QueryObject(strict bool, objs interface{}) (size int, err error) {
	rows, err := stmt.Query()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err1 := rows.Close(); err1 != nil {
			err = errors.New(err1.Error() + err.Error())
		}
	}()

	return fetch.Object(strict, rows, objs)
}

// QueryString 查询指定列的第一行数据，并将其转换成 string
func (stmt *SelectStmt) QueryString(colName string) (v string, err error) {
	rows, err := stmt.Query()
	if err != nil {
		return "", err
	}
	defer func() {
		if err1 := rows.Close(); err1 != nil {
			err = errors.New(err1.Error() + err.Error())
		}
	}()

	cols, err := fetch.ColumnString(true, colName, rows)
	if err != nil {
		return "", err
	}

	if len(cols) == 0 {
		return "", ErrNoData
	}

	return cols[0], nil
}

// QueryFloat 查询指定列的第一行数据，并将其转换成 float64
func (stmt *SelectStmt) QueryFloat(colName string) (float64, error) {
	v, err := stmt.QueryString(colName)
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(v, 10)
}

// QueryInt 查询指定列的第一行数据，并将其转换成 int64
func (stmt *SelectStmt) QueryInt(colName string) (int64, error) {
	// NOTE: 可能会出现浮点数的情况。比如：
	// select avg(xx) as avg form xxx where xxx
	// 查询 avg 的值可能是 5.000 等值。
	v, err := stmt.QueryString(colName)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(v, 10, 64)
}
