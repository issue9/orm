// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import "context"

// CreateViewStmt 创建视图的语句
type CreateViewStmt struct {
	engine      Engine
	name        string
	selectStmt  *SelectStmt
	columns     []string
	checkOption string
	temporary   bool
	replace     bool
}

// CreateView 创建视图
func CreateView(e Engine) *CreateViewStmt {
	return &CreateViewStmt{
		engine: e,
	}
}

// View 将当前查询语句转换为视图
func (stmt *SelectStmt) View(name string) *CreateViewStmt {
	return CreateView(stmt.engine).From(stmt)
}

// Reset 重置对象
func (stmt *CreateViewStmt) Reset() {
	stmt.name = ""
	stmt.selectStmt = nil
	stmt.columns = stmt.columns[:0]
	stmt.temporary = false
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

	builder := New("CREATE ")

	if stmt.replace {
		builder.WriteString(" OR REPLACE ")
	}

	if stmt.temporary {
		builder.WriteString(" TEMPORARY ")
	}

	builder.WriteString(" VIEW ").WriteString(stmt.name)

	if len(stmt.columns) > 0 {
		builder.WriteByte('(')
		for _, col := range stmt.columns {
			builder.WriteString(col).WriteByte(',')
		}
		builder.TruncateLast(1)
		builder.WriteByte(')')
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

// Exec 执行 SQL 语句
func (stmt *CreateViewStmt) Exec() error {
	return stmt.ExecContext(context.Background())
}

// ExecContext 执行 SQL 语句
func (stmt *CreateViewStmt) ExecContext(ctx context.Context) error {
	return ddlExecContext(ctx, stmt.engine, stmt)
}
