// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"strconv"
)

// 内置命名参数的前缀。
// 最终会生成 @___key_1 这样格式的命名参数。
const innerArgsPrefix = "@___key_"

// SQL 语句的 where 部分
type where struct {
	buffer        *stringBuilder
	args          []interface{}
	argsName      []string // 参数对应的命名参数
	innerArgIndex int      // 内置命名参数的计数器，用于生成唯一参数名称
}

func newWhere() *where {
	return &where{
		buffer: new(stringBuilder),
		args:   make([]interface{}, 0, 10),
	}
}

func (w *where) Reset() {
	w.buffer.reset()
	w.args = w.args[:0]
	w.innerArgIndex = 0
}

func (w *where) SQL() (string, []interface{}, error) {
	return w.buffer.string(), w.args, nil
}

func (w *where) writeInnerArgName() {
	w.innerArgIndex++
	w.buffer.writeString(innerArgsPrefix)
	w.buffer.writeString(strconv.Itoa(w.innerArgIndex))
}

func (w *where) writeAnd(and bool) {
	if w.buffer.len() == 0 {
		w.buffer.writeString(" WHERE ")
		return
	}

	v := " AND "
	if !and {
		v = " OR "
	}
	w.buffer.writeString(v)
}

// and 表示当前的语句是 and 还是 or；
// cond 表示条件语句部分，比如 "id=?"
// args 则表示 cond 中表示的值，可以是直接的值或是 sql.NamedArg
func (w *where) where(and bool, cond string, args ...interface{}) {
	w.writeAnd(and)

	w.buffer.writeString(cond)
	w.args = append(w.args, args...)
}

func (w *where) and(cond string, args ...interface{}) {
	w.where(true, cond, args...)
}

func (w *where) or(cond string, args ...interface{}) {
	w.where(false, cond, args...)
}

func (w *where) in(and, not bool, col string, args ...interface{}) {
	w.writeAnd(and)

	w.buffer.writeString(col)
	if not {
		w.buffer.writeString(" NOT")
	}
	w.buffer.writeString(" IN(")
	for range args {
		w.writeInnerArgName()
		w.buffer.writeByte(',')
	}
	w.buffer.truncateLast(1) // 去掉最后一 个逗号
	w.buffer.writeByte(')')
}

func (w *where) between(and, not bool, col string, arg1, arg2 interface{}) {
	w.writeAnd(and)

	w.buffer.writeString(col)
	if not {
		w.buffer.writeString(" NOT")
	}
	w.buffer.writeString(" BETWEEN ")

	w.writeInnerArgName()
	w.buffer.writeString(" AND ")
	w.writeInnerArgName()

	w.args = append(w.args, arg1, arg2)
}

func (w *where) null(and, not bool, col string) {
	w.writeAnd(and)

	w.buffer.writeString(col)
	if not {
		w.buffer.writeString(" IS NOT NULL ")
		return
	}

	w.buffer.writeString(" IS NULL ")
}
