// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package sqlbuilder

import "github.com/issue9/orm/v5/core"

// CreateIndexStmt 创建索引的语句
type CreateIndexStmt struct {
	*ddlStmt
	table string
	name  string   // 索引名称
	cols  []string // 索引列
	typ   core.IndexType
}

// CreateIndex 生成创建索引的语句
func (sql *SQLBuilder) CreateIndex() *CreateIndexStmt {
	return CreateIndex(sql.engine)
}

// CreateIndex 声明一条 CreateIndexStmt 语句
func CreateIndex(e core.Engine) *CreateIndexStmt {
	stmt := &CreateIndexStmt{typ: core.IndexDefault}
	stmt.ddlStmt = newDDLStmt(e, stmt)

	return stmt
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
func (stmt *CreateIndexStmt) Type(t core.IndexType) *CreateIndexStmt {
	stmt.typ = t
	return stmt
}

// Columns 列名
func (stmt *CreateIndexStmt) Columns(col ...string) *CreateIndexStmt {
	if stmt.err != nil {
		return stmt
	}

	if stmt.cols == nil {
		stmt.cols = col
		return stmt
	}

	stmt.cols = append(stmt.cols, col...)
	return stmt
}

// DDLSQL 生成 SQL 语句
func (stmt *CreateIndexStmt) DDLSQL() ([]string, error) {
	if stmt.err != nil {
		return nil, stmt.Err()
	}

	if stmt.table == "" {
		return nil, ErrTableIsEmpty
	}

	if len(stmt.cols) == 0 {
		return nil, ErrColumnsIsEmpty
	}

	var builder *core.Builder

	if stmt.typ == core.IndexDefault {
		builder = core.NewBuilder("CREATE INDEX ")
	} else {
		builder = core.NewBuilder("CREATE UNIQUE INDEX ")
	}

	builder.WString(stmt.name).
		WString(" ON ").
		QuoteKey(stmt.table).
		WBytes('(')
	for _, col := range stmt.cols {
		builder.QuoteKey(col).
			WBytes(',')
	}
	builder.TruncateLast(1).WBytes(')')

	query, err := builder.String()
	if err != nil {
		return nil, err
	}
	return []string{query}, nil
}

// Reset 重置
func (stmt *CreateIndexStmt) Reset() *CreateIndexStmt {
	stmt.baseStmt.Reset()
	stmt.table = ""
	stmt.cols = stmt.cols[:0]
	stmt.name = ""
	stmt.typ = core.IndexDefault

	return stmt
}

// DropIndexStmt 删除索引
type DropIndexStmt struct {
	*ddlStmt
	tableName string
	indexName string
}

// DropIndex 生成删除索引的语句
func (sql *SQLBuilder) DropIndex() *DropIndexStmt { return DropIndex(sql.engine) }

// DropIndex 声明一条 DropIndexStmt 语句
func DropIndex(e core.Engine) *DropIndexStmt {
	stmt := &DropIndexStmt{}
	stmt.ddlStmt = newDDLStmt(e, stmt)
	return stmt
}

// Table 指定表名
func (stmt *DropIndexStmt) Table(tbl string) *DropIndexStmt {
	stmt.tableName = tbl
	return stmt
}

// Name 指定索引名
func (stmt *DropIndexStmt) Name(col string) *DropIndexStmt {
	stmt.indexName = col
	return stmt
}

// DDLSQL 生成 SQL 语句
func (stmt *DropIndexStmt) DDLSQL() ([]string, error) {
	q, err := stmt.Dialect().DropIndexSQL(stmt.tableName, stmt.indexName)
	if err != nil {
		return nil, err
	}
	return []string{q}, nil
}

// Reset 重置
func (stmt *DropIndexStmt) Reset() *DropIndexStmt {
	stmt.baseStmt.Reset()
	stmt.tableName = ""
	stmt.indexName = ""
	return stmt
}
