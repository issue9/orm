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
func (s *Sqlite3) QuoteTuple() (byte, byte) {
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

// implement orm.Dialect.AIColSQL()
func (s *Sqlite3) AIColSQL(w *bytes.Buffer, model *orm.Model) error {
	if model.AI == nil {
		return nil
	}

	if err := createColSQL(s, w, model.AI); err != nil {
		return err
	}

	_, err := w.WriteString(" PRIMARY KEY AUTOINCREMENT,")
	return err
}

// implement orm.Dialect.NoAIColSQL()
func (s *Sqlite3) NoAIColSQL(w *bytes.Buffer, model *orm.Model) error {
	for _, col := range model.Cols {
		if col.IsAI() { // 忽略AI列
			continue
		}

		if err := createColSQL(s, w, col); err != nil {
			return err
		}
		w.WriteByte(',')
	}
	return nil
}

// implement orm.Dialect.ConstraintsSQL()
func (s *Sqlite3) ConstraintsSQL(w *bytes.Buffer, m *orm.Model) error {
	// PK，若有自增，则已经在上面指定
	if len(m.PK) > 0 && !m.PK[0].IsAI() {
		createPKSQL(s, w, m.PK, pkName)
		w.WriteByte(',')
	}

	createConstraints(s, w, m)
	return nil
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
