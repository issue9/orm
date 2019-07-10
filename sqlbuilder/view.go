// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

// CreateViewStmt 创建视图的语句
type CreateViewStmt struct {
	*ddlStmt

	name        string
	selectStmt  *SelectStmt
	columns     []string
	checkOption string
	temporary   bool
	replace     bool
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
	stmt.name = ""
	stmt.selectStmt = nil
	stmt.columns = stmt.columns[:0]
	stmt.temporary = false
	stmt.replace = false

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

// CheckOption 指定 CHECK OPTION 选项
func (stmt *CreateViewStmt) CheckOption(opt string) *CreateViewStmt {
	stmt.checkOption = opt
	return stmt
}

// From 指定 Select 语句
func (stmt *CreateViewStmt) From(sel *SelectStmt) *CreateViewStmt {
	stmt.selectStmt = sel
	return stmt
}

// DDLSQL 返回创建视图的 SQL 语句
func (stmt *CreateViewStmt) DDLSQL() ([]string, error) {
	if stmt.name == "" {
		return nil, ErrTableIsEmpty
	}

	if stmt.selectStmt == nil {
		return nil, ErrValueIsEmpty
	}

	if hook, ok := stmt.Dialect().(CreateViewStmtHooker); ok {
		return hook.CreateViewStmtHook(stmt)
	}

	builder := New("CREATE ")

	if stmt.replace {
		builder.WriteString(" OR REPLACE ")
	}

	if stmt.temporary {
		builder.WriteString(" TEMPORARY ")
	}

	builder.WriteString(" VIEW ").
		WriteBytes(stmt.l).
		WriteString(stmt.name).
		WriteBytes(stmt.r)

	if len(stmt.columns) > 0 {
		builder.WriteBytes('(')
		for _, col := range stmt.columns {
			builder.WriteBytes(stmt.l).
				WriteString(col).
				WriteBytes(stmt.r, ',')
		}
		builder.TruncateLast(1)
		builder.WriteBytes(')')
	}

	q, args, err := stmt.selectStmt.SQL()
	if err != nil {
		return nil, err
	}
	if len(args) > 0 {
		return nil, ErrViewSelectNotAllowArgs
	}
	builder.WriteString(q)

	return []string{builder.String()}, nil
}
