// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"bytes"
	"database/sql"
	"reflect"
	"time"

	"github.com/issue9/orm/forward"
)

const (
	pkName = "pk" // 默认的主键约束名
)

var (
	nullString  = reflect.TypeOf(sql.NullString{})
	nullInt64   = reflect.TypeOf(sql.NullInt64{})
	nullBool    = reflect.TypeOf(sql.NullBool{})
	nullFloat64 = reflect.TypeOf(sql.NullFloat64{})
	timeType    = reflect.TypeOf(time.Time{})
)

type base interface {
	forward.Dialect

	// 将col转换成sql类型，并写入buf中。
	sqlType(buf *bytes.Buffer, col *forward.Column) error
}

// 用于产生在createTable中使用的普通列信息表达式，不包含autoincrement和primary key的关键字。
func createColSQL(b base, buf *bytes.Buffer, col *forward.Column) error {
	// col_name VARCHAR(100) NOT NULL DEFAULT 'abc'
	b.Quote(buf, col.Name)
	buf.WriteByte(' ')

	// 写入字段类型
	if err := b.sqlType(buf, col); err != nil {
		return err
	}

	if !col.Nullable {
		buf.WriteString(" NOT NULL")
	}

	if col.HasDefault {
		buf.WriteString(" DEFAULT '")
		buf.WriteString(col.Default)
		buf.WriteByte('\'')
	}

	return nil
}

// create table语句中pk约束的语句
func createPKSQL(b base, buf *bytes.Buffer, cols []*forward.Column, pkName string) {
	//CONSTRAINT pk_name PRIMARY KEY (id,lastName)
	buf.WriteString(" CONSTRAINT ")
	buf.WriteString(pkName)
	buf.WriteString(" PRIMARY KEY(")
	for _, col := range cols {
		b.Quote(buf, col.Name)
		buf.WriteByte(',')
	}
	buf.Truncate(buf.Len() - 1) // 去掉最后一个逗号

	buf.WriteByte(')')
}

// create table语句中的unique约束部分的语句。
func createUniqueSQL(b base, buf *bytes.Buffer, cols []*forward.Column, indexName string) {
	//CONSTRAINT unique_name UNIQUE (id,lastName)
	buf.WriteString(" CONSTRAINT ")
	buf.WriteString(indexName)
	buf.WriteString(" UNIQUE(")
	for _, col := range cols {
		b.Quote(buf, col.Name)
		buf.WriteByte(',')
	}
	buf.Truncate(buf.Len() - 1) // 去掉最后一个逗号

	buf.WriteByte(')')
}

// create table语句中fk的约束部分的语句
func createFKSQL(b base, buf *bytes.Buffer, fk *forward.ForeignKey, fkName string) {
	//CONSTRAINT fk_name FOREIGN KEY (id) REFERENCES user(id)
	buf.WriteString(" CONSTRAINT ")
	buf.WriteString(fkName)

	buf.WriteString(" FOREIGN KEY(")
	b.Quote(buf, fk.Col.Name)

	buf.WriteString(") REFERENCES ")
	buf.WriteString(fk.RefTableName)

	buf.WriteByte('(')
	b.Quote(buf, fk.RefColName)
	buf.WriteByte(')')

	if len(fk.UpdateRule) > 0 {
		buf.WriteString(" ON UPDATE ")
		buf.WriteString(fk.UpdateRule)
	}

	if len(fk.DeleteRule) > 0 {
		buf.WriteString(" ON DELETE ")
		buf.WriteString(fk.DeleteRule)
	}
}

// create table语句中check约束部分的语句
func createCheckSQL(b base, buf *bytes.Buffer, expr, chkName string) {
	//CONSTRAINT chk_name CHECK (id>0 AND username='admin')
	buf.WriteString(" CONSTRAINT ")
	buf.WriteString(chkName)
	buf.WriteString(" CHECK(")
	buf.WriteString(expr)
	buf.WriteByte(')')
}

// 创建标准的几种约束(除PK约束，该约束有专门的函数createPKSQL()产生)：unique, foreign key, check
func createConstraints(b base, buf *bytes.Buffer, model *forward.Model) {
	// Unique Index
	for name, index := range model.UniqueIndexes {
		createUniqueSQL(b, buf, index, name)
		buf.WriteByte(',')
	}

	// foreign  key
	for name, fk := range model.FK {
		createFKSQL(b, buf, fk, name)
		buf.WriteByte(',')
	}

	// Check
	for name, chk := range model.Check {
		createCheckSQL(b, buf, chk, name)
		buf.WriteByte(',')
	}
}

// mysq系列数据库分页语法的实现。支持以下数据库：
// MySQL, H2, HSQLDB, Postgres, SQLite3
func mysqlLimitSQL(w *bytes.Buffer, limit int, offset ...int) ([]int, error) {
	if _, err := w.WriteString(" LIMIT ? "); err != nil {
		return nil, err
	}

	if len(offset) == 0 {
		return []int{limit}, nil
	}

	if _, err := w.WriteString(" OFFSET ? "); err != nil {
		return nil, err
	}
	return []int{limit, offset[0]}, nil
}

// oracle系列数据库分页语法的实现。支持以下数据库：
// Derby, SQL Server 2012, Oracle 12c, the SQL 2008 standard
func oracleLimitSQL(w *bytes.Buffer, limit int, offset ...int) ([]int, error) {
	if len(offset) == 0 {
		w.WriteString(" FETCH NEXT ? ROWS ONLY ")
		return []int{limit}, nil
	}

	w.WriteString(" OFFSET ? ROWS FETCH NEXT ? ROWS ONLY ")
	return []int{offset[0], limit}, nil
}
