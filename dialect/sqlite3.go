// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"errors"
	"reflect"
	"strconv"

	"github.com/issue9/orm/model"
	"github.com/issue9/orm/sqlbuilder"
	"github.com/issue9/orm/types"
)

// Sqlite3 返回一个适配 sqlite3 的 types.Dialect 接口
//
// Meta 可以接受以下参数：
//  rowid 可以是 rowid(false);rowid(true),rowid，其中只有 rowid(false) 等同于 without rowid
func Sqlite3() types.Dialect {
	return &sqlite3{}
}

type sqlite3 struct{}

func (s *sqlite3) Name() string {
	return "sqlite3"
}

func (s *sqlite3) QuoteTuple() (byte, byte) {
	return '`', '`'
}

func (s *sqlite3) SQL(sql string) (string, error) {
	return sql, nil
}

func (s *sqlite3) CreateTableSQL(model *model.Model) (string, error) {
	w := sqlbuilder.New("CREATE TABLE IF NOT EXISTS ").
		WriteString("{#").
		WriteString(model.Name).
		WriteString("}(")

	// 自增列
	if model.AI != nil {
		if err := createColSQL(s, w, model.AI); err != nil {
			return "", err
		}
		w.WriteString(" PRIMARY KEY AUTOINCREMENT,")
	}

	// 普通列
	for _, col := range model.Cols {
		if col.IsAI() { // 忽略 AI 列
			continue
		}

		if err := createColSQL(s, w, col); err != nil {
			return "", err
		}
		w.WriteByte(',')
	}

	// 约束
	if len(model.PK) > 0 && !model.PK[0].IsAI() { // PK，若有自增，则已经在上面指定
		createPKSQL(s, w, model.PK, pkName)
		w.WriteByte(',')
	}
	createConstraints(s, w, model)
	w.TruncateLast(1).WriteByte(')')

	if err := s.createTableOptions(w, model); err != nil {
		return "", err
	}
	return w.String(), nil
}

func (s *sqlite3) createTableOptions(w *sqlbuilder.SQLBuilder, model *model.Model) error {
	if len(model.Meta["rowid"]) == 1 {
		val, err := strconv.ParseBool(model.Meta["rowid"][0])
		if err != nil {
			return err
		}

		if !val {
			w.WriteString("WITHOUT ROWID")
		}
	} else if len(model.Meta["rowid"]) > 0 {
		return errors.New("rowid 只接受一个参数")
	}

	return nil
}

func (s *sqlite3) LimitSQL(limit int, offset ...int) (string, []interface{}) {
	return mysqlLimitSQL(limit, offset...)
}

func (s *sqlite3) TruncateTableSQL(model *model.Model) string {
	return sqlbuilder.New("DELETE FROM ").
		WriteString("#" + model.Name).
		WriteString(";update sqlite_sequence set seq=0 where name='").
		WriteString("#" + model.Name).
		WriteString("';").
		String()
}

// 具体规则参照:http://www.sqlite.org/datatype3.html
func (s *sqlite3) sqlType(buf *sqlbuilder.SQLBuilder, col *model.Column) error {
	if col == nil {
		return errors.New("sqlType:col参数是个空值")
	}

	if col.GoType == nil {
		return errors.New("sqlType:无效的col.GoType值")
	}

	switch col.GoType.Kind() {
	case reflect.Bool:
		buf.WriteString("INTEGER")
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
		}
		buf.WriteString("TEXT")
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
