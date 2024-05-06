// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package sqlbuilder

import (
	"context"
	"database/sql"
	"errors"

	"github.com/issue9/orm/v6/core"
	"github.com/issue9/orm/v6/fetch"
)

// ErrNoData 在 [Select.QueryInt] 等函数中，
// 如果没有符合条件的数据，则返回此错误。
var ErrNoData = errors.New("不存在符合和条件的数据")

// SelectQuery 预编译之后的查询语句
type SelectQuery struct {
	stmt *core.Stmt
}

type selectWhere = WhereStmtOf[*SelectStmt]

// SelectStmt 查询语句
type SelectStmt struct {
	*queryStmt
	*selectWhere

	tableExpr string
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
	havingVals  []any

	limitQuery string
	limitVals  []any
}

type unionSelect struct {
	typ  string
	stmt *SelectStmt
}

// Select 生成插入语句
func (sql *SQLBuilder) Select() *SelectStmt { return Select(sql.engine) }

// Select 声明一条 SELECT 语句
func Select(e core.Engine) *SelectStmt {
	stmt := &SelectStmt{columns: make([]string, 0, 10)}
	stmt.queryStmt = newQueryStmt(e, stmt)
	stmt.selectWhere = NewWhereStmtOf(stmt)
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
	stmt.WhereStmt().Reset()
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
func (stmt *SelectStmt) SQL() (string, []any, error) {
	if stmt.err != nil {
		return "", nil, stmt.Err()
	}

	if stmt.tableExpr == "" {
		return "", nil, SyntaxError("SELECT", "未指定表名")
	}

	builder := core.NewBuilder("SELECT ")
	args := make([]any, 0, 10)

	stmt.buildColumns(builder)

	builder.WString(" FROM ").WString(stmt.tableExpr)

	// join
	if stmt.joins != nil {
		builder.Append(stmt.joins).WBytes(' ')
	}

	// where
	wq, wa, err := stmt.WhereStmt().SQL()
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

func (stmt *SelectStmt) buildUnions(builder *core.Builder) (args []any, err error) {
	l := len(stmt.columns)

	for _, u := range stmt.unions {
		if len(u.stmt.columns) != l {
			return nil, SyntaxError("SELECT", "union 各个 select 的列数量不相同")
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

	if len(stmt.columns) == 0 {
		builder.WBytes('*')
		return
	}

	for _, col := range stmt.columns {
		builder.WString(col).WBytes(',')
	}
	builder.TruncateLast(1)
}

// Column 指定列
//
// 一次只能指定一列，当未指定任何列时，默认会采用 *。
//
// col 表示列名，可以是以下形式：
//
//	*
//	col
//	table.col
//	table.*
//	sum({table}.{col}) as col1
//
// 如果列名是关键字，可以使用 {} 包含。如果包含了表名，则需要自行添加表名前缀。
func (stmt *SelectStmt) Column(col string) *SelectStmt {
	stmt.columns = append(stmt.columns, col)
	return stmt
}

// Columns 指定列名
//
// 相当于按参数顺序依次调用 [Select.Column]，如果存在别名，
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
func (stmt *SelectStmt) Having(expr string, args ...any) *SelectStmt {
	stmt.havingQuery = expr
	stmt.havingVals = args
	return stmt
}

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
// col 为分组的列名，格式可以是单纯的列名，或是带表名的列：
//
//	col
//	table.col
//
// table 和 col 都可以是关键字，系统会自动处理。
func (stmt *SelectStmt) Desc(col ...string) *SelectStmt {
	return stmt.orderBy(false, col...)
}

// Asc 正序查询
//
// col 为分组的列名，格式可以是单纯的列名，或是带表名的列：
//
//	col
//	table.col
//
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
//
//	col
//	table.col
//
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
func (stmt *SelectStmt) Limit(limit any, offset ...any) *SelectStmt {
	query, vals := stmt.Dialect().LimitSQL(limit, offset...)
	stmt.limitQuery = query
	stmt.limitVals = vals
	return stmt
}

// Count 指定 Count 表达式
//
// 如果指定了 count 表达式，则会造成 limit 失效，
// 如果设置为空值，则取消 count，恢复普通的 select 。
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

// QueryObject 将符合当前条件的所有记录依次写入 objs 中
//
// 关于 objs 的类型，可以参考 [fetch.Object] 函数的相关介绍。
func (stmt *SelectStmt) QueryObject(strict bool, objs any) (size int, err error) {
	return stmt.QueryObjectContext(context.Background(), strict, objs)
}

func (stmt *SelectStmt) QueryObjectContext(ctx context.Context, strict bool, objs any) (size int, err error) {
	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return 0, err
	}
	return fetchObject(rows, strict, objs)
}

// QueryString 查询指定列的第一行数据，并将其转换成 string
func (stmt *SelectStmt) QueryString(colName string) (v string, err error) {
	return stmt.QueryStringContext(context.Background(), colName)
}

func (stmt *SelectStmt) QueryStringContext(ctx context.Context, colName string) (v string, err error) {
	return fetchSelectStmtColumn[string](stmt, ctx, colName)
}

// QueryFloat 查询指定列的第一行数据，并将其转换成 float64
func (stmt *SelectStmt) QueryFloat(colName string) (float64, error) {
	return stmt.QueryFloatContext(context.Background(), colName)
}

func (stmt *SelectStmt) QueryFloatContext(ctx context.Context, colName string) (float64, error) {
	return fetchSelectStmtColumn[float64](stmt, ctx, colName)
}

// QueryInt 查询指定列的第一行数据，并将其转换成 int64
func (stmt *SelectStmt) QueryInt(colName string) (int64, error) {
	return stmt.QueryIntContext(context.Background(), colName)
}

func (stmt *SelectStmt) QueryIntContext(ctx context.Context, colName string) (int64, error) {
	return fetchSelectStmtColumn[int64](stmt, ctx, colName)
}

// Select 生成 select 语句
func (stmt *WhereStmt) Select(e core.Engine) *SelectStmt {
	sel := Select(e)
	sel.selectWhere.w = stmt
	return sel
}

func (stmt *SelectStmt) PrepareContext(ctx context.Context) (*SelectQuery, error) {
	s, err := stmt.queryStmt.PrepareContext(ctx)
	if err != nil {
		return nil, err
	}
	return &SelectQuery{stmt: s}, nil
}

func (stmt *SelectStmt) Prepare() (*SelectQuery, error) {
	return stmt.PrepareContext(context.Background())
}

// QueryObject 将符合当前条件的所有记录依次写入 objs 中。
//
// 关于 objs 的值类型，可以参考 github.com/issue9/orm/fetch.Object 函数的相关介绍。
func (stmt *SelectQuery) QueryObject(strict bool, objs any, arg ...any) (size int, err error) {
	rows, err := stmt.stmt.Query(arg...)
	if err != nil {
		return 0, err
	}
	return fetchObject(rows, strict, objs)
}

// QueryString 查询指定列的第一行数据，并将其转换成 string
func (stmt *SelectQuery) QueryString(colName string, arg ...any) (string, error) {
	return fetchSelectQueryColumn[string](stmt, colName, arg...)
}

// QueryFloat 查询指定列的第一行数据，并将其转换成 float64
func (stmt *SelectQuery) QueryFloat(colName string, arg ...any) (float64, error) {
	return fetchSelectQueryColumn[float64](stmt, colName, arg...)
}

// QueryInt 查询指定列的第一行数据，并将其转换成 int64
func (stmt *SelectQuery) QueryInt(colName string, arg ...any) (int64, error) {
	return fetchSelectQueryColumn[int64](stmt, colName, arg...)
}

func (stmt *SelectQuery) Close() error { return stmt.stmt.Close() }

func fetchObject(rows *sql.Rows, strict bool, objs any) (size int, err error) {
	defer func() { err = errors.Join(err, rows.Close()) }()
	size, err = fetch.Object(strict, rows, objs)
	return // 注意 defer，独立为一行
}

func fetchSelectStmtColumn[T any](stmt *SelectStmt, ctx context.Context, colName string) (v T, err error) {
	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return v, err
	}
	return fetchColumn[T](rows, colName)
}

func fetchSelectQueryColumn[T any](stmt *SelectQuery, colName string, arg ...any) (v T, err error) {
	rows, err := stmt.stmt.Query(arg...)
	if err != nil {
		return v, err
	}
	return fetchColumn[T](rows, colName)
}

func fetchColumn[T any](rows *sql.Rows, colName string) (v T, err error) {
	defer func() { err = errors.Join(err, rows.Close()) }()

	cols, err := fetch.Column[T](true, colName, rows)
	if err != nil {
		return v, err
	}

	if len(cols) == 0 {
		return v, ErrNoData
	}

	return cols[0], nil
}
