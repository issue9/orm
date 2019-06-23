// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
)

// CreateTableStmt 创建表的语句
type CreateTableStmt struct {
	engine  Engine
	dialect Dialect

	name    string
	columns []*column
	indexes []*indexColumn

	// 约束
	constraints []*constraintColumn
	foreignKeys []*foreignKey // 外键约束
	ai, pk      *constraintColumn

	// 一些附加的信息
	//
	// 比如可以指定创建表时的编码等，各个数据库各不相同。
	options map[string][]string
}

type column struct {
	Name string // 数据库的字段名
	Type string // 类型，包含长度，可能是 BIGIINT，或是 VARCHAR(1024) 等格式
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

// Column 添加列
//
// name 列名
// typ 包括了长度 PK 等所有信息，比如 INT NOT NULL PRIMARY KEY AUTO_INCREMENT
func (stmt *CreateTableStmt) Column(name, typ string) *CreateTableStmt {
	col := &column{
		Name: name,
		Type: typ,
	}

	stmt.columns = append(stmt.columns, col)

	return stmt
}

// AutoIncrement 指定自增列，自增列必定是主键，如果已经存在主键，会替换主键内容
//
// name 自增约束的名称；
// col 自增对应的列；
func (stmt *CreateTableStmt) AutoIncrement(name, col string) *CreateTableStmt {
	stmt.ai = &constraintColumn{
		Name:    name,
		Type:    ConstraintAI,
		Columns: []string{col},
	}

	return stmt
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
func (stmt *CreateTableStmt) Index(name string, typ Index, col ...string) *CreateTableStmt {
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

// SQL 获取 SQL 的语句及参数部分
func (stmt *CreateTableStmt) SQL() ([]string, error) {
	w := New("CREATE TABLE IF NOT EXISTS ").
		WriteString(stmt.name).
		WriteByte('(')

	// 普通列
	for _, col := range stmt.columns {
		w.WriteString(col.Name).
			WriteByte(' ').
			WriteString(col.Type).
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
func (stmt *CreateTableStmt) Exec() (sql.Result, error) {
	return stmt.ExecContext(context.Background())
}

// ExecContext 执行 SQL 语句
func (stmt *CreateTableStmt) ExecContext(ctx context.Context) (sql.Result, error) {
	qs, err := stmt.SQL()
	if err != nil {
		return nil, err
	}

	for _, query := range qs {
		if r, err := stmt.engine.ExecContext(ctx, query); err != nil {
			return r, err
		}
	}

	return nil, nil
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

	return nil
}

func createIndexSQL(model *CreateTableStmt) ([]string, error) {
	if len(model.indexes) == 0 {
		return nil, nil
	}

	sqls := make([]string, 0, len(model.indexes))
	buf := CreateIndex(nil)
	for _, index := range model.indexes {
		buf.Reset()
		buf.Table(model.name).
			Name(index.Name)
		for _, col := range index.Columns {
			buf.Columns(col)
		}

		sql, _, err := buf.SQL()
		if err != nil {
			return nil, err
		}
		sqls = append(sqls, sql)
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
