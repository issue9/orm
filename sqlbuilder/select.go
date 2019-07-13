// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"errors"
	"strconv"
	"strings"

	"github.com/issue9/orm/v2/fetch"
)

// ErrNoData 在 Select.QueryInt 等函数中，
// 如果没有符合条件的数据，则返回此错误。
var ErrNoData = errors.New("不存在符合和条件的数据")

// SelectStmt 查询语句
type SelectStmt struct {
	*queryStmt

	tableExpr string
	where     *WhereStmt
	columns   []*column
	distinct  bool
	forUpdate bool

	// COUNT 查询的列内容
	countExpr string

	joins  *SQLBuilder
	orders *SQLBuilder
	group  string

	havingQuery string
	havingVals  []interface{}

	limitQuery string
	limitVals  []interface{}
}

type column struct {
	// 分别表示列前的表名，列名以及列的别名
	table, name, alias string
}

// Select 声明一条 Select 语句
func Select(e Engine, d Dialect) *SelectStmt {
	stmt := &SelectStmt{
		where:   Where(),
		columns: make([]*column, 0, 10),
	}
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
	stmt.tableExpr = ""
	stmt.where.Reset()
	stmt.columns = stmt.columns[:0]
	stmt.distinct = false
	stmt.forUpdate = false

	stmt.countExpr = ""

	if stmt.joins != nil {
		stmt.joins.Reset()
	}
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
	if stmt.tableExpr == "" {
		return "", nil, ErrTableIsEmpty
	}

	if len(stmt.columns) == 0 && stmt.countExpr == "" {
		return "", nil, ErrColumnsIsEmpty
	}

	buf := New("SELECT ")
	args := make([]interface{}, 0, 10)

	stmt.buildColumns(buf)

	buf.WriteString(" FROM ").WriteString(stmt.tableExpr)

	// join
	if stmt.joins != nil {
		buf.Append(stmt.joins).WriteBytes(' ')
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
			buf.Append(stmt.orders)
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

func (stmt *SelectStmt) buildColumns(builder *SQLBuilder) {
	if stmt.countExpr != "" {
		builder.WriteString(stmt.countExpr)
		return
	}

	if stmt.distinct {
		builder.WriteString("DISTINCT ")
	}

	for _, col := range stmt.columns {
		if col.table != "" {
			builder.Quote(col.table, stmt.l, stmt.r).WriteBytes('.')
		}

		if col.name == "*" {
			builder.WriteBytes('*')
		} else {
			builder.Quote(col.name, stmt.l, stmt.r)
		}

		if col.alias != "" {
			builder.WriteString(" AS ").Quote(col.alias, stmt.l, stmt.r)
		}

		builder.WriteBytes(',')
	}

	builder.TruncateLast(1)
}

// Column 指定列，一次只能指定一列，别外可使用 alias 参数
//
// col 表示列名，可以是以下形式：
//  *
//  col
//  table.col
//  table.*
func (stmt *SelectStmt) Column(col string, alias ...string) *SelectStmt {
	c := &column{name: col}
	if index := strings.IndexByte(col, '.'); index > 0 {
		c.table = col[:index]
		c.name = col[index+1:]
	}

	switch len(alias) {
	case 0:
	case 1:
		if alias[0] == "" {
			break
		}
		c.alias = alias[0]
	default:
		panic("过多的别名参数")
	}

	stmt.columns = append(stmt.columns, c)

	return stmt
}

// AliasColumns 同时指定多列，必须存在别名。
//
// 参数中，键名为别名，键值为列名。
func (stmt *SelectStmt) AliasColumns(cols map[string]string) *SelectStmt {
	for alias, col := range cols {
		stmt.Column(col, alias)
	}

	return stmt
}

// Columns 指定列名，可以指定多列。
//
// 相当于按参数顺序依次调用 Select.Column，如果存在别名，
// 可以使用 col AS alias 的方式指定每一个参数。
func (stmt *SelectStmt) Columns(cols ...string) *SelectStmt {
	for _, col := range cols {
		if len(col) <= 6 { // " AS " 再加上前后最少一个字符，最少 6 个字符
			stmt.Column(col)
			continue
		}

		col, alias := splitWithAS(col)
		stmt.Column(col, alias)
	}
	return stmt
}

// From 指定表名
//
// table 为表名，如果需要指定别名，可以通过 alias 指定。
func (stmt *SelectStmt) From(table string, alias ...string) *SelectStmt {
	if stmt.tableExpr != "" {
		panic("不能重复指定表名")
	}

	builder := New("").Quote(table, stmt.l, stmt.r)

	switch len(alias) {
	case 0:
		stmt.tableExpr = builder.String()
	case 1:
		if alias[0] == "" {
			break
		}

		builder.WriteString(" AS ").Quote(alias[0], stmt.l, stmt.r)
		stmt.tableExpr = builder.String()
	default:
		panic("过多的 alias 参数")
	}

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
	stmt.where.AndIsNull(col)
	return stmt
}

// OrIsNull 指定 WHERE ... OR col IS NULL
func (stmt *SelectStmt) OrIsNull(col string) *SelectStmt {
	stmt.where.OrIsNull(col)
	return stmt
}

// AndIsNotNull 指定 WHERE ... AND col IS NOT NULL
func (stmt *SelectStmt) AndIsNotNull(col string) *SelectStmt {
	stmt.where.AndIsNotNull(col)
	return stmt
}

// OrIsNotNull 指定 WHERE ... OR col IS NOT NULL
func (stmt *SelectStmt) OrIsNotNull(col string) *SelectStmt {
	stmt.where.OrIsNotNull(col)
	return stmt
}

// Join 添加一条 Join 语句
func (stmt *SelectStmt) Join(typ, table, alias, on string) *SelectStmt {
	if stmt.joins == nil {
		stmt.joins = New("")
	}

	stmt.joins.WriteBytes(' ').
		WriteString(typ).
		WriteString(" JOIN ").
		Quote(table, stmt.l, stmt.r).
		WriteString(" AS ").
		Quote(alias, stmt.l, stmt.r).
		WriteString(" ON ").
		WriteString(on)

	return stmt
}

// Desc 倒序查询
//
// col 为分组的列名，格式可以单纯的列名，或是带表名的列：
//  col
//  table.col
// table 和 col 都可以是关键字，系统会自动处理。
func (stmt *SelectStmt) Desc(col ...string) *SelectStmt {
	return stmt.orderBy(false, col...)
}

// Asc 正序查询
//
// col 为分组的列名，格式可以单纯的列名，或是带表名的列：
//  col
//  table.col
// table 和 col 都可以是关键字，系统会自动处理。
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
		stmt.quoteColumn(stmt.orders, c)
		stmt.orders.WriteBytes(',')
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
//
// col 为分组的列名，格式可以单纯的列名，或是带表名的列：
//  col
//  table.col
// table 和 col 都可以是关键字，系统会自动处理。
func (stmt *SelectStmt) Group(col string) *SelectStmt {
	b := New(" GROUP BY ")
	stmt.quoteColumn(b, col)
	stmt.group = b.String()
	return stmt
}

// 为列名添加数据库专属的引号，列名可以带表名前缀。
func (stmt *baseStmt) quoteColumn(b *SQLBuilder, col string) {
	index := strings.IndexByte(col, ',')
	if index <= 0 {
		b.Quote(col, stmt.l, stmt.r).WriteBytes(' ')
	} else {
		b.Quote(col[:index], stmt.l, stmt.r).
			WriteBytes(' ').
			Quote(col[index+1:], stmt.l, stmt.r).
			WriteBytes(' ')
	}
}

// Limit 生成 SQL 的 Limit 语句
func (stmt *SelectStmt) Limit(limit interface{}, offset ...interface{}) *SelectStmt {
	query, vals := stmt.dialect.LimitSQL(limit, offset...)
	stmt.limitQuery = query
	stmt.limitVals = vals
	return stmt
}

// ResetCount 重置 Count
func (stmt *SelectStmt) ResetCount() *SelectStmt {
	stmt.countExpr = ""
	return stmt
}

// Count 指定 Count 表达式，如果指定了 count 表达式，则会造成 limit 失效。
//
// 会被拼接成以下格式：
//  COUNT(DISTINCT col) AS cnt
func (stmt *SelectStmt) Count(cnt, col string, distinct bool) *SelectStmt {
	builder := New("COUNT(")

	if distinct {
		builder.WriteString(" DISTINCT ")
	}

	if col == "*" {
		builder.WriteString("*)")
	} else {
		builder.Quote(col, stmt.l, stmt.r).
			WriteBytes(')')
	}

	builder.WriteString(" AS ").Quote(cnt, stmt.l, stmt.r)

	stmt.countExpr = builder.String()

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
