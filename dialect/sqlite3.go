// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"errors"
	"fmt"
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

func (s *sqlite3) AddConstraintStmtHook(stmt *sqlbuilder.AddConstraintStmt) ([]string, error) {
	// TODO
	return nil, sqlbuilder.ErrNotImplemented
}

func (s *sqlite3) DropConstraintStmtHook(stmt *sqlbuilder.DropConstraintStmt) ([]string, error) {
	// TODO
	return nil, sqlbuilder.ErrNotImplemented
}

func (s *sqlite3) DropColumnStmtHook(stmt *sqlbuilder.DropColumnStmt) ([]string, error) {
	// TODO https://www.sqlite.org/lang_altertable.html
	return nil, sqlbuilder.ErrNotImplemented
}

func (s *sqlite3) TruncateTableStmtHook(stmt *sqlbuilder.TruncateTableStmt) ([]string, error) {
	builder := sqlbuilder.New("DELETE FROM ").
		WriteString(stmt.TableName)
	if stmt.AIColumnName == "" {
		return []string{builder.String()}, nil
	}

	// 获取表名，以下表名仅用为字符串使用，需要去掉 {} 两个符号
	tablename := stmt.TableName
	if tablename[0] == '{' {
		tablename = tablename[1 : len(tablename)-1]
	}

	ret := make([]string, 2)
	ret[0] = builder.String()
	builder.Reset()
	ret[1] = builder.WriteString("DELETE FROM SQLITE_SEQUENCE WHERE name='").
		WriteString(tablename).
		WriteByte('\'').
		String()

	return ret, nil
}

func (s *sqlite3) TransactionalDDL() bool {
	return true
}

// 具体规则参照:http://www.sqlite.org/datatype3.html
func (s *sqlite3) SQLType(col *sqlbuilder.Column) (string, error) {
	if col == nil {
		return "", errColIsNil
	}

	if col.GoType == nil {
		return "", errGoTypeIsNil
	}

	switch col.GoType.Kind() {
	case reflect.Bool:
		return buildSqlite3Type("INTEGER", col), nil
	case reflect.String:
		return buildSqlite3Type("TEXT", col), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return buildSqlite3Type("INTEGER", col), nil
	case reflect.Float32, reflect.Float64:
		return buildSqlite3Type("REAL", col), nil
	case reflect.Array, reflect.Slice:
		if col.GoType.Elem().Kind() == reflect.Uint8 {
			return buildSqlite3Type("BLOB", col), nil
		}
	case reflect.Struct:
		switch col.GoType {
		case rawBytes:
			return buildSqlite3Type("BLOB", col), nil
		case nullBool:
			return buildSqlite3Type("INTEGER", col), nil
		case nullFloat64:
			return buildSqlite3Type("REAL", col), nil
		case nullInt64:
			return buildSqlite3Type("INTEGER", col), nil
		case nullString:
			return buildSqlite3Type("TEXT", col), nil
		case timeType:
			return buildSqlite3Type("DATETIME", col), nil
		}
	}

	return "", errUncovert(col.GoType.Name())
}

// l 表示需要取的长度数量
func buildSqlite3Type(typ string, col *sqlbuilder.Column) string {
	w := sqlbuilder.New(typ)

	if col.AI {
		w.WriteString(" PRIMARY KEY AUTOINCREMENT ")
	}

	if !col.Nullable {
		w.WriteString(" NOT NULL")
	}

	if col.HasDefault {
		w.WriteString(" DEFAULT '").
			WriteString(fmt.Sprint(col.Default)).
			WriteByte('\'')
	}

	return w.String()
}
