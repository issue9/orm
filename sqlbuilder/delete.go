// SPDX-License-Identifier: MIT

package sqlbuilder

import "github.com/issue9/orm/v5/core"

// DeleteStmt 表示删除操作的 SQL 语句
type DeleteStmt struct {
	*execStmt
	*deleteWhere

	table string
}

type deleteWhere = WhereStmtOf[DeleteStmt]

// Delete 生成删除语句
func (sql *SQLBuilder) Delete() *DeleteStmt { return Delete(sql.engine) }

// Delete 声明一条删除语句
func Delete(e core.Engine) *DeleteStmt {
	stmt := &DeleteStmt{}
	stmt.execStmt = newExecStmt(e, stmt)
	stmt.deleteWhere = NewWhereStmtOf(stmt)

	return stmt
}

// Table 指定表名
func (stmt *DeleteStmt) Table(table string) *DeleteStmt {
	stmt.table = table
	return stmt
}

// SQL 获取 SQL 语句，以及其参数对应的具体值
func (stmt *DeleteStmt) SQL() (string, []any, error) {
	if stmt.err != nil {
		return "", nil, stmt.Err()
	}

	if stmt.table == "" {
		return "", nil, ErrTableIsEmpty
	}

	query, args, err := stmt.WhereStmt().SQL()
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
	stmt.WhereStmt().Reset()
	return stmt
}

// Delete 删除指定条件的内容
func (stmt *WhereStmt) Delete(e core.Engine) *DeleteStmt {
	del := Delete(e)
	del.deleteWhere.w = stmt
	return del
}
