// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

// DeleteStmt 表示删除操作的 SQL 语句
type DeleteStmt struct {
	*execStmt
	table string
	where *WhereStmt
}

// Delete 声明一条删除语句
func Delete(e Engine, d Dialect) *DeleteStmt {
	stmt := &DeleteStmt{where: Where()}
	stmt.execStmt = newExecStmt(e, d, stmt)

	return stmt
}

// Table 指定表名
func (stmt *DeleteStmt) Table(table string) *DeleteStmt {
	stmt.table = table
	return stmt
}

// SQL 获取 SQL 语句，以及其参数对应的具体值
func (stmt *DeleteStmt) SQL() (string, []interface{}, error) {
	if stmt.table == "" {
		return "", nil, ErrTableIsEmpty
	}

	query, args, err := stmt.where.SQL()
	if err != nil {
		return "", nil, err
	}

	builder := New("DELETE FROM ").
		WriteBytes(stmt.l).
		WriteString(stmt.table).
		WriteBytes(stmt.r).
		WriteString(" WHERE ").
		WriteString(query)

	return builder.String(), args, nil
}

// Reset 重置语句
func (stmt *DeleteStmt) Reset() *DeleteStmt {
	stmt.table = ""
	stmt.where.Reset()
	return stmt
}

// WhereStmt 实现 WhereStmter 接口
func (stmt *DeleteStmt) WhereStmt() *WhereStmt {
	return stmt.where
}

// Where 指定 where 语句
func (stmt *DeleteStmt) Where(cond string, args ...interface{}) *DeleteStmt {
	return stmt.And(cond, args...)
}

// And 指定 where ... AND ... 语句
func (stmt *DeleteStmt) And(cond string, args ...interface{}) *DeleteStmt {
	stmt.where.And(cond, args...)
	return stmt
}

// Or 指定 where ... OR ... 语句
func (stmt *DeleteStmt) Or(cond string, args ...interface{}) *DeleteStmt {
	stmt.where.Or(cond, args...)
	return stmt
}
