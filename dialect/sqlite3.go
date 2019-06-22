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

func (s *sqlite3) LastInsertIDSQL(table, col string) (sql string, append bool) {
	return "", false
}

func (s *sqlite3) VersionSQL() string {
	return `select sqlite_version();`
}

func (s *sqlite3) CreateColumnSQL(buf *sqlbuilder.SQLBuilder, col *sqlbuilder.Column, isAI bool) error {
	buf.WriteByte('{').WriteString(col.Name).WriteByte('}')
	buf.WriteByte(' ')

	buf.WriteString(col.Type)
	if isAI {
		buf.WriteString(" PRIMARY KEY AUTOINCREMENT ")
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

func (s *sqlite3) CreateTableOptionsSQL(w *sqlbuilder.SQLBuilder, options map[string][]string) error {
	if len(options[sqlite3RowID]) == 1 {
		val, err := strconv.ParseBool(options[sqlite3RowID][0])
		if err != nil {
			return err
		}

		if !val {
			w.WriteString("WITHOUT ROWID")
		}
	} else if len(options[sqlite3RowID]) > 0 {
		return errors.New("rowid 只接受一个参数")
	}

	return nil
}

func (s *sqlite3) LimitSQL(limit interface{}, offset ...interface{}) (string, []interface{}) {
	return mysqlLimitSQL(limit, offset...)
}

func (s *sqlite3) DropIndexSQL(table, index string) (string, []interface{}) {
	return "DROP INDEX IF EXISTS {" + index + "}", nil
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
func (s *sqlite3) SQLType(col *orm.Column) (string, error) {
	if col == nil {
		return "", errColIsNil
	}

	if col.GoType == nil {
		return "", errGoTypeIsNil
	}

	switch col.GoType.Kind() {
	case reflect.Bool:
		return "INTEGER", nil
	case reflect.String:
		return "TEXT", nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "INTEGER", nil
	case reflect.Float32, reflect.Float64:
		return "REAL", nil
	case reflect.Array, reflect.Slice:
		if col.GoType.Elem().Kind() == reflect.Uint8 {
			return "BLOB", nil
		}
	case reflect.Struct:
		switch col.GoType {
		case rawBytes:
			return "BLOB", nil
		case nullBool:
			return "INTEGER", nil
		case nullFloat64:
			return "REAL", nil
		case nullInt64:
			return "INTEGER", nil
		case nullString:
			return "TEXT", nil
		case timeType:
			return "DATETIME", nil
		}
	}

	return "", errUncovert(col.GoType.Name())
}
