// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"bytes"
	"errors"
	"reflect"

	"github.com/issue9/orm/forward"
)

// 返回一个适配sqlite3的forward.Dialect接口
func Sqlite3() forward.Dialect {
	return &sqlite3{}
}

type sqlite3 struct{}

// implement forward.Dialect.SupportInsertMany()
func (s *sqlite3) SupportInsertMany() bool {
	return true
}

// implement forward.Dialect.QuoteTuple()
func (s *sqlite3) QuoteTuple() (byte, byte) {
	return '`', '`'
}

// implement forward.Dialect.Quote()
func (s *sqlite3) Quote(w *bytes.Buffer, name string) error {
	if err := w.WriteByte('`'); err != nil {
		return err
	}

	if _, err := w.WriteString(name); err != nil {
		return err
	}

	return w.WriteByte('`')
}

// implement forward.Dialect.ReplaceMarks()
func (s *sqlite3) ReplaceMarks(sql *string) error {
	return nil
}

// implement forward.Dialect.LimitSQL()
func (s *sqlite3) LimitSQL(w *bytes.Buffer, limit int, offset ...int) ([]int, error) {
	return mysqlLimitSQL(w, limit, offset...)
}

// implement forward.Dialect.AIColSQL()
func (s *sqlite3) AIColSQL(w *bytes.Buffer, model *forward.Model) error {
	if model.AI == nil {
		return nil
	}

	if err := createColSQL(s, w, model.AI); err != nil {
		return err
	}

	_, err := w.WriteString(" PRIMARY KEY AUTOINCREMENT,")
	return err
}

// implement forward.Dialect.NoAIColSQL()
func (s *sqlite3) NoAIColSQL(w *bytes.Buffer, model *forward.Model) error {
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

// implement forward.Dialect.ConstraintsSQL()
func (s *sqlite3) ConstraintsSQL(w *bytes.Buffer, m *forward.Model) error {
	// PK，若有自增，则已经在上面指定
	if len(m.PK) > 0 && !m.PK[0].IsAI() {
		createPKSQL(s, w, m.PK, pkName)
		w.WriteByte(',')
	}

	createConstraints(s, w, m)
	return nil
}

// implement forward.Dialect.TruncateTableSQL()
func (s *sqlite3) TruncateTableSQL(w *bytes.Buffer, tableName, aiColumn string) error {
	w.WriteString("DELETE FROM ")
	w.WriteString(tableName)
	w.WriteString(";update sqlite_sequence set seq=0 where name='")
	w.WriteString(tableName)
	_, err := w.WriteString("';")
	return err
}

// implement base.sqlType()
// 具体规则参照:http://www.sqlite.org/datatype3.html
func (s *sqlite3) sqlType(buf *bytes.Buffer, col *forward.Column) error {
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
