// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package sqlbuilder

import (
	"database/sql"

	"github.com/issue9/sliceutil"

	"github.com/issue9/orm/v6/core"
)

// UpdateStmt 更新语句
type UpdateStmt struct {
	*execStmt
	*updateWhere

	table  string
	values []*updateSet

	occColumn string // 乐观锁的列名
	occValue  any    // 乐观锁的当前值
}

type updateWhere = WhereStmtOf[*UpdateStmt]

// 表示一条 SET 语句。比如 set key=val
type updateSet struct {
	column string
	value  any
	typ    byte // 类型，可以是 + 自增类型，- 自减类型，或是空值表示正常表达式
}

// Update 生成更新语句
func (sql *SQLBuilder) Update() *UpdateStmt { return Update(sql.engine) }

// Update 声明一条 UPDATE 的 SQL 语句
func Update(e core.Engine) *UpdateStmt {
	stmt := &UpdateStmt{values: []*updateSet{}}
	stmt.execStmt = newExecStmt(e, stmt)
	stmt.updateWhere = NewWhereStmtOf(stmt)

	return stmt
}

// Table 指定表名
func (stmt *UpdateStmt) Table(table string) *UpdateStmt {
	stmt.table = table
	return stmt
}

// Set 设置值，若 col 相同，则会覆盖
//
// val 可以是 sql.NamedArg 类型
func (stmt *UpdateStmt) Set(col string, val any) *UpdateStmt {
	stmt.values = append(stmt.values, &updateSet{
		column: col,
		value:  val,
		typ:    0,
	})
	return stmt
}

// Increase 给列增加值
func (stmt *UpdateStmt) Increase(col string, val any) *UpdateStmt {
	stmt.values = append(stmt.values, &updateSet{
		column: col,
		value:  val,
		typ:    '+',
	})
	return stmt
}

// Decrease 给列减少值
func (stmt *UpdateStmt) Decrease(col string, val any) *UpdateStmt {
	stmt.values = append(stmt.values, &updateSet{
		column: col,
		value:  val,
		typ:    '-',
	})
	return stmt
}

// OCC 指定一个用于乐观锁的字段
//
// val 表示乐观锁原始的值，更新时如果值不等于 val，将更新失败。
func (stmt *UpdateStmt) OCC(col string, val any) *UpdateStmt {
	stmt.occColumn = col
	stmt.occValue = val
	stmt.Increase(col, 1)
	return stmt
}

// Reset 重置语句
func (stmt *UpdateStmt) Reset() *UpdateStmt {
	stmt.baseStmt.Reset()

	stmt.table = ""
	stmt.WhereStmt().Reset()
	stmt.values = stmt.values[:0]

	stmt.occColumn = ""
	stmt.occValue = nil

	return stmt
}

// SQL 获取 SQL 语句以及对应的参数
func (stmt *UpdateStmt) SQL() (string, []any, error) {
	if stmt.err != nil {
		return "", nil, stmt.Err()
	}

	if err := stmt.checkErrors(); err != nil {
		return "", nil, err
	}

	buf := core.NewBuilder("UPDATE ").
		QuoteKey(stmt.table).
		WString(" SET ")

	args := make([]any, 0, len(stmt.values))

	for _, val := range stmt.values {
		buf.QuoteKey(val.column).WBytes('=')

		if val.typ != 0 {
			buf.QuoteKey(val.column).WBytes(val.typ)
		}

		if named, ok := val.value.(sql.NamedArg); ok && named.Name != "" {
			buf.WBytes('@').WString(named.Name)
		} else {
			buf.WBytes('?')
		}
		buf.WBytes(',')
		args = append(args, val.value)
	}
	buf.TruncateLast(1)

	wq, wa, err := stmt.getWhereSQL()
	if err != nil {
		return "", nil, err
	}

	if wq != "" {
		buf.WString(" WHERE ").WString(wq)
		args = append(args, wa...)
	}

	query, err := buf.String()
	if err != nil {
		return "", nil, err
	}
	return query, args, nil
}

func (stmt *UpdateStmt) getWhereSQL() (string, []any, error) {
	if stmt.occColumn == "" {
		return stmt.WhereStmt().SQL()
	}

	w := Where()
	w.appendGroup(true, stmt.WhereStmt())

	if named, ok := stmt.occValue.(sql.NamedArg); ok && named.Name != "" {
		w.AndGroup(func(occ *WhereStmt) {
			occ.And(stmt.occColumn+"=@"+named.Name, stmt.occValue)
		})
	} else {
		w.AndGroup(func(occ *WhereStmt) {
			occ.And(stmt.occColumn+"=?", stmt.occValue)
		})
	}

	return w.SQL()
}

// 检测列名是否存在重复，先排序，再与后一元素比较。
func (stmt *UpdateStmt) checkErrors() error {
	if stmt.table == "" {
		return SyntaxError("UPDATE", "未指定表名")
	}

	if len(stmt.values) == 0 {
		return SyntaxError("UPDATE", "未指定任何更新的值")
	}

	if len(sliceutil.Dup(stmt.values, func(i, j *updateSet) bool { return i.column == j.column })) > 0 {
		return SyntaxError("UPDATE", "存在重复的列名")
	}

	return nil
}

// Update 更新指定条件内容
func (stmt *WhereStmt) Update(e core.Engine) *UpdateStmt {
	upd := Update(e)
	upd.updateWhere.w = stmt
	return upd
}
