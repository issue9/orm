// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"bytes"
	"database/sql"
	"reflect"
	"time"

	"github.com/issue9/orm/core"
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

// 对core.Dialect接口的扩展，包含一些包内通用的接口。
type base interface {
	core.Dialect

	// 将col转换成sql类型，并写入buf中。
	sqlType(buf *bytes.Buffer, col *core.Column) error
}

// 用于产生在createTable中使用的普通列信息表达式，不包含autoincrement和primary key的关键字。
func createColSQL(b base, buf *bytes.Buffer, col *core.Column) error {
	// col_name VARCHAR(100) NOT NULL DEFAULT 'abc'
	buf.WriteString(col.Name)
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
func createPKSQL(b base, buf *bytes.Buffer, cols []*core.Column, pkName string) {
	//CONSTRAINT pk_name PRIMARY KEY (id,lastName)
	buf.WriteString(" CONSTRAINT ")
	buf.WriteString(pkName)
	buf.WriteString(" PRIMARY KEY(")
	for _, col := range cols {
		buf.WriteString(col.Name)
		buf.WriteByte(',')
	}
	buf.Truncate(buf.Len() - 1) // 去掉最后一个逗号

	buf.WriteByte(')')
}

// create table语句中的unique约束部分的语句。
func createUniqueSQL(b base, buf *bytes.Buffer, cols []*core.Column, indexName string) {
	//CONSTRAINT unique_name UNIQUE (id,lastName)
	buf.WriteString(" CONSTRAINT ")
	buf.WriteString(indexName)
	buf.WriteString(" UNIQUE(")
	for _, col := range cols {
		buf.WriteString(col.Name)
		buf.WriteByte(',')
	}
	buf.Truncate(buf.Len() - 1) // 去掉最后一个逗号

	buf.WriteByte(')')
}

// create table语句中fk的约束部分的语句
func createFKSQL(b base, buf *bytes.Buffer, fk *core.ForeignKey, fkName string) {
	//CONSTRAINT fk_name FOREIGN KEY (id) REFERENCES user(id)
	buf.WriteString(" CONSTRAINT ")
	buf.WriteString(fkName)

	buf.WriteString(" FOREIGN KEY(")
	buf.WriteString(fk.Col.Name)

	buf.WriteString(") REFERENCES ")
	buf.WriteString(fk.RefTableName)

	buf.WriteByte('(')
	buf.WriteString(fk.RefColName)
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

// mysq系列数据库分页语法的实现。支持以下数据库：
// MySQL, H2, HSQLDB, Postgres, SQLite3
func mysqlLimitSQL(limit int, offset ...int) (string, []interface{}) {
	if len(offset) == 0 {
		return " LIMIT ? ", []interface{}{limit}
	}

	return " LIMIT ? OFFSET ? ", []interface{}{limit, offset[0]}
}

// oracle系列数据库分页语法的实现。支持以下数据库：
// Derby, SQL Server 2012, Oracle 12c, the SQL 2008 standard
func oracleLimitSQL(limit int, offset ...int) (string, []interface{}) {
	if len(offset) == 0 {
		return " FETCH NEXT ? ROWS ONLY ", []interface{}{limit}
	}

	return " OFFSET ? ROWS FETCH NEXT ? ROWS ONLY ", []interface{}{offset[0], limit}
}