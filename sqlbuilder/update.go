// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
	"sort"

	"github.com/issue9/orm/types"
)

// UpdateStmt 更新语句
type UpdateStmt struct {
	engine types.Engine
	table  string
	where  *WhereStmt
	values []*updateSet
}

type updateSet struct {
	column string
	value  interface{}
	typ    byte // 类型，可以是 + 自增类型，- 自减类型，或是空值表示正常表达式
}

// Update 声明一条 UPDATE 的 SQL 语句
func Update(e types.Engine) *UpdateStmt {
	return &UpdateStmt{
		engine: e,
		where:  newWhereStmt(),
		values: []*updateSet{},
	}
}

// Table 指定表名
func (stmt *UpdateStmt) Table(table string) *UpdateStmt {
	stmt.table = table
	return stmt
}

// Set 设置值，若 col 相同，则会覆盖
func (stmt *UpdateStmt) Set(col string, val interface{}) *UpdateStmt {
	stmt.values = append(stmt.values, &updateSet{
		column: col,
		value:  val,
		typ:    0,
	})
	return stmt
}

// Increase 给列增加值
func (stmt *UpdateStmt) Increase(col string, val interface{}) *UpdateStmt {
	stmt.values = append(stmt.values, &updateSet{
		column: col,
		value:  val,
		typ:    '+',
	})
	return stmt
}

// Decrease 给钱减少值
func (stmt *UpdateStmt) Decrease(col string, val interface{}) *UpdateStmt {
	stmt.values = append(stmt.values, &updateSet{
		column: col,
		value:  val,
		typ:    '-',
	})
	return stmt
}

// WhereStmt 实现 WhereStmter 接口
func (stmt *UpdateStmt) WhereStmt() *WhereStmt {
	return stmt.where
}

// Where 指定 where 语句
func (stmt *UpdateStmt) Where(cond string, args ...interface{}) *UpdateStmt {
	return stmt.And(cond, args...)
}

// And 指定 where ... AND ... 语句
func (stmt *UpdateStmt) And(cond string, args ...interface{}) *UpdateStmt {
	stmt.where.And(cond, args...)
	return stmt
}

// Or 指定 where ... OR ... 语句
func (stmt *UpdateStmt) Or(cond string, args ...interface{}) *UpdateStmt {
	stmt.where.Or(cond, args...)
	return stmt
}

// Reset 重置语句
func (stmt *UpdateStmt) Reset() {
	stmt.table = ""
	stmt.where.Reset()
	stmt.values = stmt.values[:0]
}

// SQL 获取 SQL 语句以及对应的参数
func (stmt *UpdateStmt) SQL() (string, []interface{}, error) {
	if err := stmt.checkErrors(); err != nil {
		return "", nil, err
	}

	buf := New("UPDATE ")
	buf.WriteString(stmt.table)
	buf.WriteString(" SET ")

	args := make([]interface{}, 0, len(stmt.values))

	for _, val := range stmt.values {
		buf.WriteString(val.column)
		buf.WriteByte('=')

		if val.typ != 0 {
			buf.WriteString(val.column)
			buf.WriteByte(val.typ)
		}

		if named, ok := val.value.(sql.NamedArg); ok && named.Name != "" {
			buf.WriteByte('@')
			buf.WriteString(named.Name)
		} else {
			buf.WriteByte('?')
		}
		buf.WriteByte(',')
		args = append(args, val.value)
	}
	buf.TruncateLast(1)

	wq, wa, err := stmt.where.SQL()
	if err != nil {
		return "", nil, err
	}

	buf.WriteString(wq)
	args = append(args, wa...)
	return buf.String(), args, nil
}

// 检测列名是否存在重复，先排序，再与后一元素比较。
func (stmt *UpdateStmt) checkErrors() error {
	if stmt.table == "" {
		return ErrTableIsEmpty
	}

	if len(stmt.values) == 0 {
		return ErrValueIsEmpty
	}

	if stmt.columnsHasDup() {
		return ErrDupColumn
	}

	return nil
}

// 检测列名是否存在重复，先排序，再与后一元素比较。
func (stmt *UpdateStmt) columnsHasDup() bool {
	sort.SliceStable(stmt.values, func(i, j int) bool {
		return stmt.values[i].column < stmt.values[j].column
	})

	for index, col := range stmt.values {
		if index+1 >= len(stmt.values) {
			return false
		}

		if col.column == stmt.values[index+1].column {
			return true
		}
	}

	return false
}

// Exec 执行 SQL 语句
func (stmt *UpdateStmt) Exec() (sql.Result, error) {
	return exec(stmt.engine, stmt)
}

// ExecContext 执行 SQL 语句
func (stmt *UpdateStmt) ExecContext(ctx context.Context) (sql.Result, error) {
	return execContext(ctx, stmt.engine, stmt)
}

// Prepare 预编译
func (stmt *UpdateStmt) Prepare() (*sql.Stmt, error) {
	return prepare(stmt.engine, stmt)
}

// PrepareContext 预编译
func (stmt *UpdateStmt) PrepareContext(ctx context.Context) (*sql.Stmt, error) {
	return prepareContext(ctx, stmt.engine, stmt)
}
