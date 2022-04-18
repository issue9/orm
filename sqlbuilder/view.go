// SPDX-License-Identifier: MIT

package sqlbuilder

import "github.com/issue9/orm/v5/core"

// CreateViewStmt 创建视图的语句
type CreateViewStmt struct {
	*ddlStmt

	selectQuery string
	name        string
	columns     []string
	temporary   bool
	replace     bool
}

// CreateView 创建视图
func (sql *SQLBuilder) CreateView() *CreateViewStmt { return CreateView(sql.engine) }

// CreateView 创建视图
func CreateView(e core.Engine) *CreateViewStmt {
	stmt := &CreateViewStmt{}
	stmt.ddlStmt = newDDLStmt(e, stmt)

	return stmt
}

// View 将当前查询语句转换为视图
func (stmt *SelectStmt) View(name string) *CreateViewStmt {
	return CreateView(stmt.Engine()).From(stmt).Name(name)
}

// Reset 重置对象
func (stmt *CreateViewStmt) Reset() *CreateViewStmt {
	stmt.baseStmt.Reset()
	stmt.name = ""
	stmt.selectQuery = ""
	stmt.columns = stmt.columns[:0]
	stmt.temporary = false
	stmt.replace = false

	return stmt
}

// Column 指定视图的列，如果未指定，则会直接采用 Select 中的列信息
func (stmt *CreateViewStmt) Column(col ...string) *CreateViewStmt {
	if stmt.columns == nil {
		stmt.columns = col
		return stmt
	}

	stmt.columns = append(stmt.columns, col...)
	return stmt
}

// Name 指定视图名称
func (stmt *CreateViewStmt) Name(name string) *CreateViewStmt {
	stmt.name = name
	return stmt
}

// Temporary 临时视图
func (stmt *CreateViewStmt) Temporary() *CreateViewStmt {
	stmt.temporary = true
	return stmt
}

// Replace 如果已经存在，则更新视图内容
func (stmt *CreateViewStmt) Replace() *CreateViewStmt {
	stmt.replace = true
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

	stmt.selectQuery = query
	return stmt
}

// FromQuery 指定查询语句
//
// FromQuery 和 From 会相互覆盖。
func (stmt *CreateViewStmt) FromQuery(query string) *CreateViewStmt {
	stmt.selectQuery = query
	return stmt
}

// DDLSQL 返回创建视图的 SQL 语句
func (stmt *CreateViewStmt) DDLSQL() ([]string, error) {
	return stmt.Dialect().CreateViewSQL(stmt.replace, stmt.temporary, stmt.name, stmt.selectQuery, stmt.columns)
}

// DropViewStmt 删除视图
type DropViewStmt struct {
	*ddlStmt
	name string
}

func (sql *SQLBuilder) DropView() *DropViewStmt { return DropView(sql.engine) }

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
