// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

// DeleteStmt 表示删除操作的 SQL 语句
type DeleteStmt struct {
	table string
	where *where
}

// Delete 声明一条删除语句
func Delete(table string) *DeleteStmt {
	return &DeleteStmt{
		table: table,
		where: newWhere(),
	}
}

// SQL 获取 SQL 语句，以及其参数对应的具体值
func (stmt *DeleteStmt) SQL() (string, []interface{}, error) {
	query, args, err := stmt.where.SQL()
	if err != nil {
		return "", nil, err
	}

	return "DELETE FROM " + stmt.table + query, args, nil
}

// Reset 重置语句
func (stmt *DeleteStmt) Reset() {
	stmt.table = ""
	stmt.where.Reset()
}

// Where 指定 where 语句
func (stmt *DeleteStmt) Where(and bool, cond string, args ...interface{}) *DeleteStmt {
	stmt.where.where(and, cond, args...)
	return stmt
}

// And 指定 where ... AND ... 语句
func (stmt *DeleteStmt) And(cond string, args ...interface{}) *DeleteStmt {
	stmt.where.and(cond, args...)
	return stmt
}

// Or 指定 where ... OR ... 语句
func (stmt *DeleteStmt) Or(cond string, args ...interface{}) *DeleteStmt {
	stmt.where.or(cond, args...)
	return stmt
}

// Between 指定 where ... BETWEEN k1 AND k2  语句
func (stmt *DeleteStmt) Between(and bool, col string, arg1, arg2 interface{}) *DeleteStmt {
	stmt.where.between(and, false, col, arg1, arg2)
	return stmt
}

// In 指定 where ... IN (...)  语句
func (stmt *DeleteStmt) In(and bool, col string, args ...interface{}) *DeleteStmt {
	stmt.where.in(and, false, col, args...)
	return stmt
}

// NotBetween 指定 where ... NOT BETWEEN k1 AND k2  语句
func (stmt *DeleteStmt) NotBetween(and bool, col string, arg1, arg2 interface{}) *DeleteStmt {
	stmt.where.between(and, true, col, arg1, arg2)
	return stmt
}

// NotIn 指定 where ... NOT IN (...)  语句
func (stmt *DeleteStmt) NotIn(and bool, col string, args ...interface{}) *DeleteStmt {
	stmt.where.in(and, true, col, args...)
	return stmt
}

// IsNull 指定 where ... IS NULL  语句
func (stmt *DeleteStmt) IsNull(and bool, col string) *DeleteStmt {
	stmt.where.null(and, false, col)
	return stmt
}

// IsNotNull 指定 where ... IS NOT NULL  语句
func (stmt *DeleteStmt) IsNotNull(and bool, col string) *DeleteStmt {
	stmt.where.null(and, true, col)
	return stmt
}
