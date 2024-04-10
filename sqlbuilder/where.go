// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package sqlbuilder

import "github.com/issue9/orm/v6/core"

// WhereStmt SQL 语句的 where 部分
type WhereStmt struct {
	andGroups []*WhereStmt
	orGroups  []*WhereStmt

	builder *core.Builder
	args    []any
}

// WhereStmtOf 用于将 [WhereStmt] 的方法与其它对象组合
type WhereStmtOf[T any] struct {
	w *WhereStmt
	t T
}

// Where 生成 [Where] 语句
func (sql *SQLBuilder) Where() *WhereStmt { return Where() }

// Where 生成 [Where] 语句
func Where() *WhereStmt {
	return &WhereStmt{
		builder: core.NewBuilder(""),
		args:    make([]any, 0, 10),
	}
}

// Reset 重置内容
func (stmt *WhereStmt) Reset() {
	stmt.andGroups = stmt.andGroups[:0]
	stmt.orGroups = stmt.orGroups[:0]

	stmt.builder.Reset()
	stmt.args = stmt.args[:0]
}

// SQL 生成 SQL 语句和对应的参数返回
func (stmt *WhereStmt) SQL() (string, []any, error) {
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
		return "", nil, SyntaxError("WHERE", "列与值不匹配")
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
		stmt.builder.WBytes(' ')
		return
	}

	v := " AND "
	if !and {
		v = " OR "
	}
	stmt.builder.WString(v)
}

// and 表示当前的语句是 and 还是 or；
// cond 表示条件语句部分，比如 "id=?"，可以为空；
// args 则表示 cond 中表示的值，可以是直接的值或是 sql.NamedArg
func (stmt *WhereStmt) where(and bool, cond string, args ...any) *WhereStmt {
	if cond == "" {
		if len(args) > 0 {
			panic("列与值不匹配")
		}

		return stmt
	}

	stmt.writeAnd(and)
	stmt.builder.WString(cond)
	stmt.args = append(stmt.args, args...)

	return stmt
}

// And 添加一条 AND 语句
func (stmt *WhereStmt) And(cond string, args ...any) *WhereStmt {
	return stmt.where(true, cond, args...)
}

// Or 添加一条 OR 语句
func (stmt *WhereStmt) Or(cond string, args ...any) *WhereStmt {
	return stmt.where(false, cond, args...)
}

// AndIsNull 指定 WHERE ... AND col IS NULL
func (stmt *WhereStmt) AndIsNull(col string) *WhereStmt {
	stmt.writeAnd(true)
	stmt.builder.QuoteKey(col).WString(" IS NULL ")
	return stmt
}

// OrIsNull 指定 WHERE ... OR col IS NULL
func (stmt *WhereStmt) OrIsNull(col string) *WhereStmt {
	stmt.writeAnd(false)
	stmt.builder.QuoteKey(col).WString(" IS NULL ")
	return stmt
}

// AndIsNotNull 指定 WHERE ... AND col IS NOT NULL
func (stmt *WhereStmt) AndIsNotNull(col string) *WhereStmt {
	stmt.writeAnd(true)
	stmt.builder.QuoteKey(col).WString(" IS NOT NULL ")
	return stmt
}

// OrIsNotNull 指定 WHERE ... OR col IS NOT NULL
func (stmt *WhereStmt) OrIsNotNull(col string) *WhereStmt {
	stmt.writeAnd(false)
	stmt.builder.QuoteKey(col).WString(" IS NOT NULL ")
	return stmt
}

// AndBetween 指定 WHERE ... AND col BETWEEN v1 AND v2
func (stmt *WhereStmt) AndBetween(col string, v1, v2 any) *WhereStmt {
	stmt.writeAnd(true)
	stmt.builder.QuoteKey(col).WString(" BETWEEN ? AND ? ")
	stmt.args = append(stmt.args, v1, v2)
	return stmt
}

// OrBetween 指定 WHERE ... OR col BETWEEN v1 AND v2
func (stmt *WhereStmt) OrBetween(col string, v1, v2 any) *WhereStmt {
	stmt.writeAnd(false)
	stmt.builder.QuoteKey(col).WString(" BETWEEN ? AND ? ")
	stmt.args = append(stmt.args, v1, v2)
	return stmt
}

// AndNotBetween 指定 WHERE ... AND col NOT BETWEEN v1 AND v2
func (stmt *WhereStmt) AndNotBetween(col string, v1, v2 any) *WhereStmt {
	stmt.writeAnd(true)
	stmt.builder.QuoteKey(col).WString(" NOT BETWEEN ? AND ? ")
	stmt.args = append(stmt.args, v1, v2)
	return stmt
}

// OrNotBetween 指定 WHERE ... OR col BETWEEN v1 AND v2
func (stmt *WhereStmt) OrNotBetween(col string, v1, v2 any) *WhereStmt {
	stmt.writeAnd(false)
	stmt.builder.QuoteKey(col).WString(" NOT BETWEEN ? AND ? ")
	stmt.args = append(stmt.args, v1, v2)
	return stmt
}

// AndLike 指定 WHERE ... AND col LIKE content
func (stmt *WhereStmt) AndLike(col string, content any) *WhereStmt {
	stmt.writeAnd(true)
	stmt.builder.QuoteKey(col).WString(" LIKE ?")
	stmt.args = append(stmt.args, content)
	return stmt
}

// OrLike 指定 WHERE ... OR col LIKE content
func (stmt *WhereStmt) OrLike(col string, content any) *WhereStmt {
	stmt.writeAnd(false)
	stmt.builder.QuoteKey(col).WString(" LIKE ?")
	stmt.args = append(stmt.args, content)
	return stmt
}

// AndNotLike 指定 WHERE ... AND col NOT LIKE content
func (stmt *WhereStmt) AndNotLike(col string, content any) *WhereStmt {
	stmt.writeAnd(true)
	stmt.builder.QuoteKey(col).WString(" NOT LIKE ?")
	stmt.args = append(stmt.args, content)
	return stmt
}

// OrNotLike 指定 WHERE ... OR col NOT LIKE content
func (stmt *WhereStmt) OrNotLike(col string, content any) *WhereStmt {
	stmt.writeAnd(false)
	stmt.builder.QuoteKey(col).WString(" NOT LIKE ?")
	stmt.args = append(stmt.args, content)
	return stmt
}

// AndIn 指定 WHERE ... AND col IN(v...)
func (stmt *WhereStmt) AndIn(col string, v ...any) *WhereStmt {
	return stmt.in(true, false, col, v...)
}

// OrIn 指定 WHERE ... OR col IN(v...)
func (stmt *WhereStmt) OrIn(col string, v ...any) *WhereStmt {
	return stmt.in(false, false, col, v...)
}

// AndNotIn 指定 WHERE ... AND col NOT IN(v...)
func (stmt *WhereStmt) AndNotIn(col string, v ...any) *WhereStmt {
	return stmt.in(true, true, col, v...)
}

// OrNotIn 指定 WHERE ... OR col IN(v...)
func (stmt *WhereStmt) OrNotIn(col string, v ...any) *WhereStmt {
	return stmt.in(false, true, col, v...)
}

func (stmt *WhereStmt) in(and, not bool, col string, v ...any) *WhereStmt {
	if len(v) == 0 {
		return stmt
	}

	stmt.writeAnd(and)
	stmt.builder.QuoteKey(col)

	if not {
		stmt.builder.WString(" NOT")
	}

	stmt.builder.WString(" IN(")
	for range v {
		stmt.builder.WBytes('?', ',')
	}
	stmt.builder.TruncateLast(1).WBytes(')')

	stmt.args = append(stmt.args, v...)

	return stmt
}

// AndGroup 开始一个子条件语句
func (stmt *WhereStmt) AndGroup(f func(*WhereStmt)) *WhereStmt {
	w := Where()
	f(w)
	stmt.appendGroup(true, w)
	return stmt
}

// OrGroup 开始一个子条件语句
func (stmt *WhereStmt) OrGroup(f func(*WhereStmt)) *WhereStmt {
	w := Where()
	f(w)
	stmt.appendGroup(false, w)
	return stmt
}

func (stmt *WhereStmt) appendGroup(and bool, w *WhereStmt) {
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

// Cond 在 expr 为真时才执行 f 中的内容
//
// expr 为条件表达式，此为 true 时，才会执行 f 函数；
// f 的原型为 `func(stmt *WhereStmt)` 其中的参数 stmt 即为当前对象的实例；
//
//	sql := Where()
//	sql.Cond(uid > 0, func(sql *WhereStmt) {
//	    sql.And("uid>?", uid)
//	});
//
// 相当于：
//
//	sql := Where()
//	if uid > 0 {
//	    sql.And("uid>?", uid)
//	}
func (stmt *WhereStmt) Cond(expr bool, f func(stmt *WhereStmt)) *WhereStmt {
	if expr {
		f(stmt)
	}
	return stmt
}

func NewWhereStmtOf[T any](t T) *WhereStmtOf[T] {
	return &WhereStmtOf[T]{w: Where(), t: t}
}

func (stmt *WhereStmtOf[T]) Where(cond string, args ...any) T {
	return stmt.And(cond, args...)
}

func (stmt *WhereStmtOf[T]) And(cond string, args ...any) T {
	stmt.w.And(cond, args...)
	return stmt.t
}

// Or 添加一条 OR 语句
func (stmt *WhereStmtOf[T]) Or(cond string, args ...any) T {
	stmt.w.Or(cond, args...)
	return stmt.t
}

// AndIsNull 指定 WHERE ... AND col IS NULL
func (stmt *WhereStmtOf[T]) AndIsNull(col string) T {
	stmt.w.AndIsNull(col)
	return stmt.t
}

// OrIsNull 指定 WHERE ... OR col IS NULL
func (stmt *WhereStmtOf[T]) OrIsNull(col string) T {
	stmt.w.OrIsNull(col)
	return stmt.t
}

// AndIsNotNull 指定 WHERE ... AND col IS NOT NULL
func (stmt *WhereStmtOf[T]) AndIsNotNull(col string) T {
	stmt.w.AndIsNotNull(col)
	return stmt.t
}

// OrIsNotNull 指定 WHERE ... OR col IS NOT NULL
func (stmt *WhereStmtOf[T]) OrIsNotNull(col string) T {
	stmt.w.OrIsNotNull(col)
	return stmt.t
}

// AndBetween 指定 WHERE ... AND col BETWEEN v1 AND v2
func (stmt *WhereStmtOf[T]) AndBetween(col string, v1, v2 any) T {
	stmt.w.AndBetween(col, v1, v2)
	return stmt.t
}

// OrBetween 指定 WHERE ... OR col BETWEEN v1 AND v2
func (stmt *WhereStmtOf[T]) OrBetween(col string, v1, v2 any) T {
	stmt.w.OrBetween(col, v1, v2)
	return stmt.t
}

// AndNotBetween 指定 WHERE ... AND col NOT BETWEEN v1 AND v2
func (stmt *WhereStmtOf[T]) AndNotBetween(col string, v1, v2 any) T {
	stmt.w.AndNotBetween(col, v1, v2)
	return stmt.t
}

// OrNotBetween 指定 WHERE ... OR col BETWEEN v1 AND v2
func (stmt *WhereStmtOf[T]) OrNotBetween(col string, v1, v2 any) T {
	stmt.w.OrNotBetween(col, v1, v2)
	return stmt.t
}

// AndLike 指定 WHERE ... AND col LIKE content
func (stmt *WhereStmtOf[T]) AndLike(col string, content any) T {
	stmt.w.AndLike(col, content)
	return stmt.t
}

// OrLike 指定 WHERE ... OR col LIKE content
func (stmt *WhereStmtOf[T]) OrLike(col string, content any) T {
	stmt.w.OrLike(col, content)
	return stmt.t
}

// AndNotLike 指定 WHERE ... AND col NOT LIKE content
func (stmt *WhereStmtOf[T]) AndNotLike(col string, content any) T {
	stmt.w.AndNotLike(col, content)
	return stmt.t
}

// OrNotLike 指定 WHERE ... OR col NOT LIKE content
func (stmt *WhereStmtOf[T]) OrNotLike(col string, content any) T {
	stmt.w.OrNotLike(col, content)
	return stmt.t
}

// AndIn 指定 WHERE ... AND col IN(v...)
func (stmt *WhereStmtOf[T]) AndIn(col string, v ...any) T {
	stmt.w.AndIn(col, v...)
	return stmt.t
}

// OrIn 指定 WHERE ... OR col IN(v...)
func (stmt *WhereStmtOf[T]) OrIn(col string, v ...any) T {
	stmt.w.OrIn(col, v...)
	return stmt.t
}

// AndNotIn 指定 WHERE ... AND col NOT IN(v...)
func (stmt *WhereStmtOf[T]) AndNotIn(col string, v ...any) T {
	stmt.w.AndNotIn(col, v...)
	return stmt.t
}

// OrNotIn 指定 WHERE ... OR col IN(v...)
func (stmt *WhereStmtOf[T]) OrNotIn(col string, v ...any) T {
	stmt.w.OrNotIn(col, v...)
	return stmt.t
}

// AndGroup 开始一个子条件语句
func (stmt *WhereStmtOf[T]) AndGroup(f func(*WhereStmt)) T {
	stmt.w.AndGroup(f)
	return stmt.t
}

// OrGroup 开始一个子条件语句
func (stmt *WhereStmtOf[T]) OrGroup(f func(*WhereStmt)) T {
	stmt.w.OrGroup(f)
	return stmt.t
}

// Cond 在 expr 为真时才执行 f 中的内容
func (stmt *WhereStmtOf[T]) Cond(expr bool, f func(stmt *WhereStmt)) T {
	stmt.w.Cond(expr, f)
	return stmt.t
}

func (stmt *WhereStmtOf[T]) WhereStmt() *WhereStmt { return stmt.w }
