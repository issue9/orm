// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"database/sql"
	"sort"
)

// UpdateStmt 更新语句
type UpdateStmt struct {
	*execStmt

	table  string
	where  *WhereStmt
	values []*updateSet

	occColumn string      // 乐观锁的列名
	occValue  interface{} // 乐观锁的当前值
}

// 表示一条 SET 语句。比如 set key=val
type updateSet struct {
	column string
	value  interface{}
	typ    byte // 类型，可以是 + 自增类型，- 自减类型，或是空值表示正常表达式
}

// Update 声明一条 UPDATE 的 SQL 语句
func Update(e Engine, d Dialect) *UpdateStmt {
	stmt := &UpdateStmt{
		where:  Where(),
		values: []*updateSet{},
	}
	stmt.execStmt = newExecStmt(e, d, stmt)

	return stmt
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

// OCC 指定一个用于乐观锁的字段。
//
// val 表示乐观锁原始的值，更新时如果值不等于 val，将更新失败。
func (stmt *UpdateStmt) OCC(col string, val interface{}) *UpdateStmt {
	stmt.occColumn = col
	stmt.occValue = val
	stmt.Increase(col, 1)
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

	stmt.occColumn = ""
	stmt.occValue = nil
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
		buf.WriteBytes(stmt.l).WriteString(val.column).WriteBytes(stmt.r)
		buf.WriteBytes('=')

		if val.typ != 0 {
			buf.WriteBytes(stmt.l).WriteString(val.column).WriteBytes(stmt.r)
			buf.WriteBytes(val.typ)
		}

		if named, ok := val.value.(sql.NamedArg); ok && named.Name != "" {
			buf.WriteBytes('@')
			buf.WriteString(named.Name)
		} else {
			buf.WriteBytes('?')
		}
		buf.WriteBytes(',')
		args = append(args, val.value)
	}
	buf.TruncateLast(1)

	wq, wa, err := stmt.getWhereSQL()
	if err != nil {
		return "", nil, err
	}

	if wq != "" {
		buf.WriteString(" WHERE ")
		buf.WriteString(wq)
		args = append(args, wa...)
	}

	return buf.String(), args, nil
}

func (stmt *UpdateStmt) getWhereSQL() (string, []interface{}, error) {
	if stmt.occColumn == "" {
		return stmt.where.SQL()
	}

	occColumn := string(stmt.l) + stmt.occColumn + string(stmt.r)
	occ := Where()
	if named, ok := stmt.occValue.(sql.NamedArg); ok && named.Name != "" {
		occ.And(occColumn+"=@"+named.Name, stmt.occValue)
	} else {
		occ.And(occColumn+"=?", stmt.occValue)
	}

	return Where().AndWhere(stmt.where).AndWhere(occ).SQL()
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
