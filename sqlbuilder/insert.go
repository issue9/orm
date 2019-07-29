// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
	"errors"

	"github.com/issue9/orm/v2/core"
)

// InsertStmt 表示插入操作的 SQL 语句
type InsertStmt struct {
	*execStmt

	table      string
	cols       []string
	args       [][]interface{}
	selectStmt *SelectStmt
}

// Insert 声明一条插入语句
func Insert(e core.Engine) *InsertStmt {
	stmt := &InsertStmt{
		cols: make([]string, 0, 10),
		args: make([][]interface{}, 0, 10),
	}
	stmt.execStmt = newExecStmt(e, stmt)

	return stmt
}

// Insert 将当前查询结果作为 Insert 的值
//
// 构建 insert into (...) select .... 语句
func (stmt *SelectStmt) Insert() *InsertStmt {
	insert := Insert(stmt.Engine())
	return insert.Select(stmt)
}

// Select 当前插入数据从 Select 中获取
//
// 构建 insert into (...) select .... 语句
func (stmt *InsertStmt) Select(sel *SelectStmt) *InsertStmt {
	stmt.selectStmt = sel
	return stmt
}

// Table 指定表名
func (stmt *InsertStmt) Table(table string) *InsertStmt {
	stmt.table = table
	return stmt
}

// KeyValue 指定键值对
//
// 当通过 Values() 指定多行数据时，再使用 KeyValue 会出错
func (stmt *InsertStmt) KeyValue(col string, val interface{}) *InsertStmt {
	if len(stmt.args) > 1 {
		stmt.err = errors.New("多列模式，不能调用 KeyValue 函数")
	}

	if len(stmt.args) == 0 {
		stmt.args = append(stmt.args, []interface{}{})
	}

	stmt.cols = append(stmt.cols, col)
	stmt.args[0] = append(stmt.args[0], val)

	return stmt
}

// Columns 指定插入的列，多次指定，之前的会被覆盖。
func (stmt *InsertStmt) Columns(cols ...string) *InsertStmt {
	stmt.cols = append(stmt.cols, cols...)
	return stmt
}

// Values 指定需要插入的值
//
// NOTE: vals 传入时，并不会被解压
func (stmt *InsertStmt) Values(vals ...interface{}) *InsertStmt {
	stmt.args = append(stmt.args, vals)
	return stmt
}

// Reset 重置语句
func (stmt *InsertStmt) Reset() *InsertStmt {
	stmt.baseStmt.Reset()
	stmt.table = ""
	stmt.cols = stmt.cols[:0]
	stmt.args = stmt.args[:0]
	stmt.selectStmt = nil
	return stmt
}

// InsertDefaultValueHooker 插入值全部为默认值时的钩子处理函数
type InsertDefaultValueHooker interface {
	InsertDefaultValueHook(tableName string) (string, []interface{}, error)
}

// SQL 获取 SQL 的语句及参数部分
func (stmt *InsertStmt) SQL() (string, []interface{}, error) {
	if stmt.err != nil {
		return "", nil, stmt.Err()
	}

	if stmt.table == "" {
		return "", nil, ErrTableIsEmpty
	}

	/*if len(stmt.cols) == 0 && (stmt.selectStmt == nil || len(stmt.selectStmt.columns) == 0) {
		return "", nil, ErrColumnsIsEmpty
	}

	if len(stmt.args) == 0 && stmt.selectStmt == nil {
		return "", nil, ErrValueIsEmpty
	}
	*/

	builder := core.NewBuilder("INSERT INTO ").WriteString(stmt.table)

	if stmt.selectStmt != nil {
		return stmt.fromSelect(builder)
	}

	if len(stmt.cols) == 0 && len(stmt.args) == 0 {
		return stmt.insertDefault(builder)
	}

	for _, vals := range stmt.args {
		if len(vals) != len(stmt.cols) {
			return "", nil, ErrArgsNotMatch
		}
	}

	builder.WriteBytes('(')
	for _, col := range stmt.cols {
		builder.QuoteKey(col).
			WriteBytes(',')
	}
	builder.TruncateLast(1)
	builder.WriteBytes(')')

	args := make([]interface{}, 0, len(stmt.cols)*len(stmt.args))
	builder.WriteString(" VALUES ")
	for _, vals := range stmt.args {
		builder.WriteBytes('(')
		for _, v := range vals {
			if named, ok := v.(sql.NamedArg); ok && named.Name != "" {
				builder.WriteBytes('@')
				builder.WriteString(named.Name)
			} else {
				builder.WriteBytes('?')
			}
			builder.WriteBytes(',')
			args = append(args, v)
		}
		builder.TruncateLast(1) // 去掉最后的逗号
		builder.WriteString("),")
	}
	builder.TruncateLast(1)

	query, err := builder.String()
	if err != nil {
		return "", nil, err
	}
	return query, args, nil
}

func (stmt *InsertStmt) insertDefault(builder *core.Builder) (string, []interface{}, error) {
	if hook, ok := stmt.Dialect().(InsertDefaultValueHooker); ok {
		return hook.InsertDefaultValueHook(stmt.table)
	}

	query, err := builder.WriteString(" DEFAULT VALUES").String()
	if err != nil {
		return "", nil, err
	}

	return query, nil, nil
}

func (stmt *InsertStmt) fromSelect(builder *core.Builder) (string, []interface{}, error) {
	builder.WriteBytes('(')
	if len(stmt.cols) > 0 {
		for _, col := range stmt.cols {
			builder.QuoteKey(col).WriteBytes(',')
		}
		builder.TruncateLast(1)
	} else {
		for _, col := range stmt.selectStmt.columns {
			builder.WriteString(getColumnName(col)).WriteBytes(',')
		}
		builder.TruncateLast(1)
	}
	builder.WriteBytes(')')

	query, args, err := stmt.selectStmt.SQL()
	if err != nil {
		return "", nil, err
	}

	q, err := builder.WriteString(query).String()
	if err != nil {
		return "", nil, err
	}
	return q, args, nil
}

// LastInsertID 执行 SQL 语句
//
// 并根据表名和自增列 ID 返回当前行的自增 ID 值。
//
// NOTE: 对于指定了自增值的，其结果是未知的。
func (stmt *InsertStmt) LastInsertID(table, col string) (int64, error) {
	return stmt.LastInsertIDContext(context.Background(), table, col)
}

// LastInsertIDContext 执行 SQL 语句
//
// 并根据表名和自增列 ID 返回当前行的自增 ID 值。
func (stmt *InsertStmt) LastInsertIDContext(ctx context.Context, table, col string) (id int64, err error) {
	if len(stmt.args) > 1 {
		// mysql 没有好的方法可以处理多行插入数据时，返回最大的 ID 值。
		return 0, errors.New("多行插入语句，无法获取 LastInsertIDContext")
	}

	sql, append := stmt.Dialect().LastInsertIDSQL(stmt.table, col)
	if sql == "" {
		rslt, err := stmt.ExecContext(ctx)
		if err != nil {
			return 0, err
		}

		return rslt.LastInsertId()
	}

	var args []interface{}
	if !append {
		_, err := stmt.ExecContext(ctx)
		if err != nil {
			return 0, err
		}

		err = stmt.engine.QueryRowContext(ctx, sql, args...).Scan(&id)
		return id, err
	}

	// 当 append 为 true 时，将 sql 添加到 query 之后，合并为一条语句
	query, as, err := stmt.SQL()
	if err != nil {
		return 0, err
	}
	sql = query + sql
	args = as

	sql, args, err = stmt.Dialect().SQL(sql, args)
	if err != nil {
		return 0, err
	}
	err = stmt.engine.QueryRowContext(ctx, sql, args...).Scan(&id)
	return
}
