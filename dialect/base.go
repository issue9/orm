// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"bytes"

	"github.com/issue9/orm/core"
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

// 添加标准的索引约束：pk,unique,foreign key,check
// 一些非标准的索引需要各个Dialect自己去实现：如mysql的KEY索引
func addIndexes(b base, db core.DB, model *core.Model) error {
	// ALTER TABLE语句的公共语句部分，可以重复利用：
	// ALTER TABLE table_name ADD CONSTRAINT
	buf := bytes.NewBufferString("ALTER TABLE ")
	buf.WriteString(model.Name)
	buf.WriteString(" ADD CONSTRAINT ")
	size := buf.Len()

	// ALTER TABLE tbname ADD CONSTRAINT pk PRIMARY KEY
	buf.WriteString("pk PRIMARY KEY(")
	for _, col := range model.PK {
		buf.WriteString(col.Name)
		buf.WriteByte(',')
	}
	buf.UnreadByte()
	buf.WriteByte(')')
	if _, err := db.Exec(buf.String()); err != nil {
		return err
	}

	// ALTER TABLE tbname ADD CONSTRAINT uniquteName unique(...)
	for name, cols := range model.UniqueIndexes {
		buf.Truncate(size)
		buf.WriteString(name)
		buf.WriteString(" UNIQUE(")
		for _, col := range cols {
			buf.WriteString(col.Name)
			buf.WriteByte(',')
		}
		buf.UnreadByte()
		buf.WriteByte(')')

		if _, err := db.Exec(buf.String()); err != nil {
			return err
		}
	}

	// fk ALTER TABLE tbname ADD CONSTRAINT fkname FOREIGN KEY (col) REFERENCES tbl(tblcol)
	for name, fk := range model.FK {
		buf.Truncate(size)
		buf.WriteString(name)
		buf.WriteString(" FOREIGN KEY(")
		buf.WriteString(fk.Col.Name)
		buf.WriteByte(')')

		buf.WriteString(" REFERENCES ")
		buf.WriteString(fk.RefTableName)
		buf.WriteByte('(')
		buf.WriteString(fk.RefColName)
		buf.WriteByte(')')

		if _, err := db.Exec(buf.String()); err != nil {
			return err
		}
	}

	// chk ALTER TABLE tblname ADD CONSTRAINT chkName CHECK (id>0 AND city='abc')
	for name, expr := range model.Check {
		buf.Truncate(size)
		buf.WriteString(name) // checkName
		buf.WriteString(" CHECK(")
		buf.WriteString(expr)
		buf.WriteByte(')')

		if _, err := db.Exec(buf.String()); err != nil {
			return err
		}
	}

	return nil
}
