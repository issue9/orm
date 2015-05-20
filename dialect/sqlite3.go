// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"bytes"
	"errors"
	"reflect"

	"github.com/issue9/orm"
)

type Sqlite3 struct{}

// implement orm.Dialect.QuoteTuple()
func (m *Sqlite3) QuoteTuple() (byte, byte) {
	return '`', '`'
}

// implement orm.Dialect.Quote()
func (s *Sqlite3) Quote(w *bytes.Buffer, name string) error {
	if err := w.WriteByte('`'); err != nil {
		return err
	}

	if _, err := w.WriteString(name); err != nil {
		return err
	}

	return w.WriteByte('`')
}

// implement orm.Dialect.LimitSQL()
func (s *Sqlite3) LimitSQL(w *bytes.Buffer, limit int, offset ...int) ([]int, error) {
	return mysqlLimitSQL(w, limit, offset...)
}

// implement orm.Dialect.CreateTableSQL()
func (s *Sqlite3) CreateTableSQL(model *orm.Model) (string, error) {
	buf := bytes.NewBufferString("CREATE TABLE IF NOT EXISTS ")
	buf.Grow(300)

	buf.WriteString(model.Name)
	buf.WriteByte('(')

	// 写入字段信息
	for _, col := range model.Cols {
		if err := createColSQL(s, buf, col); err != nil {
			return "", err
		}

		if col.IsAI() {
			buf.WriteString(" PRIMARY KEY AUTOINCREMENT")
		}
		buf.WriteByte(',')
	}

	// PK，若有自增，则已经在上面指定
	if len(model.PK) > 0 && !model.PK[0].IsAI() {
		createPKSQL(s, buf, model.PK, pkName)
		buf.WriteByte(',')
	}

	createConstraints(s, buf, model)

	buf.Truncate(buf.Len() - 1) // 去掉最后的逗号
	buf.WriteByte(')')          // end CreateTable

	return buf.String(), nil
}

// implement orm.Dialect.TruncateTableSQL()
func (s *Sqlite3) TruncateTableSQL(tableName string) string {
	return "DELETE FROM " + tableName +
		";update sqlite_sequence set seq=0 where name='" + tableName + "';"
}

// implement base.sqlType()
// 具体规则参照:http://www.sqlite.org/datatype3.html
func (s *Sqlite3) sqlType(buf *bytes.Buffer, col *orm.Column) error {
	if col == nil {
		return errors.New("sqlType:col参数是个空值")
	}

	if col.GoType == nil {
		return errors.New("sqlType:无效的col.GoType值")
	}

	switch col.GoType.Kind() {
	case reflect.String:
		buf.WriteString("TEXT")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		buf.WriteString("INTEGER")
	case reflect.Float32, reflect.Float64:
		buf.WriteString("REAL")
	case reflect.Array, reflect.Slice:
		k := col.GoType.Elem().Kind()
		if (k != reflect.Uint8) && (k != reflect.Int32) {
			return errors.New("sqlType:不支持数组类型")
		} else {
			buf.WriteString("TEXT")
		}
	case reflect.Struct:
		switch col.GoType {
		case nullBool:
			buf.WriteString("INTEGER")
		case nullFloat64:
			buf.WriteString("REAL")
		case nullInt64:
			buf.WriteString("INTEGER")
		case nullString:
			buf.WriteString("TEXT")
		case timeType:
			buf.WriteString("DATETIME")
		}
	}

	return nil
}
