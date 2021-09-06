// SPDX-License-Identifier: MIT

package sqlbuilder

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/issue9/orm/v4/core"
	"github.com/issue9/orm/v4/fetch"
)

// ErrNoData 在 Select.QueryInt 等函数中，
// 如果没有符合条件的数据，则返回此错误。
var ErrNoData = errors.New("不存在符合和条件的数据")

// SelectStmt 查询语句
type SelectStmt struct {
	*queryStmt

	tableExpr string
	where     *WhereStmt
	columns   []string
	distinct  bool
	forUpdate bool

	// COUNT 查询的列内容
	countExpr string

	unions []*unionSelect

	joins  *core.Builder
	orders *core.Builder
	group  string

	havingQuery string
	havingVals  []interface{}

	limitQuery string
	limitVals  []interface{}
}

type unionSelect struct {
	typ  string
	stmt *SelectStmt
}

// Select 生成插入语句
func (sql *SQLBuilder) Select() *SelectStmt { return Select(sql.engine) }

// Select 声明一条 Select 语句
func Select(e core.Engine) *SelectStmt {
	stmt := &SelectStmt{columns: make([]string, 0, 10)}
	stmt.queryStmt = newQueryStmt(e, stmt)
	stmt.where = Where()

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
	stmt.baseStmt.Reset()

	stmt.tableExpr = ""
	stmt.where.Reset()
	stmt.columns = stmt.columns[:0]
	stmt.distinct = false
	stmt.forUpdate = false

	stmt.countExpr = ""

	if stmt.unions != nil {
		stmt.unions = stmt.unions[:0]
	}

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
	if stmt.err != nil {
		return "", nil, stmt.Err()
	}

	if stmt.tableExpr == "" {
		return "", nil, ErrTableIsEmpty
	}

	if len(stmt.columns) == 0 && stmt.countExpr == "" {
		return "", nil, ErrColumnsIsEmpty
	}

	builder := core.NewBuilder("SELECT ")
	args := make([]interface{}, 0, 10)

	stmt.buildColumns(builder)

	builder.WString(" FROM ").WString(stmt.tableExpr)

	// join
	if stmt.joins != nil {
		builder.Append(stmt.joins).WBytes(' ')
	}

	// where
	wq, wa, err := stmt.where.SQL()
	if err != nil {
		return "", nil, err
	}
	if wq != "" {
		builder.WString(" WHERE ")
		builder.WString(wq)
		args = append(args, wa...)
	}

	// group by
	if stmt.group != "" {
		builder.WString(stmt.group)
	}

	// having
	if stmt.havingQuery != "" {
		builder.WString(stmt.havingQuery)
		args = append(args, stmt.havingVals...)
	}

	if stmt.countExpr == "" {
		// order by
		if stmt.orders != nil && stmt.orders.Len() > 0 {
			builder.Append(stmt.orders)
		}

		// limit
		if stmt.limitQuery != "" {
			builder.WString(stmt.limitQuery)
			args = append(args, stmt.limitVals...)
		}
	}

	// union
	if stmt.unions != nil {
		a, err := stmt.buildUnions(builder)
		if err != nil {
			return "", nil, err
		}
		args = append(args, a...)
	}

	// for update
	if stmt.forUpdate {
		builder.WString(" FOR UPDATE")
	}

	query, err := builder.String()
	if err != nil {
		return "", nil, err
	}
	return query, args, nil
}

func (stmt *SelectStmt) buildUnions(builder *core.Builder) (args []interface{}, err error) {
	l := len(stmt.columns)

	for _, u := range stmt.unions {
		if len(u.stmt.columns) != l {
			return nil, ErrUnionColumnNotMatch
		}

		query, a, err := u.stmt.SQL()
		if err != nil {
			return nil, err
		}

		builder.WString(u.typ).WBytes(' ').WString(query)
		args = append(args, a...)
	}

	return args, nil
}

func (stmt *SelectStmt) buildColumns(builder *core.Builder) {
	if stmt.countExpr != "" {
		builder.WString(stmt.countExpr)
		return
	}

	if stmt.distinct {
		builder.WString("DISTINCT ")
	}

	for _, col := range stmt.columns {
		builder.WString(col).WBytes(',')
	}

	builder.TruncateLast(1)
}

// Column 指定列，一次只能指定一列。
//
// col 表示列名，可以是以下形式：
//  *
//  col
//  table.col
//  table.*
//  sum({table}.{col}) as col1
//
// 如果列名是关键字，可以使用 {} 包含。
func (stmt *SelectStmt) Column(col string) *SelectStmt {
	stmt.columns = append(stmt.columns, col)
	return stmt
}

// Columns 指定列名，可以指定多列。
//
// 相当于按参数顺序依次调用 Select.Column，如果存在别名，
// 可以使用 col AS alias 的方式指定每一个参数。
//
// 如果列名是关键字，可以使用 {} 包含。
func (stmt *SelectStmt) Columns(cols ...string) *SelectStmt {
	for _, col := range cols {
		stmt.Column(col)
	}
	return stmt
}

// From 指定表名
//
// table 为表名，如果需要指定别名，可以通过 alias 指定。
func (stmt *SelectStmt) From(table string, alias ...string) *SelectStmt {
	if stmt.err != nil {
		return stmt
	}

	if stmt.tableExpr != "" {
		stmt.err = errors.New("不能重复指定表名")
		return stmt
	}

	builder := core.NewBuilder("").QuoteKey(table)

	switch len(alias) {
	case 0:
		stmt.tableExpr, stmt.err = builder.String()
	case 1:
		if alias[0] == "" {
			break
		}

		builder.WString(" AS ").QuoteKey(alias[0])
		stmt.tableExpr, stmt.err = builder.String()
	default:
		stmt.err = errors.New("过多的 alias 参数")
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
func (stmt *SelectStmt) WhereStmt() *WhereStmt { return stmt.where }

// Join 添加一条 Join 语句
func (stmt *SelectStmt) Join(typ, table, alias, on string) *SelectStmt {
	if stmt.joins == nil {
		stmt.joins = core.NewBuilder("")
	}

	stmt.joins.WBytes(' ').
		WString(typ).
		WString(" JOIN ").
		QuoteKey(table).
		WString(" AS ").
		QuoteKey(alias).
		WString(" ON ").
		WString(on)

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
		stmt.orders = core.NewBuilder("")
	}

	if stmt.orders.Len() == 0 {
		stmt.orders.WString(" ORDER BY ")
	} else {
		stmt.orders.WBytes(',')
	}

	for _, c := range col {
		stmt.orders.WString(c).WBytes(',')
	}
	stmt.orders.TruncateLast(1)

	if asc {
		stmt.orders.WString(" ASC ")
	} else {
		stmt.orders.WString(" DESC ")
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
	if stmt.err != nil {
		return stmt
	}

	stmt.group, stmt.err = core.NewBuilder(" GROUP BY ").
		WString(col).
		WBytes(' ').
		String()
	return stmt
}

// Limit 生成 SQL 的 Limit 语句
func (stmt *SelectStmt) Limit(limit interface{}, offset ...interface{}) *SelectStmt {
	query, vals := stmt.Dialect().LimitSQL(limit, offset...)
	stmt.limitQuery = query
	stmt.limitVals = vals
	return stmt
}

// Count 指定 Count 表达式，如果指定了 count 表达式，则会造成 limit 失效。
//
// 如果设置为空值，则取消 count，恢复普通的 select
func (stmt *SelectStmt) Count(expr string) *SelectStmt {
	stmt.countExpr = expr

	return stmt
}

// Union 语句
//
// all 表示是否执行 Union all 语法；
// sel 表示需要进行并接的 Select 语句，传入 sel 之后，
// 后续对 sel 的操作依赖会影响到语句的最终生成。
func (stmt *SelectStmt) Union(all bool, sel ...*SelectStmt) *SelectStmt {
	if stmt.unions == nil {
		stmt.unions = make([]*unionSelect, 0, len(sel))
	}

	typ := " UNION "
	if all {
		typ += "ALL "
	}

	for _, s := range sel {
		stmt.unions = append(stmt.unions, &unionSelect{
			typ:  typ,
			stmt: s,
		})
	}

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
			err = fmt.Errorf("在抛出错误 %s 时再次发生错误 %w", err.Error(), err1)
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
			err = fmt.Errorf("在抛出错误 %s 时再次发生错误 %w", err.Error(), err1)
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

	return strconv.ParseFloat(v, 64)
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

// Where SelectStmt.And 的别名
func (stmt *SelectStmt) Where(cond string, args ...interface{}) *SelectStmt {
	return stmt.And(cond, args...)
}

// And 添加一条 and 语句
func (stmt *SelectStmt) And(cond string, args ...interface{}) *SelectStmt {
	stmt.where.And(cond, args...)
	return stmt
}

// Or 添加一条 OR 语句
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

// AndBetween 指定 WHERE ... AND col BETWEEN v1 AND v2
func (stmt *SelectStmt) AndBetween(col string, v1, v2 interface{}) *SelectStmt {
	stmt.where.AndBetween(col, v1, v2)
	return stmt
}

// OrBetween 指定 WHERE ... OR col BETWEEN v1 AND v2
func (stmt *SelectStmt) OrBetween(col string, v1, v2 interface{}) *SelectStmt {
	stmt.where.OrBetween(col, v1, v2)
	return stmt
}

// AndNotBetween 指定 WHERE ... AND col NOT BETWEEN v1 AND v2
func (stmt *SelectStmt) AndNotBetween(col string, v1, v2 interface{}) *SelectStmt {
	stmt.where.AndNotBetween(col, v1, v2)
	return stmt
}

// OrNotBetween 指定 WHERE ... OR col BETWEEN v1 AND v2
func (stmt *SelectStmt) OrNotBetween(col string, v1, v2 interface{}) *SelectStmt {
	stmt.where.OrNotBetween(col, v1, v2)
	return stmt
}

// AndLike 指定 WHERE ... AND col LIKE content
func (stmt *SelectStmt) AndLike(col string, content interface{}) *SelectStmt {
	stmt.where.AndLike(col, content)
	return stmt
}

// OrLike 指定 WHERE ... OR col LIKE content
func (stmt *SelectStmt) OrLike(col string, content interface{}) *SelectStmt {
	stmt.where.OrLike(col, content)
	return stmt
}

// AndNotLike 指定 WHERE ... AND col NOT LIKE content
func (stmt *SelectStmt) AndNotLike(col string, content interface{}) *SelectStmt {
	stmt.where.AndNotLike(col, content)
	return stmt
}

// OrNotLike 指定 WHERE ... OR col NOT LIKE content
func (stmt *SelectStmt) OrNotLike(col string, content interface{}) *SelectStmt {
	stmt.where.OrNotLike(col, content)
	return stmt
}

// AndIn 指定 WHERE ... AND col IN(v...)
func (stmt *SelectStmt) AndIn(col string, v ...interface{}) *SelectStmt {
	stmt.where.AndIn(col, v...)
	return stmt
}

// OrIn 指定 WHERE ... OR col IN(v...)
func (stmt *SelectStmt) OrIn(col string, v ...interface{}) *SelectStmt {
	stmt.where.OrIn(col, v...)
	return stmt
}

// AndNotIn 指定 WHERE ... AND col NOT IN(v...)
func (stmt *SelectStmt) AndNotIn(col string, v ...interface{}) *SelectStmt {
	stmt.where.AndNotIn(col, v...)
	return stmt
}

// OrNotIn 指定 WHERE ... OR col IN(v...)
func (stmt *SelectStmt) OrNotIn(col string, v ...interface{}) *SelectStmt {
	stmt.where.OrNotIn(col, v...)
	return stmt
}

// AndGroup 开始一个子条件语句
func (stmt *SelectStmt) AndGroup() *WhereStmt { return stmt.where.AndGroup() }

// OrGroup 开始一个子条件语句
func (stmt *SelectStmt) OrGroup() *WhereStmt { return stmt.where.OrGroup() }

// Select 生成 select 语句
func (stmt *WhereStmt) Select(e core.Engine) *SelectStmt {
	sel := Select(e)
	sel.where = stmt
	return sel
}
