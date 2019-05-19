// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package dialect 提供了部分数据库对 orm.Dialect 接口的实现。
package dialect

import (
	"database/sql"
	"reflect"
	"time"

	"github.com/issue9/orm/v2"
	"github.com/issue9/orm/v2/sqlbuilder"
)

const pkName = "pk" // 默认的主键约束名

var (
	nullString  = reflect.TypeOf(sql.NullString{})
	nullInt64   = reflect.TypeOf(sql.NullInt64{})
	nullBool    = reflect.TypeOf(sql.NullBool{})
	nullFloat64 = reflect.TypeOf(sql.NullFloat64{})
	rawBytes    = reflect.TypeOf(sql.RawBytes{})
	timeType    = reflect.TypeOf(time.Time{})
)

type base interface {
	orm.Dialect

	// 将 col 转换成 sql 类型，并写入 buf 中。
	sqlType(buf *sqlbuilder.SQLBuilder, col *orm.Column) error
}

// 用于产生在 createTable 中使用的普通列信息表达式，不包含 autoincrement 和 primary key 的关键字。
func createColSQL(b base, buf *sqlbuilder.SQLBuilder, col *orm.Column) error {
	// col_name VARCHAR(100) NOT NULL DEFAULT 'abc'
	buf.WriteByte('{').WriteString(col.Name).WriteByte('}')
	buf.WriteByte(' ')

	// 写入字段类型
	if err := b.sqlType(buf, col); err != nil {
		return err
	}

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
func createPKSQL(buf *sqlbuilder.SQLBuilder, cols []*orm.Column, pkName string) {
	// CONSTRAINT pk_name PRIMARY KEY (id,lastName)
	buf.WriteString(" CONSTRAINT ").
		WriteString(pkName).
		WriteString(" PRIMARY KEY(")

	for _, col := range cols {
		buf.WriteByte('{').WriteString(col.Name).WriteByte('}')
		buf.WriteByte(',')
	}
	buf.TruncateLast(1) // 去掉最后一个逗号
	buf.WriteByte(')')
}

// create table 语句中的 unique 约束部分的语句。
func createUniqueSQL(buf *sqlbuilder.SQLBuilder, cols []*orm.Column, indexName string) {
	// CONSTRAINT unique_name UNIQUE (id,lastName)
	buf.WriteString(" CONSTRAINT ").
		WriteString(indexName).
		WriteString(" UNIQUE(")
	for _, col := range cols {
		buf.WriteByte('{').WriteString(col.Name).WriteByte('}')
		buf.WriteByte(',')
	}
	buf.TruncateLast(1) // 去掉最后一个逗号

	buf.WriteByte(')')
}

// create table 语句中 fk 的约束部分的语句
func createFKSQL(buf *sqlbuilder.SQLBuilder, fk *orm.ForeignKey, fkName string) {
	// CONSTRAINT fk_name FOREIGN KEY (id) REFERENCES user(id)
	buf.WriteString(" CONSTRAINT ").WriteString(fkName)

	buf.WriteString(" FOREIGN KEY(")
	buf.WriteByte('{').WriteString(fk.Col.Name).WriteByte('}')

	buf.WriteString(") REFERENCES ").WriteString(fk.RefTableName)

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
func createCheckSQL(buf *sqlbuilder.SQLBuilder, expr, chkName string) {
	// CONSTRAINT chk_name CHECK (id>0 AND username='admin')
	buf.WriteString(" CONSTRAINT ").
		WriteString(chkName).
		WriteString(" CHECK(").
		WriteString(expr).
		WriteByte(')')
}

// 创建标准的几种约束(除 PK 约束，该约束有专门的函数 createPKSQL() 产生)：unique, foreign key, check
func createConstraints(buf *sqlbuilder.SQLBuilder, model *orm.Model) {
	// Unique Index
	for name, index := range model.UniqueIndexes {
		createUniqueSQL(buf, index, name)
		buf.WriteByte(',')
	}

	// foreign  key
	for name, fk := range model.FK {
		createFKSQL(buf, fk, name)
		buf.WriteByte(',')
	}

	// Check
	for name, chk := range model.Check {
		createCheckSQL(buf, chk, name)
		buf.WriteByte(',')
	}
}

func createIndexSQL(model *orm.Model) ([]string, error) {
	if len(model.KeyIndexes) == 0 {
		return nil, nil
	}

	sqls := make([]string, 0, len(model.KeyIndexes))
	buf := sqlbuilder.CreateIndex(nil)
	for name, cols := range model.KeyIndexes {
		buf.Reset()
		buf.Table("{#" + model.Name + "}").Name(name)
		for _, col := range cols {
			buf.Columns("{" + col.Name + "}")
		}

		sql, _, err := buf.SQL()
		if err != nil {
			return nil, err
		}
		sqls = append(sqls, sql)
	}

	return sqls, nil
}

// mysql 系列数据库分页语法的实现。支持以下数据库：
// MySQL, H2, HSQLDB, Postgres, SQLite3
func mysqlLimitSQL(limit interface{}, offset ...interface{}) (string, []interface{}) {
	query := " LIMIT "

	if named, ok := limit.(sql.NamedArg); ok && named.Name != "" {
		query += "@" + named.Name
	} else {
		query += "?"
	}

	if len(offset) == 0 {
		return query + " ", []interface{}{limit}
	}

	query += " OFFSET "
	o := offset[0]
	if named, ok := o.(sql.NamedArg); ok && named.Name != "" {
		query += "@" + named.Name
	} else {
		query += "?"
	}

	return query + " ", []interface{}{limit, offset[0]}
}

// oracle系列数据库分页语法的实现。支持以下数据库：
// Derby, SQL Server 2012, Oracle 12c, the SQL 2008 standard
func oracleLimitSQL(limit interface{}, offset ...interface{}) (string, []interface{}) {
	query := "FETCH NEXT "

	if named, ok := limit.(sql.NamedArg); ok && named.Name != "" {
		query += "@" + named.Name
	} else {
		query += "?"
	}
	query += " ROWS ONLY "

	if len(offset) == 0 {
		return query, []interface{}{limit}
	}

	o := offset[0]
	if named, ok := o.(sql.NamedArg); ok && named.Name != "" {
		query = "OFFSET @" + named.Name + " ROWS " + query
	} else {
		query = "OFFSET ? ROWS " + query
	}

	return query, []interface{}{offset[0], limit}
}
