// SPDX-License-Identifier: MIT

package sqlbuilder

import "github.com/issue9/orm/v4/core"

// DeleteStmt 表示删除操作的 SQL 语句
type DeleteStmt struct {
	*execStmt
	table string
	where *WhereStmt
}

// Delete 生成删除语句
func (sql *SQLBuilder) Delete() *DeleteStmt { return Delete(sql.engine) }

// Delete 声明一条删除语句
func Delete(e core.Engine) *DeleteStmt {
	stmt := &DeleteStmt{}
	stmt.execStmt = newExecStmt(e, stmt)
	stmt.where = Where()

	return stmt
}

// Table 指定表名
func (stmt *DeleteStmt) Table(table string) *DeleteStmt {
	stmt.table = table
	return stmt
}

// SQL 获取 SQL 语句，以及其参数对应的具体值
func (stmt *DeleteStmt) SQL() (string, []interface{}, error) {
	if stmt.err != nil {
		return "", nil, stmt.Err()
	}

	if stmt.table == "" {
		return "", nil, ErrTableIsEmpty
	}

	query, args, err := stmt.where.SQL()
	if err != nil {
		return "", nil, err
	}

	q, err := core.NewBuilder("DELETE FROM ").
		QuoteKey(stmt.table).
		WString(" WHERE ").
		WString(query).
		String()
	if err != nil {
		return "", nil, err
	}
	return q, args, nil
}

// Reset 重置语句
func (stmt *DeleteStmt) Reset() *DeleteStmt {
	stmt.baseStmt.Reset()
	stmt.table = ""
	stmt.where.Reset()
	return stmt
}

// WhereStmt 实现 WhereStmter 接口
func (stmt *DeleteStmt) WhereStmt() *WhereStmt { return stmt.where }

// Where DeleteStmt.And 的别名
func (stmt *DeleteStmt) Where(cond string, args ...interface{}) *DeleteStmt {
	return stmt.And(cond, args...)
}

// And 添加一条 and 语句
func (stmt *DeleteStmt) And(cond string, args ...interface{}) *DeleteStmt {
	stmt.where.And(cond, args...)
	return stmt
}

// Or 添加一条 OR 语句
func (stmt *DeleteStmt) Or(cond string, args ...interface{}) *DeleteStmt {
	stmt.where.Or(cond, args...)
	return stmt
}

// AndIsNull 指定 WHERE ... AND col IS NULL
func (stmt *DeleteStmt) AndIsNull(col string) *DeleteStmt {
	stmt.where.AndIsNull(col)
	return stmt
}

// OrIsNull 指定 WHERE ... OR col IS NULL
func (stmt *DeleteStmt) OrIsNull(col string) *DeleteStmt {
	stmt.where.OrIsNull(col)
	return stmt
}

// AndIsNotNull 指定 WHERE ... AND col IS NOT NULL
func (stmt *DeleteStmt) AndIsNotNull(col string) *DeleteStmt {
	stmt.where.AndIsNotNull(col)
	return stmt
}

// OrIsNotNull 指定 WHERE ... OR col IS NOT NULL
func (stmt *DeleteStmt) OrIsNotNull(col string) *DeleteStmt {
	stmt.where.OrIsNotNull(col)
	return stmt
}

// AndBetween 指定 WHERE ... AND col BETWEEN v1 AND v2
func (stmt *DeleteStmt) AndBetween(col string, v1, v2 interface{}) *DeleteStmt {
	stmt.where.AndBetween(col, v1, v2)
	return stmt
}

// OrBetween 指定 WHERE ... OR col BETWEEN v1 AND v2
func (stmt *DeleteStmt) OrBetween(col string, v1, v2 interface{}) *DeleteStmt {
	stmt.where.OrBetween(col, v1, v2)
	return stmt
}

// AndNotBetween 指定 WHERE ... AND col NOT BETWEEN v1 AND v2
func (stmt *DeleteStmt) AndNotBetween(col string, v1, v2 interface{}) *DeleteStmt {
	stmt.where.AndNotBetween(col, v1, v2)
	return stmt
}

// OrNotBetween 指定 WHERE ... OR col BETWEEN v1 AND v2
func (stmt *DeleteStmt) OrNotBetween(col string, v1, v2 interface{}) *DeleteStmt {
	stmt.where.OrNotBetween(col, v1, v2)
	return stmt
}

// AndLike 指定 WHERE ... AND col LIKE content
func (stmt *DeleteStmt) AndLike(col string, content interface{}) *DeleteStmt {
	stmt.where.AndLike(col, content)
	return stmt
}

// OrLike 指定 WHERE ... OR col LIKE content
func (stmt *DeleteStmt) OrLike(col string, content interface{}) *DeleteStmt {
	stmt.where.OrLike(col, content)
	return stmt
}

// AndNotLike 指定 WHERE ... AND col NOT LIKE content
func (stmt *DeleteStmt) AndNotLike(col string, content interface{}) *DeleteStmt {
	stmt.where.AndNotLike(col, content)
	return stmt
}

// OrNotLike 指定 WHERE ... OR col NOT LIKE content
func (stmt *DeleteStmt) OrNotLike(col string, content interface{}) *DeleteStmt {
	stmt.where.OrNotLike(col, content)
	return stmt
}

// AndIn 指定 WHERE ... AND col IN(v...)
func (stmt *DeleteStmt) AndIn(col string, v ...interface{}) *DeleteStmt {
	stmt.where.AndIn(col, v...)
	return stmt
}

// OrIn 指定 WHERE ... OR col IN(v...)
func (stmt *DeleteStmt) OrIn(col string, v ...interface{}) *DeleteStmt {
	stmt.where.OrIn(col, v...)
	return stmt
}

// AndNotIn 指定 WHERE ... AND col NOT IN(v...)
func (stmt *DeleteStmt) AndNotIn(col string, v ...interface{}) *DeleteStmt {
	stmt.where.AndNotIn(col, v...)
	return stmt
}

// OrNotIn 指定 WHERE ... OR col IN(v...)
func (stmt *DeleteStmt) OrNotIn(col string, v ...interface{}) *DeleteStmt {
	stmt.where.OrNotIn(col, v...)
	return stmt
}

// AndGroup 开始一个子条件语句
func (stmt *DeleteStmt) AndGroup() *WhereStmt {
	return stmt.where.AndGroup()
}

// OrGroup 开始一个子条件语句
func (stmt *DeleteStmt) OrGroup() *WhereStmt {
	return stmt.where.OrGroup()
}

// Delete 删除指定条件的内容
func (stmt *WhereStmt) Delete(e core.Engine) *DeleteStmt {
	del := Delete(e)
	del.where = stmt
	return del
}
