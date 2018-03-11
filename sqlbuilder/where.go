// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

// WhereStmt SQL 语句的 where 部分
type WhereStmt struct {
	buffer     *SQLBuilder
	args       []interface{}
	groupDepth int
}

func newWhereStmt() *WhereStmt {
	return &WhereStmt{
		buffer: New(""),
		args:   make([]interface{}, 0, 10),
	}
}

// Reset 重置内容
func (stmt *WhereStmt) Reset() {
	stmt.buffer.Reset()
	stmt.args = stmt.args[:0]
	stmt.groupDepth = 0
}

// SQL 生成 SQL 语句和对应的参数返回
func (stmt *WhereStmt) SQL() (string, []interface{}, error) {
	cnt := 0
	for _, c := range stmt.buffer.Bytes() {
		if c == '?' || c == '@' {
			cnt++
		}
	}
	if cnt != len(stmt.args) {
		return "", nil, ErrArgsNotMatch
	}

	return stmt.buffer.String(), stmt.args, nil
}

func (stmt *WhereStmt) writeAnd(and bool) {
	if stmt.buffer.Len() == 0 {
		stmt.buffer.WriteString(" WHERE ")
		return
	}

	v := " AND "
	if !and {
		v = " OR "
	}
	stmt.buffer.WriteString(v)
}

// and 表示当前的语句是 and 还是 or；
// cond 表示条件语句部分，比如 "id=?"
// args 则表示 cond 中表示的值，可以是直接的值或是 sql.NamedArg
func (stmt *WhereStmt) where(and bool, cond string, args ...interface{}) *WhereStmt {
	stmt.writeAnd(and)
	stmt.buffer.WriteString(cond)
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

func (stmt *WhereStmt) groupWhere(and bool, cond string, args ...interface{}) *WhereStmt {
	stmt.writeAnd(and)
	stmt.buffer.WriteByte('(')
	stmt.buffer.WriteString(cond)
	stmt.args = append(stmt.args, args...)
	stmt.groupDepth++

	return stmt
}

// AndGroup 开始一个子条件语句
func (stmt *WhereStmt) AndGroup(cond string, args ...interface{}) *WhereStmt {
	return stmt.groupWhere(true, cond, args...)
}

// OrGroup 开始一个子条件语句
func (stmt *WhereStmt) OrGroup(cond string, args ...interface{}) *WhereStmt {
	return stmt.groupWhere(false, cond, args...)
}

// EndGroup 结束一个子条件语句
func (stmt *WhereStmt) EndGroup() *WhereStmt {
	stmt.buffer.WriteByte(')')
	stmt.groupDepth--

	if stmt.groupDepth < 0 {
		panic("() 必须结对出现")
	}

	return stmt
}
