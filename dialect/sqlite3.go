// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	s3 "github.com/issue9/orm/v2/internal/sqlite3"
	"github.com/issue9/orm/v2/sqlbuilder"
)

const (
	sqlite3Name  = "sqlite3"
	sqlite3RowID = sqlite3Name + "_rowid"
)

var sqlite3Inst *sqlite3

type sqlite3 struct {
	replacer *strings.Replacer
}

var (
	_ sqlbuilder.TruncateTableStmtHooker  = &sqlite3{}
	_ sqlbuilder.DropColumnStmtHooker     = &sqlite3{}
	_ sqlbuilder.DropConstraintStmtHooker = &sqlite3{}
	_ sqlbuilder.AddConstraintStmtHooker  = &sqlite3{}
)

// Sqlite3 返回一个适配 sqlite3 的 Dialect 接口
//
// Meta 可以接受以下参数：
//  rowid 可以是 rowid(false);rowid(true),rowid，其中只有 rowid(false) 等同于 without rowid
func Sqlite3() sqlbuilder.Dialect {
	if sqlite3Inst == nil {
		sqlite3Inst = &sqlite3{
			replacer: strings.NewReplacer("{", "`", "}", "`"),
		}
	}

	return sqlite3Inst
}

func (s *sqlite3) Name() string {
	return sqlite3Name
}

func (s *sqlite3) SQL(query string, args []interface{}) (string, []interface{}, error) {
	query = replaceNamedArgs(query, args)
	return s.replacer.Replace(query), args, nil
}

func (s *sqlite3) LastInsertIDSQL(table, col string) (sql string, append bool) {
	return "", false
}

func (s *sqlite3) VersionSQL() string {
	return `select sqlite_version();`
}

func (s *sqlite3) Prepare(query string) (string, map[string]int) {
	query, orders := PrepareNamedArgs(query)
	return s.replacer.Replace(query), orders
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

// https://www.sqlite.org/lang_altertable.html
// BUG: 可能会让视图失去关联
func (s *sqlite3) AddConstraintStmtHook(stmt *sqlbuilder.AddConstraintStmt) ([]string, error) {
	builder := sqlbuilder.New("CONSTRAINT ").
		WriteString(stmt.Name)
	switch stmt.Type {
	case sqlbuilder.ConstraintUnique:
		builder.WriteString(" UNIQUE(")
		for _, col := range stmt.Data {
			builder.WriteString(col).WriteBytes(',')
		}
		builder.TruncateLast(1).
			WriteBytes(')')
	case sqlbuilder.ConstraintPK:
		builder.WriteString(" PRIMARY KEY(")
		for _, col := range stmt.Data {
			builder.WriteString(col).WriteBytes(',')
		}
		builder.TruncateLast(1).
			WriteBytes(')')
	case sqlbuilder.ConstraintCheck:
		builder.WriteString(" CHECK(").
			WriteString(stmt.Data[0]).
			WriteBytes(')')
	case sqlbuilder.ConstraintFK:
		builder.WriteString(" FOREIGN KEY(").
			WriteString(stmt.Data[0]).
			WriteString(") REFERENCES ").
			WriteString(stmt.Data[1]).
			WriteBytes('(').
			WriteString(stmt.Data[2]).
			WriteBytes(')')
		if len(stmt.Data) >= 3 && stmt.Data[3] != "" {
			builder.WriteString(" ON UPDATE ").WriteString(stmt.Data[3])
		}
		if len(stmt.Data) >= 4 && stmt.Data[4] != "" {
			builder.WriteString(" ON DELETE ").WriteString(stmt.Data[4])
		}
	default:
		return nil, fmt.Errorf("未知的约束类型：%s", stmt.Type)
	}

	info, err := s3.ParseCreateTable(stmt.TableName, stmt.Engine())
	if err != nil {
		return nil, err
	}

	if _, found := info.Constraints[stmt.Name]; found {
		return nil, fmt.Errorf("已经存在相同的约束名：%s", stmt.Name)
	}

	info.Constraints[stmt.Name] = &s3.Constraint{
		Type: stmt.Type,
		SQL:  builder.String(),
	}

	return s.buildSQLS(stmt.Engine(), info, stmt.TableName)
}

// https://www.sqlite.org/lang_altertable.html
// BUG: 可能会让视图失去关联
func (s *sqlite3) DropConstraintStmtHook(stmt *sqlbuilder.DropConstraintStmt) ([]string, error) {
	info, err := s3.ParseCreateTable(stmt.TableName, stmt.Engine())
	if err != nil {
		return nil, err
	}

	if _, found := info.Constraints[stmt.Name]; !found {
		return nil, fmt.Errorf("约束 %s 不存在", stmt.Name)
	}

	delete(info.Constraints, stmt.Name)

	return s.buildSQLS(stmt.Engine(), info, stmt.TableName)
}

// https://www.sqlite.org/lang_altertable.html
// BUG: 可能会让视图失去关联
func (s *sqlite3) DropColumnStmtHook(stmt *sqlbuilder.DropColumnStmt) ([]string, error) {
	info, err := s3.ParseCreateTable(stmt.TableName, stmt.Engine())
	if err != nil {
		return nil, err
	}

	if _, found := info.Columns[stmt.ColumnName]; !found {
		return nil, fmt.Errorf("列 %s 不存在", stmt.ColumnName)
	}

	delete(info.Columns, stmt.ColumnName)

	return s.buildSQLS(stmt.Engine(), info, stmt.TableName)
}

func (s *sqlite3) buildSQLS(e sqlbuilder.Engine, table *s3.Table, tableName string) ([]string, error) {
	ret := make([]string, 0, len(table.Indexes)+1)

	tmpName := "temp_" + tableName + "_temp"
	ret = append(ret, table.CreateTableSQL(tmpName))

	sel := sqlbuilder.Select(e).From(tableName)
	for col := range table.Columns {
		sel.Column(col)
	}

	query, args, err := sel.Insert().Table(tmpName).SQL()
	if err != nil {
		return nil, err
	}
	if len(args) > 0 {
		panic("复制表时，SELECT 不应该有参数")
	}
	ret = append(ret, query)

	// 删除旧表
	ret = append(ret, "DROP TABLE "+tableName)

	// 重命名新表名称
	ret = append(ret, fmt.Sprintf("ALTER TABLE %s RENAME TO %s", tmpName, tableName))

	// 在新表生成之后，重新创建索引
	for _, index := range table.Indexes {
		ret = append(ret, index.SQL)
	}

	return ret, nil
}

func (s *sqlite3) TruncateTableStmtHook(stmt *sqlbuilder.TruncateTableStmt) ([]string, error) {
	builder := sqlbuilder.New("DELETE FROM ").
		QuoteKey(stmt.TableName)

	if stmt.AIColumnName == "" {
		return []string{builder.String()}, nil
	}

	ret := make([]string, 2)
	ret[0] = builder.String()
	builder.Reset()
	ret[1] = builder.WriteString("DELETE FROM SQLITE_SEQUENCE WHERE name='").
		WriteString(stmt.TableName).
		WriteBytes('\'').
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
		case sqlbuilder.RawBytesType:
			return buildSqlite3Type("BLOB", col), nil
		case sqlbuilder.NullBoolType:
			return buildSqlite3Type("INTEGER", col), nil
		case sqlbuilder.NullFloat64Type:
			return buildSqlite3Type("REAL", col), nil
		case sqlbuilder.NullInt64Type:
			return buildSqlite3Type("INTEGER", col), nil
		case sqlbuilder.NullStringType:
			return buildSqlite3Type("TEXT", col), nil
		case sqlbuilder.TimeType:
			return buildSqlite3Type("TIMESTAMP", col), nil
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
			WriteBytes('\'')
	}

	return w.String()
}
