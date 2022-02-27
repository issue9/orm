// SPDX-License-Identifier: MIT

package sqlbuilder

import (
	"fmt"

	"github.com/issue9/orm/v4/core"
)

// CreateTableStmt 创建表的语句
type CreateTableStmt struct {
	*ddlStmt
	model *core.Model
}

// CreateTable 生成创建表的语句
func (sql *SQLBuilder) CreateTable() *CreateTableStmt {
	return CreateTable(sql.engine)
}

// CreateTable 创建表的语句
//
// 执行创建表操作，可能包含了创建索引等多个语句，
// 如果 e 是一个事务类型，且 e.Dialect() 是支持事务 DDL 的，
// 那么在执行时，会当作一个事务处理，否则为多个语句依次执行。
func CreateTable(e core.Engine) *CreateTableStmt {
	stmt := &CreateTableStmt{
		model: core.NewModel(core.Table, "", 10),
	}
	stmt.ddlStmt = newDDLStmt(e, stmt)
	return stmt
}

func (stmt *CreateTableStmt) Reset() *CreateTableStmt {
	stmt.baseStmt.Reset()
	stmt.model.Reset()
	return stmt
}

// Table 指定表名
func (stmt *CreateTableStmt) Table(t string) *CreateTableStmt {
	stmt.model.Type = core.Table
	stmt.model.Name = t
	return stmt
}

func newColumn(name string, p core.PrimitiveType, ai, nullable, hasDefault bool, def any, length ...int) (*core.Column, error) {
	col, err := core.NewColumn(p)
	if err != nil {
		return nil, err
	}

	col.Name = name
	col.AI = ai
	col.Nullable = nullable
	col.HasDefault = hasDefault
	col.Default = def
	col.Length = length

	return col, nil
}

// Column 添加列
//
// name 列的名称；
// p Go 中的类型，该类型会被转换成相应的数据库类型；
// ai 是否自增列；
// nullable 表示该列是否可以为 NULL；
// hasDefault 表示是否拥有默认值，如果为 true，则 v 同时会被当作默认值；
// def 默认值；
// length 表示长度信息。
func (stmt *CreateTableStmt) Column(name string, p core.PrimitiveType, ai, nullable, hasDefault bool, def any, length ...int) *CreateTableStmt {
	if stmt.err != nil {
		return stmt
	}

	var col *core.Column
	col, stmt.err = newColumn(name, p, ai, nullable, hasDefault, def, length...)
	return stmt.Columns(col)
}

// Columns 添加列
func (stmt *CreateTableStmt) Columns(col ...*core.Column) *CreateTableStmt {
	if stmt.err != nil {
		return stmt
	}

	stmt.err = stmt.model.AddColumns(col...)
	return stmt
}

// AutoIncrement 指定自增列
//
// 自增列必定是主键。如果指定了自增，则主键必定不启作用。
// 功能与 Column() 中将 ai 设置 true 是一样的。
//
// col 列名；
func (stmt *CreateTableStmt) AutoIncrement(col string, p core.PrimitiveType) *CreateTableStmt {
	return stmt.Column(col, p, true, false, false, nil)
}

// PK 指定主键约束
//
// name 为约束名，部分数据会忽略约束名，比如 mysql；
func (stmt *CreateTableStmt) PK(name string, col ...string) *CreateTableStmt {
	if stmt.err != nil {
		return stmt
	}

	if stmt.model.PrimaryKey != nil && stmt.model.PrimaryKey.Name != name {
		stmt.err = fmt.Errorf("已经存在名为 %s 的主键约束", stmt.model.PrimaryKey.Name)
		return stmt
	} else {
		stmt.model.PrimaryKey = &core.Constraint{Name: name}
	}

	for _, c := range col {
		if stmt.err = stmt.model.AddPrimaryKey(stmt.model.FindColumn(c)); stmt.err != nil {
			return stmt
		}
	}
	return stmt
}

// Index 添加索引
func (stmt *CreateTableStmt) Index(typ core.IndexType, name string, col ...string) *CreateTableStmt {
	if stmt.err != nil {
		return stmt
	}

	for _, c := range col {
		if stmt.err != nil {
			return stmt
		}

		stmt.err = stmt.model.AddIndex(typ, name, stmt.model.FindColumn(c))
	}
	return stmt
}

// Unique 添加唯一约束
func (stmt *CreateTableStmt) Unique(name string, col ...string) *CreateTableStmt {
	if stmt.err != nil {
		return stmt
	}

	for _, c := range col {
		if stmt.err != nil {
			return stmt
		}

		stmt.err = stmt.model.AddUnique(name, stmt.model.FindColumn(c))
	}

	return stmt
}

// Check 指定 check 约束
func (stmt *CreateTableStmt) Check(name string, expr string) *CreateTableStmt {
	if stmt.err != nil {
		return stmt
	}

	stmt.err = stmt.model.NewCheck(name, expr)
	return stmt
}

// ForeignKey 指定外键
func (stmt *CreateTableStmt) ForeignKey(name, col, refTable, refCol, updateRule, deleteRule string) *CreateTableStmt {
	if stmt.err != nil {
		return stmt
	}

	stmt.err = stmt.model.NewForeignKey(&core.ForeignKey{
		Name:         name,
		Column:       stmt.model.FindColumn(col),
		RefTableName: refTable,
		RefColName:   refCol,
		UpdateRule:   updateRule,
		DeleteRule:   deleteRule,
	})

	return stmt
}

// DDLSQL 获取 SQL 的语句及参数部分
func (stmt *CreateTableStmt) DDLSQL() ([]string, error) {
	if stmt.err != nil {
		return nil, stmt.Err()
	}

	if err := stmt.model.Sanitize(); err != nil {
		return nil, err
	}

	if len(stmt.model.Columns) == 0 {
		return nil, ErrColumnsIsEmpty
	}

	w := core.NewBuilder("CREATE TABLE IF NOT EXISTS ").
		QuoteKey(stmt.model.Name).
		WBytes('(')

	for _, col := range stmt.model.Columns {
		typ, err := stmt.Dialect().SQLType(col)
		if err != nil {
			return nil, err
		}
		w.QuoteKey(col.Name).WBytes(' ').WString(typ).WBytes(',')
	}

	if err := stmt.createConstraints(w); err != nil {
		return nil, err
	}
	w.TruncateLast(1).WBytes(')')

	if err := stmt.Dialect().CreateTableOptionsSQL(w, stmt.model.Meta); err != nil {
		return nil, err
	}

	q, err := w.String()
	if err != nil {
		return nil, err
	}
	sqls := []string{q}

	indexes, err := createIndexSQL(stmt)
	if err != nil {
		return nil, err
	}
	if len(indexes) > 0 {
		sqls = append(sqls, indexes...)
	}

	return sqls, nil
}

// 创建标准的几种约束
func (stmt *CreateTableStmt) createConstraints(buf *core.Builder) error {
	for name, expr := range stmt.model.Checks {
		stmt.createCheckSQL(buf, name, expr)
		buf.WBytes(',')
	}

	for _, u := range stmt.model.Uniques {
		stmt.createUniqueSQL(buf, u)
		buf.WBytes(',')
	}

	// foreign  key
	for _, fk := range stmt.model.ForeignKeys {
		stmt.createFKSQL(buf, fk)
		buf.WBytes(',')
	}

	// primary key
	if stmt.model.PrimaryKey != nil {
		stmt.createPKSQL(buf, stmt.model.PrimaryKey)
		buf.WBytes(',')
	}

	return nil
}

func createIndexSQL(stmt *CreateTableStmt) ([]string, error) {
	if len(stmt.model.Indexes) == 0 {
		return nil, nil
	}

	sqls := make([]string, 0, len(stmt.model.Indexes))
	buf := CreateIndex(stmt.Engine())
	for _, index := range stmt.model.Indexes {
		buf.Reset()
		buf.Table(stmt.model.Name).
			Name(index.Name)
		for _, col := range index.Columns {
			buf.Columns(col.Name)
		}

		query, err := buf.DDLSQL()
		if err != nil {
			return nil, err
		}
		sqls = append(sqls, query...)
	}

	return sqls, nil
}

// create table 语句中 pk 约束的语句
//
// CONSTRAINT pk_name PRIMARY KEY (id,lastName)
func (stmt *CreateTableStmt) createPKSQL(buf *core.Builder, c *core.Constraint) {
	buf.WString(" CONSTRAINT ").
		QuoteKey(c.Name).
		WString(" PRIMARY KEY(")

	for _, col := range c.Columns {
		buf.QuoteKey(col.Name).WBytes(',')
	}
	buf.TruncateLast(1).WBytes(')')
}

// create table 语句中的 unique 约束部分的语句。
//
// CONSTRAINT unique_name UNIQUE (id,lastName)
func (stmt *CreateTableStmt) createUniqueSQL(buf *core.Builder, c *core.Constraint) {
	buf.WString(" CONSTRAINT ").
		QuoteKey(c.Name).
		WString(" UNIQUE(")
	for _, col := range c.Columns {
		buf.QuoteKey(col.Name).WBytes(',')
	}
	buf.TruncateLast(1).WBytes(')')
}

// create table 语句中 fk 的约束部分的语句
func (stmt *CreateTableStmt) createFKSQL(buf *core.Builder, fk *core.ForeignKey) {
	// CONSTRAINT fk_name FOREIGN KEY (id) REFERENCES user(id)
	buf.WString(" CONSTRAINT ").
		QuoteKey(fk.Name)

	buf.WString(" FOREIGN KEY (").
		QuoteKey(fk.Column.Name)

	buf.WString(") REFERENCES ").
		QuoteKey(fk.RefTableName)

	buf.WBytes('(').
		QuoteKey(fk.RefColName).
		WBytes(')')

	if len(fk.UpdateRule) > 0 {
		buf.WString(" ON UPDATE ").WString(fk.UpdateRule)
	}

	if len(fk.DeleteRule) > 0 {
		buf.WString(" ON DELETE ").WString(fk.DeleteRule)
	}
}

// create table 语句中 check 约束部分的语句
//
// CONSTRAINT chk_name CHECK (id>0 AND username='admin')
func (stmt *CreateTableStmt) createCheckSQL(buf *core.Builder, name, expr string) {
	buf.WString(" CONSTRAINT ").
		QuoteKey(name).
		WString(" CHECK(").
		WString(expr).
		WBytes(')')
}

// TruncateTableStmt 清空表，并重置 AI
type TruncateTableStmt struct {
	*ddlStmt
	tableName    string
	aiColumnName string
}

// TruncateTable 生成清空表的语句，同时重置 AI 计算
func (sql *SQLBuilder) TruncateTable() *TruncateTableStmt {
	return TruncateTable(sql.engine)
}

// TruncateTable 生成清空表语句
func TruncateTable(e core.Engine) *TruncateTableStmt {
	stmt := &TruncateTableStmt{}
	stmt.ddlStmt = newDDLStmt(e, stmt)
	return stmt
}

// Reset 重置内容
func (stmt *TruncateTableStmt) Reset() *TruncateTableStmt {
	stmt.baseStmt.Reset()
	stmt.tableName = ""
	stmt.aiColumnName = ""
	return stmt
}

// Table 指定表名
//
// aiColumn 表示自增列；
func (stmt *TruncateTableStmt) Table(t, aiColumn string) *TruncateTableStmt {
	stmt.tableName = t
	stmt.aiColumnName = aiColumn
	return stmt
}

// DDLSQL 获取 SQL 的语句及参数部分
func (stmt *TruncateTableStmt) DDLSQL() ([]string, error) {
	if stmt.err != nil {
		return nil, stmt.Err()
	}

	return stmt.Dialect().TruncateTableSQL(stmt.tableName, stmt.aiColumnName)
}

// DropTableStmt 删除表语句
type DropTableStmt struct {
	*ddlStmt
	tables []string
}

// DropTable 生成删除表的语句
func (sql *SQLBuilder) DropTable() *DropTableStmt { return DropTable(sql.engine) }

// DropTable 声明一条删除表的语句
func DropTable(e core.Engine) *DropTableStmt {
	stmt := &DropTableStmt{}
	stmt.ddlStmt = newDDLStmt(e, stmt)
	return stmt
}

// Table 指定表名。
//
// 多次指定，则会删除多个表
func (stmt *DropTableStmt) Table(table ...string) *DropTableStmt {
	if stmt.tables == nil {
		stmt.tables = table
		return stmt
	}

	stmt.tables = append(stmt.tables, table...)
	return stmt
}

// DDLSQL 获取 SQL 语句以及对应的参数
func (stmt *DropTableStmt) DDLSQL() ([]string, error) {
	if stmt.err != nil {
		return nil, stmt.Err()
	}

	if len(stmt.tables) == 0 {
		return nil, ErrTableIsEmpty
	}

	qs := make([]string, 0, len(stmt.tables))

	for _, table := range stmt.tables {
		q, err := core.NewBuilder("DROP TABLE IF EXISTS ").
			QuoteKey(table).
			String()
		if err != nil {
			return nil, err
		}

		qs = append(qs, q)
	}
	return qs, nil
}

// Reset 重置
func (stmt *DropTableStmt) Reset() *DropTableStmt {
	stmt.tables = stmt.tables[:0]
	return stmt
}
