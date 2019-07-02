// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import "context"

// Index 索引的类型
type Index int8

// 索引的类型
const (
	IndexDefault Index = iota // 普通的索引
	IndexUnique
)

// CreateIndexStmt 创建索引的语句
type CreateIndexStmt struct {
	engine Engine
	table  string
	name   string   // 索引名称
	cols   []string // 索引列
	typ    Index
}

func (t Index) String() string {
	switch t {
	case IndexDefault:
		return "INDEX"
	case IndexUnique:
		return "UNIQUE INDEX"
	default:
		return "<unknown>"
	}
}

// CreateIndex 声明一条 CreateIndexStmt 语句
func CreateIndex(e Engine) *CreateIndexStmt {
	return &CreateIndexStmt{
		engine: e,
		typ:    IndexDefault,
	}
}

// Table 指定表名
func (stmt *CreateIndexStmt) Table(tbl string) *CreateIndexStmt {
	stmt.table = tbl
	return stmt
}

// Name 指定索引名
func (stmt *CreateIndexStmt) Name(index string) *CreateIndexStmt {
	stmt.name = index
	return stmt
}

// Type 指定索引类型
func (stmt *CreateIndexStmt) Type(t Index) *CreateIndexStmt {
	stmt.typ = t
	return stmt
}

// Columns 列名
func (stmt *CreateIndexStmt) Columns(col ...string) *CreateIndexStmt {
	if stmt.cols == nil {
		stmt.cols = col
		return stmt
	}

	stmt.cols = append(stmt.cols, col...)
	return stmt
}

// DDLSQL 生成 SQL 语句
func (stmt *CreateIndexStmt) DDLSQL() ([]string, error) {
	if stmt.table == "" {
		return nil, ErrTableIsEmpty
	}

	if len(stmt.cols) == 0 {
		return nil, ErrColumnsIsEmpty
	}

	var builder *SQLBuilder

	if stmt.typ == IndexDefault {
		builder = New("CREATE INDEX ")
	} else {
		builder = New("CREATE UNIQUE INDEX ")
	}

	builder.WriteString(stmt.name).
		WriteString(" ON ").
		WriteString(stmt.table).WriteByte('(')
	for _, col := range stmt.cols {
		builder.WriteString(col).WriteByte(',')
	}
	builder.TruncateLast(1).WriteByte(')')

	return []string{builder.String()}, nil
}

// Reset 重置
func (stmt *CreateIndexStmt) Reset() {
	stmt.table = ""
	stmt.cols = stmt.cols[:0]
	stmt.name = ""
	stmt.typ = IndexDefault
}

// Exec 执行 SQL 语句
func (stmt *CreateIndexStmt) Exec() error {
	return stmt.ExecContext(context.Background())
}

// ExecContext 执行 SQL 语句
func (stmt *CreateIndexStmt) ExecContext(ctx context.Context) error {
	return ddlExecContext(ctx, stmt.engine, stmt)
}

// DropIndexStmtHooker DropIndexStmt.DDLSQL 的勾子函数
type DropIndexStmtHooker interface {
	DropIndexStmtHook(*DropIndexStmt) ([]string, error)
}

// DropIndexStmt 删除索引
type DropIndexStmt struct {
	engine    Engine
	dialect   Dialect
	TableName string
	IndexName string
}

// DropIndex 声明一条 DropIndexStmt 语句
func DropIndex(e Engine, d Dialect) *DropIndexStmt {
	return &DropIndexStmt{
		engine:  e,
		dialect: d,
	}
}

// Table 指定表名
func (stmt *DropIndexStmt) Table(tbl string) *DropIndexStmt {
	stmt.TableName = tbl
	return stmt
}

// Name 指定索引名
func (stmt *DropIndexStmt) Name(col string) *DropIndexStmt {
	stmt.IndexName = col
	return stmt
}

// DDLSQL 生成 SQL 语句
func (stmt *DropIndexStmt) DDLSQL() ([]string, error) {
	if stmt.TableName == "" {
		return nil, ErrTableIsEmpty
	}

	if stmt.IndexName == "" {
		return nil, ErrColumnsIsEmpty
	}

	if hook, ok := stmt.dialect.(DropIndexStmtHooker); ok {
		return hook.DropIndexStmtHook(stmt)
	}

	return []string{"DROP INDEX IF EXISTS " + stmt.IndexName}, nil
}

// Reset 重置
func (stmt *DropIndexStmt) Reset() {
	stmt.TableName = ""
	stmt.IndexName = ""
}

// Exec 执行 SQL 语句
func (stmt *DropIndexStmt) Exec() error {
	return stmt.ExecContext(context.Background())
}

// ExecContext 执行 SQL 语句
func (stmt *DropIndexStmt) ExecContext(ctx context.Context) error {
	return ddlExecContext(ctx, stmt.engine, stmt)
}
