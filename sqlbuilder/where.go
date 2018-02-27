// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import "github.com/issue9/orm/internal/stringbuilder"

// SQL 语句的 where 部分
type where struct {
	buffer *stringbuilder.StringBuilder
	args   []interface{}
}

func newWhere() *where {
	return &where{
		buffer: stringbuilder.New(""),
		args:   make([]interface{}, 0, 10),
	}
}

func (w *where) Reset() {
	w.buffer.Reset()
	w.args = w.args[:0]
}

func (w *where) SQL() (string, []interface{}, error) {
	// TODO 检测 args 中的参数与 buffer 中的占位符是否相同
	return w.buffer.String(), w.args, nil
}

func (w *where) writeAnd(and bool) {
	if w.buffer.Len() == 0 {
		w.buffer.WriteString(" WHERE ")
		return
	}

	v := " AND "
	if !and {
		v = " OR "
	}
	w.buffer.WriteString(v)
}

// and 表示当前的语句是 and 还是 or；
// cond 表示条件语句部分，比如 "id=?"
// args 则表示 cond 中表示的值，可以是直接的值或是 sql.NamedArg
func (w *where) where(and bool, cond string, args ...interface{}) {
	w.writeAnd(and)

	w.buffer.WriteString(cond)
	w.args = append(w.args, args...)
}

func (w *where) and(cond string, args ...interface{}) {
	w.where(true, cond, args...)
}

func (w *where) or(cond string, args ...interface{}) {
	w.where(false, cond, args...)
}
