// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
	"strconv"
)

// CreateTableStmt 创建表的语句
type CreateTableStmt struct {
	engine  Engine
	dialect Dialect

	Name    string
	Columns []*Column
	Indexes []*IndexColumn

	// 一些附加的信息
	//
	// 比如可以指定创建表时的编码等，各个数据库各不相同。
	Options map[string][]string

	// 约束
	ForeignKeys map[string]*ForeignKey // 外键约束
	Uniques     map[string][]string    // 唯一约束
	Checks      map[string]string      // check 约束
	PKs         map[string][]string    // 主键
	AI          *AutoIncrement         // 自增
}

// Column 列结构
type Column struct {
	Name       string // 数据库的字段名
	Type       string // 类型，包含长度，可能是 BIGIINT，或是 VARCHAR(1024) 等格式
	Nullable   bool   // 是否可以为 NULL
	Default    string // 默认值
	HasDefault bool
}

// ForeignKey 外键
type ForeignKey struct {
	Column                   string
	RefTableName, RefColName string
	UpdateRule, DeleteRule   string
}

// AutoIncrement 自增约束
type AutoIncrement struct {
	Name   string // 约束名
	Column string // 对应的列名
}

// IndexColumn 索引列
type IndexColumn struct {
	Name    string
	Type    Index
	Columns []string
}

// CreateTable 创建表的语句
func CreateTable(e Engine, d Dialect) *CreateTableStmt {
	return &CreateTableStmt{
		engine:  e,
		dialect: d,
	}
}

// Table 指定表名
func (stmt *CreateTableStmt) Table(t string) *CreateTableStmt {
	stmt.Name = t
	return stmt
}

// Column 添加列
//
// hasDefault 是否需要设置默认值；
// def 如果 hasDefault 为 true，则 def 为其默认值，否则 def 不启作用；
func (stmt *CreateTableStmt) Column(name, typ string, nullable, hasDefault bool, def string, length ...int) *CreateTableStmt {
	col := &Column{
		Name:       name,
		Nullable:   nullable,
		HasDefault: hasDefault,
	}
	if hasDefault {
		col.Default = def
	}

	switch len(length) {
	case 1:
		col.Type = typ + "(" + strconv.Itoa(length[0]) + ")"
	case 2:
		col.Type = typ + "(" + strconv.Itoa(length[0]) + "," + strconv.Itoa(length[1]) + ")"
	default:
		col.Type = typ
	}

	stmt.Columns = append(stmt.Columns, col)

	return stmt
}

// AutoIncrement 指定自增列
// col 自增对应的列；
// name 自增约束的名称；
// pk 是否同时设置为主键。
func (stmt *CreateTableStmt) AutoIncrement(col, name string, pk bool) *CreateTableStmt {
	stmt.AI = &AutoIncrement{
		Name:   name,
		Column: col,
	}

	if pk {
		return stmt.PK(name, col)
	}

	return stmt
}

// PK 指定主键约束
func (stmt *CreateTableStmt) PK(name string, col ...string) *CreateTableStmt {
	stmt.PKs[name] = col
	return stmt
}

// Index 添加索引
func (stmt *CreateTableStmt) Index(name string, typ Index, col ...string) *CreateTableStmt {
	stmt.Indexes = append(stmt.Indexes, &IndexColumn{
		Name:    name,
		Type:    typ,
		Columns: col,
	})

	return stmt
}

// Unique 添加唯一约束
func (stmt *CreateTableStmt) Unique(name string, col ...string) *CreateTableStmt {
	stmt.Uniques[name] = col

	return stmt
}

// Check check 约束
func (stmt *CreateTableStmt) Check(name string, expr string) *CreateTableStmt {
	stmt.Checks[name] = expr

	return stmt
}

// ForeignKey 指定外键
func (stmt *CreateTableStmt) ForeignKey(name, col, refTable, refCol, updateRule, deleteRule string) *CreateTableStmt {
	stmt.ForeignKeys[name] = &ForeignKey{
		Column:       col,
		RefTableName: refTable,
		RefColName:   refCol,
		UpdateRule:   updateRule,
		DeleteRule:   deleteRule,
	}

	return stmt
}

// SQL 获取 SQL 的语句及参数部分
func (stmt *CreateTableStmt) SQL() ([]string, error) {
	w := New("CREATE TABLE IF NOT EXISTS ").
		WriteString(stmt.Name).
		WriteByte('(')

	// 普通列
	for _, col := range stmt.Columns {
		err := stmt.dialect.CreateColumnSQL(w, col, col.Name == stmt.AI.Column)
		if err != nil {
			return nil, err
		}
		w.WriteByte(',')
	}

	stmt.createConstraints(w)

	w.TruncateLast(1).WriteByte(')')

	if err := stmt.dialect.CreateTableOptionsSQL(w, stmt.Options); err != nil {
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
	sqls, err := stmt.SQL()
	if err != nil {
		return nil, err
	}

	for _, sql := range sqls {
		if _, err := stmt.engine.ExecContext(ctx, sql, nil); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

// 创建标准的几种约束
func (stmt *CreateTableStmt) createConstraints(buf *SQLBuilder) {
	for name, cols := range stmt.PKs {
		createPKSQL(buf, name, cols...)
	}
	// unique
	for name, index := range stmt.Uniques {
		createUniqueSQL(buf, index, name)
		buf.WriteByte(',')
	}

	// foreign  key
	for name, fk := range stmt.ForeignKeys {
		createFKSQL(buf, fk, name)
		buf.WriteByte(',')
	}

	// Check
	for name, expr := range stmt.Checks {
		createCheckSQL(buf, expr, name)
		buf.WriteByte(',')
	}
}

func createIndexSQL(model *CreateTableStmt) ([]string, error) {
	if len(model.Indexes) == 0 {
		return nil, nil
	}

	sqls := make([]string, 0, len(model.Indexes))
	buf := CreateIndex(nil)
	for _, index := range model.Indexes {
		buf.Reset()
		buf.Table("{#" + model.Name + "}").
			Name(index.Name)
		for _, col := range index.Columns {
			buf.Columns("{" + col + "}")
		}

		sql, _, err := buf.SQL()
		if err != nil {
			return nil, err
		}
		sqls = append(sqls, sql)
	}

	return sqls, nil
}

// TODO 检测约束名是否唯一，检测约束中的列是否都存在

// 用于产生在 createTable 中使用的普通列信息表达式，不包含 autoincrement 和 primary key 的关键字。
func createColSQL(buf *SQLBuilder, col *Column) error {
	// col_name VARCHAR(100) NOT NULL DEFAULT 'abc'
	buf.WriteByte('{').WriteString(col.Name).WriteByte('}')
	buf.WriteByte(' ')

	// 写入字段类型
	buf.WriteString(col.Type)

	if !col.Nullable {
		buf.WriteString(" NOT NULL")
	}

	if col.HasDefault {
		buf.WriteString(" DEFAULT '").
			WriteString(col.Default).
			WriteByte('\'')
	}

	return nil
}

// create table 语句中 pk 约束的语句
func createPKSQL(buf *SQLBuilder, pkName string, cols ...string) {
	// CONSTRAINT pk_name PRIMARY KEY (id,lastName)
	buf.WriteString(" CONSTRAINT ").
		WriteString(pkName).
		WriteString(" PRIMARY KEY(")

	for _, col := range cols {
		buf.WriteByte('{').WriteString(col).WriteByte('}')
		buf.WriteByte(',')
	}
	buf.TruncateLast(1) // 去掉最后一个逗号
	buf.WriteByte(')')
}

// create table 语句中的 unique 约束部分的语句。
func createUniqueSQL(buf *SQLBuilder, cols []string, indexName string) {
	// CONSTRAINT unique_name UNIQUE (id,lastName)
	buf.WriteString(" CONSTRAINT ").
		WriteString(indexName).
		WriteString(" UNIQUE(")
	for _, col := range cols {
		buf.WriteByte('{').
			WriteString(col).
			WriteByte('}').
			WriteByte(',')
	}
	buf.TruncateLast(1) // 去掉最后一个逗号

	buf.WriteByte(')')
}

// create table 语句中 fk 的约束部分的语句
func createFKSQL(buf *SQLBuilder, fk *ForeignKey, fkName string) {
	// CONSTRAINT fk_name FOREIGN KEY (id) REFERENCES user(id)
	buf.WriteString(" CONSTRAINT ").WriteString(fkName)

	buf.WriteString(" FOREIGN KEY(")
	buf.WriteByte('{').WriteString(fk.Column).WriteByte('}')

	buf.WriteString(") REFERENCES ").
		WriteByte('{').
		WriteString(fk.RefTableName).
		WriteByte('}')

	buf.WriteByte('(')
	buf.WriteByte('{').WriteString(fk.RefColName).WriteByte('}')
	buf.WriteByte(')')

	if len(fk.UpdateRule) > 0 {
		buf.WriteString(" ON UPDATE ").WriteString(fk.UpdateRule)
	}

	if len(fk.DeleteRule) > 0 {
		buf.WriteString(" ON DELETE ").WriteString(fk.DeleteRule)
	}
}

// create table 语句中 check 约束部分的语句
func createCheckSQL(buf *SQLBuilder, expr, chkName string) {
	// CONSTRAINT chk_name CHECK (id>0 AND username='admin')
	buf.WriteString(" CONSTRAINT ").
		WriteString(chkName).
		WriteString(" CHECK(").
		WriteString(expr).
		WriteByte(')')
}
