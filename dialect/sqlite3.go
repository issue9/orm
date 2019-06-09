// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"errors"
	"reflect"
	"strconv"

	"github.com/issue9/orm/v2"
	"github.com/issue9/orm/v2/sqlbuilder"
)

const (
	sqlite3Name  = "sqlite3"
	sqlite3RowID = sqlite3Name + "_rowid"
)

var sqlite3Inst *sqlite3

type sqlite3 struct{}

// Sqlite3 返回一个适配 sqlite3 的 orm.Dialect 接口
//
// Meta 可以接受以下参数：
//  rowid 可以是 rowid(false);rowid(true),rowid，其中只有 rowid(false) 等同于 without rowid
func Sqlite3() orm.Dialect {
	if sqlite3Inst == nil {
		sqlite3Inst = &sqlite3{}
	}

	return sqlite3Inst
}

func (s *sqlite3) Name() string {
	return sqlite3Name
}

func (s *sqlite3) QuoteTuple() (byte, byte) {
	return '`', '`'
}

func (s *sqlite3) SQL(sql string) (string, error) {
	return sql, nil
}

func (s *sqlite3) LastInsertID(table, col string) (sql string, first, append bool) {
	return "", false, false
}

func (s *sqlite3) VersionSQL() string {
	return `select sqlite_version();`
}

func (s *sqlite3) CreateTableSQL(model *orm.Model) ([]string, error) {
	w := sqlbuilder.New("CREATE TABLE IF NOT EXISTS ").
		WriteString("{#").
		WriteString(model.Name).
		WriteString("}(")

	// 自增列
	if model.AI != nil {
		if err := createColSQL(s, w, model.AI); err != nil {
			return nil, err
		}
		w.WriteString(" PRIMARY KEY AUTOINCREMENT,")
	}

	// 普通列
	for _, col := range model.Cols {
		if col.IsAI() { // 忽略 AI 列
			continue
		}

		if err := createColSQL(s, w, col); err != nil {
			return nil, err
		}
		w.WriteByte(',')
	}

	// 约束
	if len(model.PK) > 0 && !model.PK[0].IsAI() { // PK，若有自增，则已经在上面指定
		createPKSQL(w, model.PK, pkName)
		w.WriteByte(',')
	}
	createConstraints(w, model)
	w.TruncateLast(1).WriteByte(')')

	if err := s.createTableOptions(w, model); err != nil {
		return nil, err
	}

	indexs, err := createIndexSQL(model)
	if err != nil {
		return nil, err
	}
	return append([]string{w.String()}, indexs...), nil
}

func (s *sqlite3) createTableOptions(w *sqlbuilder.SQLBuilder, model *orm.Model) error {
	if len(model.Meta[sqlite3RowID]) == 1 {
		val, err := strconv.ParseBool(model.Meta[sqlite3RowID][0])
		if err != nil {
			return err
		}

		if !val {
			w.WriteString("WITHOUT ROWID")
		}
	} else if len(model.Meta[sqlite3RowID]) > 0 {
		return errors.New("rowid 只接受一个参数")
	}

	return nil
}

func (s *sqlite3) LimitSQL(limit interface{}, offset ...interface{}) (string, []interface{}) {
	return mysqlLimitSQL(limit, offset...)
}

func (s *sqlite3) TruncateTableSQL(m *orm.Model) []string {
	ret := make([]string, 2)
	ret[0] = sqlbuilder.New("DELETE FROM #").
		WriteString(m.Name).
		String()
	ret[1] = sqlbuilder.New("DELETE FROM SQLITE_SEQUENCE WHERE name='#").
		WriteString(m.Name).
		WriteByte('\'').
		String()

	return ret
}

func (s *sqlite3) TransactionalDDL() bool {
	return true
}

// 具体规则参照:http://www.sqlite.org/datatype3.html
func (s *sqlite3) sqlType(buf *sqlbuilder.SQLBuilder, col *orm.Column) error {
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
		if col.GoType.Elem().Kind() == reflect.Uint8 {
			buf.WriteString("BLOB")
		}
	case reflect.Struct:
		switch col.GoType {
		case rawBytes:
			buf.WriteString("BLOB")
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
