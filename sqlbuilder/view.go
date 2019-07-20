// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

// CreateViewStmt 创建视图的语句
type CreateViewStmt struct {
	*ddlStmt

	selectStmt  *SelectStmt
	ViewName    string
	Columns     []string
	IsTemporary bool
	IsReplace   bool
}

// CreateViewStmtHooker CreateViewStmt.DDLSQL 的钩子函数
type CreateViewStmtHooker interface {
	CreateViewStmtHook(*CreateViewStmt) ([]string, error)
}

// CreateView 创建视图
func CreateView(e Engine, d Dialect) *CreateViewStmt {
	stmt := &CreateViewStmt{}
	stmt.ddlStmt = newDDLStmt(e, d, stmt)

	return stmt
}

// View 将当前查询语句转换为视图
func (stmt *SelectStmt) View(name string) *CreateViewStmt {
	return CreateView(stmt.Engine(), stmt.Dialect()).
		From(stmt)
}

// Reset 重置对象
func (stmt *CreateViewStmt) Reset() *CreateViewStmt {
	stmt.ViewName = ""
	stmt.selectStmt = nil
	stmt.Columns = stmt.Columns[:0]
	stmt.IsTemporary = false
	stmt.IsReplace = false

	return stmt
}

// Column 指定视图的列，如果未指定，则会直接采用 Select 中的列信息
func (stmt *CreateViewStmt) Column(col ...string) *CreateViewStmt {
	if stmt.Columns == nil {
		stmt.Columns = col
		return stmt
	}

	stmt.Columns = append(stmt.Columns, col...)
	return stmt
}

// Name 指定视图名称
func (stmt *CreateViewStmt) Name(name string) *CreateViewStmt {
	stmt.ViewName = name
	return stmt
}

// Temporary 临时视图
func (stmt *CreateViewStmt) Temporary() *CreateViewStmt {
	stmt.IsTemporary = true
	return stmt
}

// Replace 如果已经存在，则更新视图内容
func (stmt *CreateViewStmt) Replace() *CreateViewStmt {
	stmt.IsReplace = true
	return stmt
}

// SelectQuery 获取 SELECT 的查询内容内容
func (stmt *CreateViewStmt) SelectQuery() (string, error) {
	if stmt.selectStmt == nil {
		return "", ErrValueIsEmpty
	}

	return fillArgs(stmt.selectStmt)
}

// From 指定 Select 语句
//
// 在调用 DDLSQL 之前，可以继续修改 sel 的内容。
func (stmt *CreateViewStmt) From(sel *SelectStmt) *CreateViewStmt {
	stmt.selectStmt = sel
	return stmt
}

// DDLSQL 返回创建视图的 SQL 语句
func (stmt *CreateViewStmt) DDLSQL() ([]string, error) {
	if stmt.ViewName == "" {
		return nil, ErrTableIsEmpty
	}

	if hook, ok := stmt.Dialect().(CreateViewStmtHooker); ok {
		return hook.CreateViewStmtHook(stmt)
	}

	selectQuery, err := stmt.SelectQuery()
	if err != nil {
		return nil, err
	}

	builder := New("CREATE ")

	if stmt.IsReplace {
		builder.WriteString(" OR REPLACE ")
	}

	if stmt.IsTemporary {
		builder.WriteString(" TEMPORARY ")
	}

	builder.WriteString(" VIEW ").Quote(stmt.ViewName, stmt.l, stmt.r)

	if len(stmt.Columns) > 0 {
		builder.WriteBytes('(')
		for _, col := range stmt.Columns {
			builder.Quote(col, stmt.l, stmt.r).
				WriteBytes(',')
		}
		builder.TruncateLast(1).WriteBytes(')')
	}

	builder.WriteString(selectQuery)

	return []string{builder.String()}, nil
}
