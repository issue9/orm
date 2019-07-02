// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"fmt"
	"reflect"
	"sort"
)

// CreateTableStmt 创建表的语句
type CreateTableStmt struct {
	engine  Engine
	dialect Dialect

	name    string
	columns []*Column
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

// Column 列结构
type Column struct {
	Name       string       // 数据库的字段名
	GoType     reflect.Type // Go 语言中的数据类型
	AI         bool
	Nullable   bool
	HasDefault bool
	Default    interface{}
	Length     []int
}

type foreignKey struct {
	Name                     string // 约束名
	Column                   string // 列名
	RefTableName, RefColName string
	UpdateRule, DeleteRule   string
}

type indexColumn struct {
	Name    string
	Type    Index
	Columns []string
}

type constraintColumn struct {
	Name    string
	Type    Constraint
	Columns []string
}

// CreateTable 创建表的语句
//
// 执行创建表操作，可能包含了创建索引等多个语句，
// 如果 e 是一个事务类型，且 d 是支持事务 DDL 的，
// 那么在执行时，会当作一个事务处理，否则为多个语句依次执行。
func CreateTable(e Engine, d Dialect) *CreateTableStmt {
	return &CreateTableStmt{
		engine:  e,
		dialect: d,
	}
}

// Reset 重置内容
func (stmt *CreateTableStmt) Reset() {
	stmt.name = ""
	stmt.columns = stmt.columns[:0]
	stmt.indexes = stmt.indexes[:0]
	stmt.options = map[string][]string{}
	stmt.foreignKeys = stmt.foreignKeys[:0]
	stmt.constraints = stmt.constraints[:0]
	stmt.pk = nil
	stmt.ai = nil
}

// Table 指定表名
func (stmt *CreateTableStmt) Table(t string) *CreateTableStmt {
	stmt.name = t
	return stmt
}

func newColumn(name string, goType reflect.Type, ai, nullable, hasDefault bool, def interface{}, length ...int) *Column {
	return &Column{
		Name:       name,
		GoType:     goType,
		AI:         ai,
		Nullable:   nullable,
		HasDefault: hasDefault,
		Default:    def,
		Length:     length,
	}
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
	return stmt.Columns(newColumn(name, goType, false, nullable, hasDefault, def, length...))
}

// Columns 添加列
//
// name 列名
// typ 包括了长度 PK 等所有信息，比如 INT NOT NULL PRIMARY KEY AUTO_INCREMENT
func (stmt *CreateTableStmt) Columns(col ...*Column) *CreateTableStmt {
	stmt.columns = append(stmt.columns, col...)

	return stmt
}

// AutoIncrement 指定自增列，自增列必定是主键。
// 如果指定了自增，则主键必定不启作用。
//
// name 自增约束的名称；
// col 列名；
// goType 对应的 Go 类型。
func (stmt *CreateTableStmt) AutoIncrement(name, col string, goType reflect.Type) *CreateTableStmt {
	stmt.ai = &constraintColumn{
		Name:    name,
		Type:    ConstraintAI,
		Columns: []string{col},
	}

	return stmt.Columns(newColumn(col, goType, true, false, false, nil))
}

// PK 指定主键约束
//
// 如果多次指定主键信息，则会 panic
// 自境会自动转换为主键
func (stmt *CreateTableStmt) PK(name string, col ...string) *CreateTableStmt {
	if stmt.pk != nil || stmt.ai != nil {
		panic("主键或是自增列已经存在")
	}

	stmt.pk = &constraintColumn{
		Name:    name,
		Type:    ConstraintPK,
		Columns: col,
	}

	return stmt
}

// Index 添加索引
func (stmt *CreateTableStmt) Index(typ Index, name string, col ...string) *CreateTableStmt {
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
		Type:    ConstraintUnique,
		Columns: col,
	})

	return stmt
}

// Check check 约束
func (stmt *CreateTableStmt) Check(name string, expr string) *CreateTableStmt {
	stmt.constraints = append(stmt.constraints, &constraintColumn{
		Name:    name,
		Type:    ConstraintCheck,
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

	w := New("CREATE TABLE IF NOT EXISTS ").
		WriteString(stmt.name).
		WriteByte('(')

	// 普通列
	for _, col := range stmt.columns {
		typ, err := stmt.dialect.SQLType(col)
		if err != nil {
			return nil, err
		}
		w.WriteString(col.Name).
			WriteByte(' ').
			WriteString(typ).
			WriteByte(',')
	}

	if err := stmt.createConstraints(w); err != nil {
		return nil, err
	}

	w.TruncateLast(1).WriteByte(')')

	if err := stmt.dialect.CreateTableOptionsSQL(w, stmt.options); err != nil {
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

// Exec 执行 SQL 语句
func (stmt *CreateTableStmt) Exec() error {
	return stmt.ExecContext(context.Background())
}

// ExecContext 执行 SQL 语句
func (stmt *CreateTableStmt) ExecContext(ctx context.Context) error {
	return ddlExecContext(ctx, stmt.engine, stmt)
}

// 创建标准的几种约束
func (stmt *CreateTableStmt) createConstraints(buf *SQLBuilder) error {
	for _, c := range stmt.constraints {
		switch c.Type {
		case ConstraintCheck:
			createCheckSQL(buf, c.Name, c.Columns[0])
		case ConstraintUnique:
			createUniqueSQL(buf, c.Name, c.Columns...)
		default:
			return ErrUnknownConstraint
		}
		buf.WriteByte(',')
	}

	// foreign  key
	for _, fk := range stmt.foreignKeys {
		createFKSQL(buf, fk)
		buf.WriteByte(',')
	}

	// primary key
	if stmt.pk != nil {
		createPKSQL(buf, stmt.pk.Name, stmt.pk.Columns...)
		buf.WriteByte(',')
	}

	// TODO 部分数据库，需要独立创建 AI 约束，比如 Oracle

	return nil
}

func createIndexSQL(model *CreateTableStmt) ([]string, error) {
	if len(model.indexes) == 0 {
		return nil, nil
	}

	sqls := make([]string, 0, len(model.indexes))
	buf := CreateIndex(model.engine)
	for _, index := range model.indexes {
		buf.Reset()
		buf.Table(model.name).
			Name(index.Name).
			Columns(index.Columns...)

		query, _, err := buf.SQL()
		if err != nil {
			return nil, err
		}
		sqls = append(sqls, query)
	}

	return sqls, nil
}

// create table 语句中 pk 约束的语句
//
// CONSTRAINT pk_name PRIMARY KEY (id,lastName)
func createPKSQL(buf *SQLBuilder, name string, cols ...string) {
	buf.WriteString(" CONSTRAINT ").
		WriteString(name).
		WriteString(" PRIMARY KEY(")

	for _, col := range cols {
		buf.WriteString(col)
		buf.WriteByte(',')
	}
	buf.TruncateLast(1) // 去掉最后一个逗号
	buf.WriteByte(')')
}

// create table 语句中的 unique 约束部分的语句。
//
// CONSTRAINT unique_name UNIQUE (id,lastName)
func createUniqueSQL(buf *SQLBuilder, name string, cols ...string) {
	buf.WriteString(" CONSTRAINT ").
		WriteString(name).
		WriteString(" UNIQUE(")
	for _, col := range cols {
		buf.WriteString(col).
			WriteByte(',')
	}
	buf.TruncateLast(1) // 去掉最后一个逗号

	buf.WriteByte(')')
}

// create table 语句中 fk 的约束部分的语句
func createFKSQL(buf *SQLBuilder, fk *foreignKey) {
	// CONSTRAINT fk_name FOREIGN KEY (id) REFERENCES user(id)
	buf.WriteString(" CONSTRAINT ").WriteString(fk.Name)

	buf.WriteString(" FOREIGN KEY(")
	buf.WriteString(fk.Column)

	buf.WriteString(") REFERENCES ").
		WriteString(fk.RefTableName)

	buf.WriteByte('(')
	buf.WriteString(fk.RefColName)
	buf.WriteByte(')')

	if len(fk.UpdateRule) > 0 {
		buf.WriteString(" ON UPDATE ").WriteString(fk.UpdateRule)
	}

	if len(fk.DeleteRule) > 0 {
		buf.WriteString(" ON DELETE ").WriteString(fk.DeleteRule)
	}
}

// create table 语句中 check 约束部分的语句
func createCheckSQL(buf *SQLBuilder, name, expr string) {
	// CONSTRAINT chk_name CHECK (id>0 AND username='admin')
	buf.WriteString(" CONSTRAINT ").
		WriteString(name).
		WriteString(" CHECK(").
		WriteString(expr).
		WriteByte(')')
}

// TruncateTableStmtHooker TruncateTableStmt.DDLSQL 的勾子函数
type TruncateTableStmtHooker interface {
	TruncateTableStmtHook(*TruncateTableStmt) ([]string, error)
}

// TruncateTableStmt 清空表，并重置 AI
type TruncateTableStmt struct {
	engine       Engine
	dialect      Dialect
	TableName    string
	AIColumnName string
	AIName       string // 约束名
}

// TruncateTable 生成清空表语句
func TruncateTable(e Engine, d Dialect) *TruncateTableStmt {
	return &TruncateTableStmt{
		engine:  e,
		dialect: d,
	}
}

// Reset 重置内容
func (stmt *TruncateTableStmt) Reset() {
	stmt.TableName = ""
	stmt.AIColumnName = ""
}

// Table 指定表名
//
// aiName 表示自增约束的约束名，大部分数据库可以为空；
// aiColumn 表示自增列；
func (stmt *TruncateTableStmt) Table(t, aiColumn, aiName string) *TruncateTableStmt {
	stmt.TableName = t
	stmt.AIColumnName = aiColumn
	stmt.AIName = aiName
	return stmt
}

// DDLSQL 获取 SQL 的语句及参数部分
func (stmt *TruncateTableStmt) DDLSQL() ([]string, error) {
	if hook, ok := stmt.dialect.(TruncateTableStmtHooker); ok {
		return hook.TruncateTableStmtHook(stmt)
	}

	return nil, ErrNotImplemented
}

// Exec 执行 SQL 语句
func (stmt *TruncateTableStmt) Exec() error {
	return stmt.ExecContext(context.Background())
}

// ExecContext 执行 SQL 语句
func (stmt *TruncateTableStmt) ExecContext(ctx context.Context) error {
	return ddlExecContext(ctx, stmt.engine, stmt)
}
