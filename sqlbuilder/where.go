// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import "github.com/issue9/orm/v2/core"

// WhereStmter 带 Where 语句的 SQL
type WhereStmter interface {
	WhereStmt() *WhereStmt
}

// WhereStmt SQL 语句的 where 部分
type WhereStmt struct {
	parent    *WhereStmt
	andGroups []*WhereStmt
	orGroups  []*WhereStmt

	builder *core.Builder
	args    []interface{}
}

func newWhere() *WhereStmt {
	return &WhereStmt{
		builder: core.NewBuilder(""),
		args:    make([]interface{}, 0, 10),
	}
}

// Reset 重置内容
func (stmt *WhereStmt) Reset() {
	stmt.parent = nil
	stmt.andGroups = stmt.andGroups[:0]
	stmt.orGroups = stmt.orGroups[:0]

	stmt.builder.Reset()
	stmt.args = stmt.args[:0]
}

// SQL 生成 SQL 语句和对应的参数返回
func (stmt *WhereStmt) SQL() (string, []interface{}, error) {
	cnt := 0
	bs, err := stmt.builder.Bytes()
	if err != nil {
		return "", nil, err
	}
	for _, c := range bs {
		if c == '?' || c == '@' {
			cnt++
		}
	}

	if cnt != len(stmt.args) {
		return "", nil, ErrArgsNotMatch
	}

	for _, w := range stmt.andGroups {
		if err := stmt.buildGroup(true, w); err != nil {
			return "", nil, err
		}
	}

	for _, w := range stmt.orGroups {
		if err := stmt.buildGroup(false, w); err != nil {
			return "", nil, err
		}
	}

	query, err := stmt.builder.String()
	if err != nil {
		return "", nil, err
	}
	return query, stmt.args, nil
}

func (stmt *WhereStmt) buildGroup(and bool, g *WhereStmt) error {
	query, args, err := g.SQL()
	if err != nil {
		return err
	}

	stmt.writeAnd(and)
	stmt.builder.Quote(query, '(', ')')
	stmt.args = append(stmt.args, args...)

	return nil
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
	stmt.writeAnd(true)
	stmt.builder.QuoteKey(col).WriteString(" IS NULL ")
	return stmt
}

// OrIsNull 指定 WHERE ... OR col IS NULL
func (stmt *WhereStmt) OrIsNull(col string) *WhereStmt {
	stmt.writeAnd(false)
	stmt.builder.QuoteKey(col).WriteString(" IS NULL ")
	return stmt
}

// AndIsNotNull 指定 WHERE ... AND col IS NOT NULL
func (stmt *WhereStmt) AndIsNotNull(col string) *WhereStmt {
	stmt.writeAnd(true)
	stmt.builder.QuoteKey(col).WriteString(" IS NOT NULL ")
	return stmt
}

// OrIsNotNull 指定 WHERE ... OR col IS NOT NULL
func (stmt *WhereStmt) OrIsNotNull(col string) *WhereStmt {
	stmt.writeAnd(false)
	stmt.builder.QuoteKey(col).WriteString(" IS NOT NULL ")
	return stmt
}

// AndBetween 指定 WHERE ... AND col BETWEEN v1 AND v2
func (stmt *WhereStmt) AndBetween(col string, v1, v2 interface{}) *WhereStmt {
	stmt.writeAnd(true)
	stmt.builder.QuoteKey(col).WriteString(" BETWEEN ? AND ? ")
	stmt.args = append(stmt.args, v1, v2)
	return stmt
}

// OrBetween 指定 WHERE ... OR col BETWEEN v1 AND v2
func (stmt *WhereStmt) OrBetween(col string, v1, v2 interface{}) *WhereStmt {
	stmt.writeAnd(false)
	stmt.builder.QuoteKey(col).WriteString(" BETWEEN ? AND ? ")
	stmt.args = append(stmt.args, v1, v2)
	return stmt
}

// AndNotBetween 指定 WHERE ... AND col NOT BETWEEN v1 AND v2
func (stmt *WhereStmt) AndNotBetween(col string, v1, v2 interface{}) *WhereStmt {
	stmt.writeAnd(true)
	stmt.builder.QuoteKey(col).WriteString(" NOT BETWEEN ? AND ? ")
	stmt.args = append(stmt.args, v1, v2)
	return stmt
}

// OrNotBetween 指定 WHERE ... OR col BETWEEN v1 AND v2
func (stmt *WhereStmt) OrNotBetween(col string, v1, v2 interface{}) *WhereStmt {
	stmt.writeAnd(false)
	stmt.builder.QuoteKey(col).WriteString(" NOT BETWEEN ? AND ? ")
	stmt.args = append(stmt.args, v1, v2)
	return stmt
}

// AndLike 指定 WHERE ... AND col LIKE content
func (stmt *WhereStmt) AndLike(col string, content interface{}) *WhereStmt {
	stmt.writeAnd(true)
	stmt.builder.QuoteKey(col).WriteString(" LIKE ?")
	stmt.args = append(stmt.args, content)
	return stmt
}

// OrLike 指定 WHERE ... OR col LIKE content
func (stmt *WhereStmt) OrLike(col string, content interface{}) *WhereStmt {
	stmt.writeAnd(false)
	stmt.builder.QuoteKey(col).WriteString(" LIKE ?")
	stmt.args = append(stmt.args, content)
	return stmt
}

// AndNotLike 指定 WHERE ... AND col NOT LIKE content
func (stmt *WhereStmt) AndNotLike(col string, content interface{}) *WhereStmt {
	stmt.writeAnd(true)
	stmt.builder.QuoteKey(col).WriteString(" NOT LIKE ?")
	stmt.args = append(stmt.args, content)
	return stmt
}

// OrNotLike 指定 WHERE ... OR col NOT LIKE content
func (stmt *WhereStmt) OrNotLike(col string, content interface{}) *WhereStmt {
	stmt.writeAnd(false)
	stmt.builder.QuoteKey(col).WriteString(" NOT LIKE ?")
	stmt.args = append(stmt.args, content)
	return stmt
}

// AndIn 指定 WHERE ... AND col IN(v...)
func (stmt *WhereStmt) AndIn(col string, v ...interface{}) *WhereStmt {
	return stmt.in(true, false, col, v...)
}

// OrIn 指定 WHERE ... OR col IN(v...)
func (stmt *WhereStmt) OrIn(col string, v ...interface{}) *WhereStmt {
	return stmt.in(false, false, col, v...)
}

// AndNotIn 指定 WHERE ... AND col NOT IN(v...)
func (stmt *WhereStmt) AndNotIn(col string, v ...interface{}) *WhereStmt {
	return stmt.in(true, true, col, v...)
}

// OrNotIn 指定 WHERE ... OR col IN(v...)
func (stmt *WhereStmt) OrNotIn(col string, v ...interface{}) *WhereStmt {
	return stmt.in(false, true, col, v...)
}

func (stmt *WhereStmt) in(and, not bool, col string, v ...interface{}) *WhereStmt {
	if len(v) == 0 {
		return stmt
	}

	stmt.writeAnd(and)
	stmt.builder.QuoteKey(col)

	if not {
		stmt.builder.WriteString(" NOT")
	}

	stmt.builder.WriteString(" IN(")
	for range v {
		stmt.builder.WriteBytes('?', ',')
	}
	stmt.builder.TruncateLast(1)
	stmt.builder.WriteBytes(')')

	stmt.args = append(stmt.args, v...)

	return stmt
}

// AndGroup 开始一个子条件语句
func (stmt *WhereStmt) AndGroup() *WhereStmt {
	w := newWhere()
	stmt.appendGroup(true, w)

	return w
}

// OrGroup 开始一个子条件语句
func (stmt *WhereStmt) OrGroup() *WhereStmt {
	w := newWhere()
	stmt.appendGroup(false, w)

	return w
}

func (stmt *WhereStmt) appendGroup(and bool, w *WhereStmt) {
	w.parent = stmt

	if and {
		if stmt.andGroups == nil {
			stmt.andGroups = []*WhereStmt{w}
		} else {
			stmt.andGroups = append(stmt.andGroups, w)
		}
	} else {
		if stmt.orGroups == nil {
			stmt.orGroups = []*WhereStmt{w}
		} else {
			stmt.orGroups = append(stmt.orGroups, w)
		}
	}
}

// EndGroup 结束当前组的条件输出，返回上一层。如果没有上一层，则 panic
func (stmt *WhereStmt) EndGroup() (parent *WhereStmt) {
	if stmt.parent == nil {
		panic("没有更高层的查询条件了")
	}

	return stmt.parent
}
