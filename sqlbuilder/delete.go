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
	if stmt.table == "" {
		return "", nil, ErrTableIsEmpty
	}

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
