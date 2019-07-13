// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import "strings"

// WhereStmt SQL 语句的 where 部分
type WhereStmt struct {
	builder *SQLBuilder
	args    []interface{}
}

// Where 生成一条 Where 语句
func Where() *WhereStmt {
	return &WhereStmt{
		builder: New(""),
		args:    make([]interface{}, 0, 10),
	}
}

// Reset 重置内容
func (stmt *WhereStmt) Reset() {
	stmt.builder.Reset()
	stmt.args = stmt.args[:0]
}

// SQL 生成 SQL 语句和对应的参数返回
func (stmt *WhereStmt) SQL() (string, []interface{}, error) {
	cnt := 0
	for _, c := range stmt.builder.Bytes() {
		if c == '?' || c == '@' {
			cnt++
		}
	}

	if cnt != len(stmt.args) {
		return "", nil, ErrArgsNotMatch
	}

	return stmt.builder.String(), stmt.args, nil
}

func (stmt *WhereStmt) writeAnd(and bool) {
	if stmt.builder.Len() == 0 {
		stmt.builder.WriteBytes(' ')
		return
	}

	v := " AND "
	if !and {
		v = " OR "
	}
	stmt.builder.WriteString(v)
}

// and 表示当前的语句是 and 还是 or；
// cond 表示条件语句部分，比如 "id=?"
// args 则表示 cond 中表示的值，可以是直接的值或是 sql.NamedArg
func (stmt *WhereStmt) where(and bool, cond string, args ...interface{}) *WhereStmt {
	stmt.writeAnd(and)
	stmt.builder.WriteString(cond)
	stmt.args = append(stmt.args, args...)

	return stmt
}

// And 添加一条 and 语句
func (stmt *WhereStmt) And(cond string, args ...interface{}) *WhereStmt {
	return stmt.where(true, cond, args...)
}

// Or 添加一条 OR 语句
func (stmt *WhereStmt) Or(cond string, args ...interface{}) *WhereStmt {
	return stmt.where(false, cond, args...)
}

// AndIsNull 指定 WHERE ... AND col IS NULL
func (stmt *WhereStmt) AndIsNull(col string) *WhereStmt {
	stmt.And(col + " IS NULL")
	return stmt
}

// OrIsNull 指定 WHERE ... OR col IS NULL
func (stmt *WhereStmt) OrIsNull(col string) *WhereStmt {
	stmt.Or(col + " IS NULL")
	return stmt
}

// AndIsNotNull 指定 WHERE ... AND col IS NOT NULL
func (stmt *WhereStmt) AndIsNotNull(col string) *WhereStmt {
	stmt.And(col + " IS NOT NULL")
	return stmt
}

// OrIsNotNull 指定 WHERE ... OR col IS NOT NULL
func (stmt *WhereStmt) OrIsNotNull(col string) *WhereStmt {
	stmt.Or(col + " IS NOT NULL")
	return stmt
}

// AndBetween 指定 WHERE ... AND col BETWEEN v1 AND v2
func (stmt *WhereStmt) AndBetween(col string, v1, v2 interface{}) *WhereStmt {
	stmt.And(col+" BETWEEN (? and ?)", v1, v2)
	return stmt
}

// OrBetween 指定 WHERE ... OR col BETWEEN v1 AND v2
func (stmt *WhereStmt) OrBetween(col string, v1, v2 interface{}) *WhereStmt {
	stmt.Or(col+" BETWEEN (? and ?)", v1, v2)
	return stmt
}

func (stmt *WhereStmt) AndLike(col, content string) *WhereStmt {
	stmt.And(col + " LIKE '" + content + "'")
	return stmt
}

func (stmt *WhereStmt) addWhere(and bool, w *WhereStmt) *WhereStmt {
	cond := w.builder.String()
	if strings.TrimSpace(cond) == "" {
		return stmt
	}

	stmt.writeAnd(and)
	stmt.builder.WriteBytes('(')

	stmt.builder.WriteString(cond)
	stmt.args = append(stmt.args, w.args...)

	stmt.builder.WriteBytes(')')

	return stmt
}

// AndWhere 开始一个子条件语句
func (stmt *WhereStmt) AndWhere(w *WhereStmt) *WhereStmt {
	return stmt.addWhere(true, w)
}

// OrWhere 开始一个子条件语句
func (stmt *WhereStmt) OrWhere(w *WhereStmt) *WhereStmt {
	return stmt.addWhere(false, w)
}
