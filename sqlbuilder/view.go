// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import "github.com/issue9/orm/v3/core"

// CreateViewStmt 创建视图的语句
type CreateViewStmt struct {
	*ddlStmt

	SelectQuery string
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
func CreateView(e core.Engine) *CreateViewStmt {
	stmt := &CreateViewStmt{}
	stmt.ddlStmt = newDDLStmt(e, stmt)

	return stmt
}

// View 将当前查询语句转换为视图
func (stmt *SelectStmt) View(name string) *CreateViewStmt {
	return CreateView(stmt.Engine()).
		From(stmt).
		Name(name)
}

// Reset 重置对象
func (stmt *CreateViewStmt) Reset() *CreateViewStmt {
	stmt.baseStmt.Reset()
	stmt.ViewName = ""
	stmt.SelectQuery = ""
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

// From 指定 Select 语句
func (stmt *CreateViewStmt) From(sel *SelectStmt) *CreateViewStmt {
	if stmt.err != nil {
		return stmt
	}

	query, err := sel.CombineSQL()
	if err != nil {
		stmt.err = err
		return stmt
	}

	stmt.SelectQuery = query
	return stmt
}

// DDLSQL 返回创建视图的 SQL 语句
func (stmt *CreateViewStmt) DDLSQL() ([]string, error) {
	if stmt.err != nil {
		return nil, stmt.Err()
	}

	if stmt.ViewName == "" {
		return nil, ErrTableIsEmpty
	}

	if hook, ok := stmt.Dialect().(CreateViewStmtHooker); ok {
		return hook.CreateViewStmtHook(stmt)
	}

	builder := core.NewBuilder("CREATE ")

	if stmt.IsReplace {
		builder.WriteString(" OR REPLACE ")
	}

	if stmt.IsTemporary {
		builder.WriteString(" TEMPORARY ")
	}

	builder.WriteString(" VIEW ").QuoteKey(stmt.ViewName)

	if len(stmt.Columns) > 0 {
		builder.WriteBytes('(')
		for _, col := range stmt.Columns {
			builder.QuoteKey(col).
				WriteBytes(',')
		}
		builder.TruncateLast(1).WriteBytes(')')
	}

	query, err := builder.WriteString(" AS ").
		WriteString(stmt.SelectQuery).
		String()
	if err != nil {
		return nil, err
	}

	return []string{query}, nil
}

// DropViewStmt 删除视图
type DropViewStmt struct {
	*ddlStmt
	name string
}

// DropView 创建视图
func DropView(e core.Engine) *DropViewStmt {
	stmt := &DropViewStmt{}
	stmt.ddlStmt = newDDLStmt(e, stmt)

	return stmt
}

// Name 指定需要删除的视图名称
func (stmt *DropViewStmt) Name(name string) *DropViewStmt {
	stmt.name = name
	return stmt
}

// DDLSQL 返回删除视图的 SQL 语句
func (stmt *DropViewStmt) DDLSQL() ([]string, error) {
	if len(stmt.name) == 0 {
		return nil, ErrTableIsEmpty
	}

	query, err := core.NewBuilder("DROP VIEW IF EXISTS ").
		QuoteKey(stmt.name).
		String()
	if err != nil {
		return nil, err
	}

	return []string{query}, nil
}

// Reset 重置对象
func (stmt *DropViewStmt) Reset() *DropViewStmt {
	stmt.baseStmt.Reset()
	stmt.name = ""

	return stmt
}
