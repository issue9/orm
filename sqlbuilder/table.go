// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/issue9/orm/v2/core"
)

// CreateTableStmt 创建表的语句
type CreateTableStmt struct {
	*ddlStmt

	name    string
	columns []*core.Column
	indexes []*indexColumn

	// 约束
	constraints []*constraintColumn
	foreignKeys []*foreignKey
	ai, pk      *constraintColumn

	// 一些附加的信息
	//
	// 比如可以指定创建表时的编码等，各个数据库各不相同。
	options map[string][]string
}

type foreignKey struct {
	Name                     string // 约束名
	Column                   string // 列名
	RefTableName, RefColName string
	UpdateRule, DeleteRule   string
}

type indexColumn struct {
	Name    string
	Type    core.Index
	Columns []string
}

type constraintColumn struct {
	Name    string
	Type    core.Constraint
	Columns []string
}

// CreateTable 创建表的语句
//
// 执行创建表操作，可能包含了创建索引等多个语句，
// 如果 e 是一个事务类型，且 e.Dialect() 是支持事务 DDL 的，
// 那么在执行时，会当作一个事务处理，否则为多个语句依次执行。
func CreateTable(e core.Engine) *CreateTableStmt {
	stmt := &CreateTableStmt{}
	stmt.ddlStmt = newDDLStmt(e, stmt)
	return stmt
}

// Reset 重置内容
func (stmt *CreateTableStmt) Reset() *CreateTableStmt {
	stmt.name = ""
	stmt.columns = stmt.columns[:0]
	stmt.indexes = stmt.indexes[:0]
	stmt.options = map[string][]string{}
	stmt.foreignKeys = stmt.foreignKeys[:0]
	stmt.constraints = stmt.constraints[:0]
	stmt.pk = nil
	stmt.ai = nil

	return stmt
}

// Table 指定表名
func (stmt *CreateTableStmt) Table(t string) *CreateTableStmt {
	stmt.name = t
	return stmt
}

func newColumn(name string, goType reflect.Type, ai, nullable, hasDefault bool, def interface{}, length ...int) (*core.Column, error) {
	col, err := core.NewColumnFromGoType(goType)
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
// goType Go 中的类型，该类型会被转换成相应的数据库类型；
// nullable 表示该列是否可以为 NULL；
// hasDefault 表示是否拥有默认值，如果为 true，则 v 同时会被当作默认值；
// def 默认值；
// length 表示长度信息。
func (stmt *CreateTableStmt) Column(name string, goType reflect.Type, nullable, hasDefault bool, def interface{}, length ...int) *CreateTableStmt {
	col, err := newColumn(name, goType, false, nullable, hasDefault, def, length...)
	if err != nil {
		panic(err)
	}
	return stmt.Columns(col)
}

// Columns 添加列
func (stmt *CreateTableStmt) Columns(col ...*core.Column) *CreateTableStmt {
	stmt.columns = append(stmt.columns, col...)

	return stmt
}

// AutoIncrement 指定自增列，自增列必定是主键。
// 如果指定了自增，则主键必定不启作用。
//
// col 列名；
// goType 对应的 Go 类型。
func (stmt *CreateTableStmt) AutoIncrement(col string, goType reflect.Type) *CreateTableStmt {
	stmt.ai = &constraintColumn{
		Type:    core.ConstraintAI,
		Columns: []string{col},
	}

	c, err := newColumn(col, goType, true, false, false, nil)
	if err != nil {
		panic(err)
	}
	return stmt.Columns(c)
}

// PK 指定主键约束
//
// 如果多次指定主键信息，则会 panic
// 自境会自动转换为主键
func (stmt *CreateTableStmt) PK(col ...string) *CreateTableStmt {
	if stmt.pk != nil || stmt.ai != nil {
		panic("主键或是自增列已经存在")
	}

	stmt.pk = &constraintColumn{
		Type:    core.ConstraintPK,
		Columns: col,
	}

	return stmt
}

// Index 添加索引
func (stmt *CreateTableStmt) Index(typ core.Index, name string, col ...string) *CreateTableStmt {
	stmt.indexes = append(stmt.indexes, &indexColumn{
		Name:    name,
		Type:    typ,
		Columns: col,
	})

	return stmt
}

// Unique 添加唯一约束
func (stmt *CreateTableStmt) Unique(name string, col ...string) *CreateTableStmt {
	stmt.constraints = append(stmt.constraints, &constraintColumn{
		Name:    name,
		Type:    core.ConstraintUnique,
		Columns: col,
	})

	return stmt
}

// Check check 约束
func (stmt *CreateTableStmt) Check(name string, expr string) *CreateTableStmt {
	stmt.constraints = append(stmt.constraints, &constraintColumn{
		Name:    name,
		Type:    core.ConstraintCheck,
		Columns: []string{expr},
	})

	return stmt
}

// ForeignKey 指定外键
func (stmt *CreateTableStmt) ForeignKey(name, col, refTable, refCol, updateRule, deleteRule string) *CreateTableStmt {
	stmt.foreignKeys = append(stmt.foreignKeys, &foreignKey{
		Name:         name,
		Column:       col,
		RefTableName: refTable,
		RefColName:   refCol,
		UpdateRule:   updateRule,
		DeleteRule:   deleteRule,
	})

	return stmt
}

func (stmt CreateTableStmt) checkNames() error {
	names := make([]string, 0, 2+len(stmt.indexes)+len(stmt.constraints)+len(stmt.foreignKeys))

	if stmt.ai != nil {
		names = append(names, stmt.ai.Name)
	}

	if stmt.pk != nil {
		names = append(names, stmt.pk.Name)
	}

	for _, index := range stmt.indexes {
		names = append(names, index.Name)
	}

	for _, constraint := range stmt.constraints {
		names = append(names, constraint.Name)
	}

	for _, fk := range stmt.foreignKeys {
		names = append(names, fk.Name)
	}

	sort.Strings(names)

	for i := 1; i < len(names); i++ {
		if names[i] == names[i-1] {
			return fmt.Errorf("存在相同的约束名 %s", names[i])
		}
	}

	return nil
}

// DDLSQL 获取 SQL 的语句及参数部分
func (stmt *CreateTableStmt) DDLSQL() ([]string, error) {
	if err := stmt.checkNames(); err != nil {
		return nil, err
	}

	if stmt.name == "" {
		return nil, ErrTableIsEmpty
	}

	if len(stmt.columns) == 0 {
		return nil, ErrColumnsIsEmpty
	}

	for _, col := range stmt.columns {
		if err := col.Check(); err != nil {
			return nil, err
		}
	}

	w := core.NewBuilder("CREATE TABLE IF NOT EXISTS ").
		QuoteKey(stmt.name).
		WriteBytes('(')

	for _, col := range stmt.columns {
		typ, err := stmt.Dialect().SQLType(col)
		if err != nil {
			return nil, err
		}
		w.QuoteKey(col.Name).
			WriteBytes(' ').
			WriteString(typ).
			WriteBytes(',')
	}

	if err := stmt.createConstraints(w); err != nil {
		return nil, err
	}

	w.TruncateLast(1).WriteBytes(')')

	if err := stmt.Dialect().CreateTableOptionsSQL(w, stmt.options); err != nil {
		return nil, err
	}

	sqls := []string{w.String()}

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
	for _, c := range stmt.constraints {
		switch c.Type {
		case core.ConstraintCheck:
			stmt.createCheckSQL(buf, c.Name, c.Columns[0])
		case core.ConstraintUnique:
			stmt.createUniqueSQL(buf, c.Name, c.Columns...)
		default:
			return ErrUnknownConstraint
		}
		buf.WriteBytes(',')
	}

	// foreign  key
	for _, fk := range stmt.foreignKeys {
		stmt.createFKSQL(buf, fk)
		buf.WriteBytes(',')
	}

	// primary key
	if stmt.pk != nil {
		stmt.createPKSQL(buf, core.PKName(stmt.name), stmt.pk.Columns...)
		buf.WriteBytes(',')
	}

	return nil
}

func createIndexSQL(stmt *CreateTableStmt) ([]string, error) {
	if len(stmt.indexes) == 0 {
		return nil, nil
	}

	sqls := make([]string, 0, len(stmt.indexes))
	buf := CreateIndex(stmt.Engine())
	for _, index := range stmt.indexes {
		buf.Reset()
		buf.Table(stmt.name).
			Name(index.Name).
			Columns(index.Columns...)

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
func (stmt *CreateTableStmt) createPKSQL(buf *core.Builder, name string, cols ...string) {
	buf.WriteString(" CONSTRAINT ").
		QuoteKey(name).
		WriteString(" PRIMARY KEY(")

	for _, col := range cols {
		buf.QuoteKey(col).WriteBytes(',')
	}
	buf.TruncateLast(1).WriteBytes(')')
}

// create table 语句中的 unique 约束部分的语句。
//
// CONSTRAINT unique_name UNIQUE (id,lastName)
func (stmt *CreateTableStmt) createUniqueSQL(buf *core.Builder, name string, cols ...string) {
	buf.WriteString(" CONSTRAINT ").
		QuoteKey(name).
		WriteString(" UNIQUE(")
	for _, col := range cols {
		buf.QuoteKey(col).WriteBytes(',')
	}
	buf.TruncateLast(1).WriteBytes(')')
}

// create table 语句中 fk 的约束部分的语句
func (stmt *CreateTableStmt) createFKSQL(buf *core.Builder, fk *foreignKey) {
	// CONSTRAINT fk_name FOREIGN KEY (id) REFERENCES user(id)
	buf.WriteString(" CONSTRAINT ").
		QuoteKey(fk.Name)

	buf.WriteString(" FOREIGN KEY (").
		QuoteKey(fk.Column)

	buf.WriteString(") REFERENCES ").
		QuoteKey(fk.RefTableName)

	buf.WriteBytes('(').
		QuoteKey(fk.RefColName).
		WriteBytes(')')

	if len(fk.UpdateRule) > 0 {
		buf.WriteString(" ON UPDATE ").WriteString(fk.UpdateRule)
	}

	if len(fk.DeleteRule) > 0 {
		buf.WriteString(" ON DELETE ").WriteString(fk.DeleteRule)
	}
}

// create table 语句中 check 约束部分的语句
func (stmt *CreateTableStmt) createCheckSQL(buf *core.Builder, name, expr string) {
	// CONSTRAINT chk_name CHECK (id>0 AND username='admin')
	buf.WriteString(" CONSTRAINT ").
		QuoteKey(name).
		WriteString(" CHECK(").
		WriteString(expr).
		WriteBytes(')')
}

// TruncateTableStmtHooker TruncateTableStmt.DDLSQL 的钩子函数
type TruncateTableStmtHooker interface {
	TruncateTableStmtHook(*TruncateTableStmt) ([]string, error)
}

// TruncateTableStmt 清空表，并重置 AI
type TruncateTableStmt struct {
	*ddlStmt
	TableName    string
	AIColumnName string
}

// TruncateTable 生成清空表语句
func TruncateTable(e core.Engine) *TruncateTableStmt {
	stmt := &TruncateTableStmt{}
	stmt.ddlStmt = newDDLStmt(e, stmt)
	return stmt
}

// Reset 重置内容
func (stmt *TruncateTableStmt) Reset() *TruncateTableStmt {
	stmt.TableName = ""
	stmt.AIColumnName = ""
	return stmt
}

// Table 指定表名
//
// aiColumn 表示自增列；
func (stmt *TruncateTableStmt) Table(t, aiColumn string) *TruncateTableStmt {
	stmt.TableName = t
	stmt.AIColumnName = aiColumn
	return stmt
}

// DDLSQL 获取 SQL 的语句及参数部分
func (stmt *TruncateTableStmt) DDLSQL() ([]string, error) {
	if hook, ok := stmt.Dialect().(TruncateTableStmtHooker); ok {
		return hook.TruncateTableStmtHook(stmt)
	}

	return nil, ErrNotImplemented
}

// DropTableStmt 删除表语句
type DropTableStmt struct {
	*ddlStmt
	tables []string
}

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
	if len(stmt.tables) == 0 {
		return nil, ErrTableIsEmpty
	}

	qs := make([]string, 0, len(stmt.tables))

	for _, table := range stmt.tables {
		buf := core.NewBuilder("DROP TABLE IF EXISTS ").
			QuoteKey(table)

		qs = append(qs, buf.String())
	}
	return qs, nil
}

// Reset 重置
func (stmt *DropTableStmt) Reset() *DropTableStmt {
	stmt.tables = stmt.tables[:0]
	return stmt
}
