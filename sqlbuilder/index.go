// SPDX-License-Identifier: MIT

package sqlbuilder

import "github.com/issue9/orm/v3/core"

// CreateIndexStmt 创建索引的语句
type CreateIndexStmt struct {
	*ddlStmt
	table string
	name  string   // 索引名称
	cols  []string // 索引列
	typ   core.Index
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
func (stmt *CreateIndexStmt) Type(t core.Index) *CreateIndexStmt {
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

	builder.WriteString(stmt.name).
		WriteString(" ON ").
		QuoteKey(stmt.table).
		WriteBytes('(')
	for _, col := range stmt.cols {
		builder.QuoteKey(col).
			WriteBytes(',')
	}
	builder.TruncateLast(1).WriteBytes(')')

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

// DropIndexStmtHooker DropIndexStmt.DDLSQL 的勾子函数
type DropIndexStmtHooker interface {
	DropIndexStmtHook(*DropIndexStmt) ([]string, error)
}

// DropIndexStmt 删除索引
type DropIndexStmt struct {
	*ddlStmt
	TableName string
	IndexName string
}

// CreateIndex 生成创建索引的语句
func (sql *SQLBuilder) DropIndex() *DropIndexStmt {
	return DropIndex(sql.engine)
}

// DropIndex 声明一条 DropIndexStmt 语句
func DropIndex(e core.Engine) *DropIndexStmt {
	stmt := &DropIndexStmt{}
	stmt.ddlStmt = newDDLStmt(e, stmt)
	return stmt
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
	if stmt.err != nil {
		return nil, stmt.Err()
	}

	if stmt.TableName == "" {
		return nil, ErrTableIsEmpty
	}

	if stmt.IndexName == "" {
		return nil, ErrColumnsIsEmpty
	}

	if hook, ok := stmt.Dialect().(DropIndexStmtHooker); ok {
		return hook.DropIndexStmtHook(stmt)
	}

	query, err := core.NewBuilder("DROP INDEX ").
		QuoteKey(stmt.IndexName).
		String()
	if err != nil {
		return nil, err
	}
	return []string{query}, nil
}

// Reset 重置
func (stmt *DropIndexStmt) Reset() *DropIndexStmt {
	stmt.baseStmt.Reset()
	stmt.TableName = ""
	stmt.IndexName = ""
	return stmt
}
