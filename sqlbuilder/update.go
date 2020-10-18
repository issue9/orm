// SPDX-License-Identifier: MIT

package sqlbuilder

import (
	"database/sql"
	"sort"

	"github.com/issue9/orm/v3/core"
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

// Update 生成更新语句
func (sql *SQLBuilder) Update() *UpdateStmt {
	return Update(sql.engine)
}

// Update 声明一条 UPDATE 的 SQL 语句
func Update(e core.Engine) *UpdateStmt {
	stmt := &UpdateStmt{values: []*updateSet{}}
	stmt.execStmt = newExecStmt(e, stmt)
	stmt.where = Where()

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

// OCC 指定一个用于乐观锁的字段
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

// Reset 重置语句
func (stmt *UpdateStmt) Reset() *UpdateStmt {
	stmt.baseStmt.Reset()

	stmt.table = ""
	stmt.where.Reset()
	stmt.values = stmt.values[:0]

	stmt.occColumn = ""
	stmt.occValue = nil

	return stmt
}

// SQL 获取 SQL 语句以及对应的参数
func (stmt *UpdateStmt) SQL() (string, []interface{}, error) {
	if stmt.err != nil {
		return "", nil, stmt.Err()
	}

	if err := stmt.checkErrors(); err != nil {
		return "", nil, err
	}

	buf := core.NewBuilder("UPDATE ")
	buf.WString(stmt.table)
	buf.WString(" SET ")

	args := make([]interface{}, 0, len(stmt.values))

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

func (stmt *UpdateStmt) getWhereSQL() (string, []interface{}, error) {
	if stmt.occColumn == "" {
		return stmt.where.SQL()
	}

	w := Where()
	w.appendGroup(true, stmt.where)

	occ := w.AndGroup()
	if named, ok := stmt.occValue.(sql.NamedArg); ok && named.Name != "" {
		occ.And(stmt.occColumn+"=@"+named.Name, stmt.occValue)
	} else {
		occ.And(stmt.occColumn+"=?", stmt.occValue)
	}

	q, a, err := w.SQL()

	return q, a, err
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

// Where UpdateStmt.And 的别名
func (stmt *UpdateStmt) Where(cond string, args ...interface{}) *UpdateStmt {
	return stmt.And(cond, args...)
}

// And 添加一条 and 语句
func (stmt *UpdateStmt) And(cond string, args ...interface{}) *UpdateStmt {
	stmt.where.And(cond, args...)
	return stmt
}

// Or 添加一条 OR 语句
func (stmt *UpdateStmt) Or(cond string, args ...interface{}) *UpdateStmt {
	stmt.where.Or(cond, args...)
	return stmt
}

// AndIsNull 指定 WHERE ... AND col IS NULL
func (stmt *UpdateStmt) AndIsNull(col string) *UpdateStmt {
	stmt.where.AndIsNull(col)
	return stmt
}

// OrIsNull 指定 WHERE ... OR col IS NULL
func (stmt *UpdateStmt) OrIsNull(col string) *UpdateStmt {
	stmt.where.OrIsNull(col)
	return stmt
}

// AndIsNotNull 指定 WHERE ... AND col IS NOT NULL
func (stmt *UpdateStmt) AndIsNotNull(col string) *UpdateStmt {
	stmt.where.AndIsNotNull(col)
	return stmt
}

// OrIsNotNull 指定 WHERE ... OR col IS NOT NULL
func (stmt *UpdateStmt) OrIsNotNull(col string) *UpdateStmt {
	stmt.where.OrIsNotNull(col)
	return stmt
}

// AndBetween 指定 WHERE ... AND col BETWEEN v1 AND v2
func (stmt *UpdateStmt) AndBetween(col string, v1, v2 interface{}) *UpdateStmt {
	stmt.where.AndBetween(col, v1, v2)
	return stmt
}

// OrBetween 指定 WHERE ... OR col BETWEEN v1 AND v2
func (stmt *UpdateStmt) OrBetween(col string, v1, v2 interface{}) *UpdateStmt {
	stmt.where.OrBetween(col, v1, v2)
	return stmt
}

// AndNotBetween 指定 WHERE ... AND col NOT BETWEEN v1 AND v2
func (stmt *UpdateStmt) AndNotBetween(col string, v1, v2 interface{}) *UpdateStmt {
	stmt.where.AndNotBetween(col, v1, v2)
	return stmt
}

// OrNotBetween 指定 WHERE ... OR col BETWEEN v1 AND v2
func (stmt *UpdateStmt) OrNotBetween(col string, v1, v2 interface{}) *UpdateStmt {
	stmt.where.OrNotBetween(col, v1, v2)
	return stmt
}

// AndLike 指定 WHERE ... AND col LIKE content
func (stmt *UpdateStmt) AndLike(col string, content interface{}) *UpdateStmt {
	stmt.where.AndLike(col, content)
	return stmt
}

// OrLike 指定 WHERE ... OR col LIKE content
func (stmt *UpdateStmt) OrLike(col string, content interface{}) *UpdateStmt {
	stmt.where.OrLike(col, content)
	return stmt
}

// AndNotLike 指定 WHERE ... AND col NOT LIKE content
func (stmt *UpdateStmt) AndNotLike(col string, content interface{}) *UpdateStmt {
	stmt.where.AndNotLike(col, content)
	return stmt
}

// OrNotLike 指定 WHERE ... OR col NOT LIKE content
func (stmt *UpdateStmt) OrNotLike(col string, content interface{}) *UpdateStmt {
	stmt.where.OrNotLike(col, content)
	return stmt
}

// AndIn 指定 WHERE ... AND col IN(v...)
func (stmt *UpdateStmt) AndIn(col string, v ...interface{}) *UpdateStmt {
	stmt.where.AndIn(col, v...)
	return stmt
}

// OrIn 指定 WHERE ... OR col IN(v...)
func (stmt *UpdateStmt) OrIn(col string, v ...interface{}) *UpdateStmt {
	stmt.where.OrIn(col, v...)
	return stmt
}

// AndNotIn 指定 WHERE ... AND col NOT IN(v...)
func (stmt *UpdateStmt) AndNotIn(col string, v ...interface{}) *UpdateStmt {
	stmt.where.AndNotIn(col, v...)
	return stmt
}

// OrNotIn 指定 WHERE ... OR col IN(v...)
func (stmt *UpdateStmt) OrNotIn(col string, v ...interface{}) *UpdateStmt {
	stmt.where.OrNotIn(col, v...)
	return stmt
}

// AndGroup 开始一个子条件语句
func (stmt *UpdateStmt) AndGroup() *WhereStmt {
	return stmt.where.AndGroup()
}

// OrGroup 开始一个子条件语句
func (stmt *UpdateStmt) OrGroup() *WhereStmt {
	return stmt.where.OrGroup()
}

// Update 更新指定条件内容
func (stmt *WhereStmt) Update(e core.Engine) *UpdateStmt {
	upd := Update(e)
	upd.where = stmt
	return upd
}
