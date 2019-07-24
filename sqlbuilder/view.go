// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import "github.com/issue9/orm/v2/core"

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

	query, args, err := stmt.selectStmt.SQL()
	if err != nil {
		return "", err
	}
	return fillArgs(query, args)
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

	builder.WriteString(" AS ").WriteString(selectQuery)

	return []string{builder.String()}, nil
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

	builder := core.NewBuilder("DROP VIEW IF EXISTS ").QuoteKey(stmt.name)

	return []string{builder.String()}, nil
}

// Reset 重置对象
func (stmt *DropViewStmt) Reset() *DropViewStmt {
	stmt.name = ""

	return stmt
}
