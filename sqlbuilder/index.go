// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
)

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

// DropIndexStmt 删除索引
type DropIndexStmt struct {
	engine  Engine
	dialect Dialect
	table   string
	name    string
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
func (stmt *CreateIndexStmt) Name(col string) *CreateIndexStmt {
	stmt.name = col
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
		stmt.cols = make([]string, 0, len(col))
	}
	stmt.cols = append(stmt.cols, col...)

	return stmt
}

// SQL 生成 SQL 语句
func (stmt *CreateIndexStmt) SQL() (string, []interface{}, error) {
	if stmt.table == "" {
		return "", nil, ErrTableIsEmpty
	}

	if len(stmt.cols) == 0 {
		return "", nil, ErrColumnsIsEmpty
	}

	var sql *SQLBuilder

	if stmt.typ == IndexDefault {
		sql = New("CREATE INDEX ")
	} else {
		sql = New("CREATE UNIQUE INDEX ")
	}

	sql.WriteString(stmt.name).
		WriteString(" ON ").
		WriteString(stmt.table).WriteByte('(')
	for _, col := range stmt.cols {
		sql.WriteString(col).WriteByte(',')
	}
	sql.TruncateLast(1).WriteByte(')')

	return sql.String(), nil, nil
}

// Reset 重置
func (stmt *CreateIndexStmt) Reset() {
	stmt.table = ""
	stmt.cols = stmt.cols[:0]
	stmt.name = ""
	stmt.typ = IndexDefault
}

// Exec 执行 SQL 语句
func (stmt *CreateIndexStmt) Exec() (sql.Result, error) {
	return stmt.ExecContext(context.Background())
}

// ExecContext 执行 SQL 语句
func (stmt *CreateIndexStmt) ExecContext(ctx context.Context) (sql.Result, error) {
	return execContext(ctx, stmt.engine, stmt)
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
	stmt.table = tbl
	return stmt
}

// Name 指定索引名
func (stmt *DropIndexStmt) Name(col string) *DropIndexStmt {
	stmt.name = col
	return stmt
}

// SQL 生成 SQL 语句
func (stmt *DropIndexStmt) SQL() (string, []interface{}, error) {
	if stmt.table == "" {
		return "", nil, ErrTableIsEmpty
	}

	if stmt.name == "" {
		return "", nil, ErrColumnsIsEmpty
	}

	sql, args := stmt.dialect.DropIndexSQL(stmt.table, stmt.name)
	return sql, args, nil
}

// Reset 重置
func (stmt *DropIndexStmt) Reset() {
	stmt.table = ""
	stmt.name = ""
}

// Exec 执行 SQL 语句
func (stmt *DropIndexStmt) Exec() (sql.Result, error) {
	return stmt.ExecContext(context.Background())
}

// ExecContext 执行 SQL 语句
func (stmt *DropIndexStmt) ExecContext(ctx context.Context) (sql.Result, error) {
	return execContext(ctx, stmt.engine, stmt)
}
