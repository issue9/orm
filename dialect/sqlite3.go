// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/issue9/orm/v2/core"
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
	_ sqlbuilder.CreateViewStmtHooker     = &sqlite3{}
)

// Sqlite3 返回一个适配 sqlite3 的 Dialect 接口
//
// Meta 可以接受以下参数：
//  rowid 可以是 rowid(false);rowid(true),rowid，其中只有 rowid(false) 等同于 without rowid
func Sqlite3() core.Dialect {
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
	query = ReplaceNamedArgs(query, args)
	return s.replacer.Replace(query), args, nil
}

func (s *sqlite3) LastInsertIDSQL(table, col string) (sql string, append bool) {
	return "", false
}

func (s *sqlite3) VersionSQL() string {
	return `select sqlite_version();`
}

func (s *sqlite3) Prepare(query string) (string, map[string]int, error) {
	query, orders, err := PrepareNamedArgs(query)
	if err != nil {
		return "", nil, err
	}
	return s.replacer.Replace(query), orders, nil
}

func (s *sqlite3) CreateTableOptionsSQL(w *core.Builder, options map[string][]string) error {
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
	return MysqlLimitSQL(limit, offset...)
}

// https://www.sqlite.org/lang_altertable.html
// BUG: 可能会让视图失去关联
func (s *sqlite3) AddConstraintStmtHook(stmt *sqlbuilder.AddConstraintStmt) ([]string, error) {
	builder := core.NewBuilder("CONSTRAINT ").
		WriteString(stmt.Name)
	switch stmt.Type {
	case core.ConstraintUnique:
		builder.WriteString(" UNIQUE(")
		for _, col := range stmt.Data {
			builder.WriteString(col).WriteBytes(',')
		}
		builder.TruncateLast(1).
			WriteBytes(')')
	case core.ConstraintPK:
		builder.WriteString(" PRIMARY KEY(")
		for _, col := range stmt.Data {
			builder.WriteString(col).WriteBytes(',')
		}
		builder.TruncateLast(1).
			WriteBytes(')')
	case core.ConstraintCheck:
		builder.WriteString(" CHECK(").
			WriteString(stmt.Data[0]).
			WriteBytes(')')
	case core.ConstraintFK:
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
		return nil, fmt.Errorf("未知的约束类型：%d", stmt.Type)
	}

	info, err := s3.ParseCreateTable(stmt.TableName, stmt.Engine())
	if err != nil {
		return nil, err
	}

	if _, found := info.Constraints[stmt.Name]; found {
		return nil, fmt.Errorf("已经存在相同的约束名：%s", stmt.Name)
	}

	query, err := builder.String()
	if err != nil {
		return nil, err
	}
	info.Constraints[stmt.Name] = &s3.Constraint{
		Type: stmt.Type,
		SQL:  query,
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
		return nil, fmt.Errorf("不存在的约束:%s", stmt.Name)
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

func (s *sqlite3) buildSQLS(e core.Engine, table *s3.Table, tableName string) ([]string, error) {
	ret := make([]string, 0, len(table.Indexes)+1)
	tmpName := "temp_" + tableName + "_temp"

	query, err := table.CreateTableSQL(tmpName)
	if err != nil {
		return nil, err
	}
	ret = append(ret, query)

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
	builder := core.NewBuilder("DELETE FROM ").
		QuoteKey(stmt.TableName)

	query, err := builder.String()
	if err != nil {
		return nil, err
	}

	if stmt.AIColumnName == "" {
		return []string{query}, nil
	}

	ret := make([]string, 2)
	ret[0] = query

	ret[1], err = builder.Reset().WriteString("DELETE FROM SQLITE_SEQUENCE WHERE name=").
		Quote(stmt.TableName, '\'', '\'').
		String()
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (s *sqlite3) CreateViewStmtHook(stmt *sqlbuilder.CreateViewStmt) ([]string, error) {
	ret := make([]string, 0, 2)
	if stmt.IsReplace {
		query, err := sqlbuilder.DropView(stmt.Engine()).Name(stmt.ViewName).DDLSQL()
		if err != nil {
			return nil, err
		}
		ret = append(ret, query...)
	}

	builder := core.NewBuilder("CREATE ")

	if stmt.IsTemporary {
		builder.WriteString(" TEMPORARY ")
	}

	builder.WriteString(" VIEW ").QuoteKey(stmt.ViewName)

	if len(stmt.Columns) > 0 {
		builder.WriteBytes('(')
		for _, col := range stmt.Columns {
			builder.QuoteKey(col).
				WriteBytes(',')
		}
		builder.TruncateLast(1).WriteBytes(')')
	}

	query, err := builder.WriteString(" AS ").
		WriteString(stmt.SelectQuery).
		String()
	if err != nil {
		return nil, err
	}
	ret = append(ret, query)

	return ret, nil
}

func (s *sqlite3) TransactionalDDL() bool {
	return true
}

// 具体规则参照:http://www.sqlite.org/datatype3.html
func (s *sqlite3) SQLType(col *core.Column) (string, error) {
	if col == nil {
		return "", errColIsNil
	}

	if col.GoType == nil {
		return "", errGoTypeIsNil
	}

	switch col.GoType.Kind() {
	case reflect.Bool:
		return s.buildType("INTEGER", col)
	case reflect.String:
		return s.buildType("TEXT", col)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return s.buildType("INTEGER", col)
	case reflect.Float32, reflect.Float64:
		return s.buildType("REAL", col)
	case reflect.Array, reflect.Slice:
		if col.GoType.Elem().Kind() == reflect.Uint8 {
			return s.buildType("BLOB", col)
		}
	case reflect.Struct:
		switch col.GoType {
		case core.RawBytesType:
			return s.buildType("BLOB", col)
		case core.NullBoolType:
			return s.buildType("INTEGER", col)
		case core.NullFloat64Type:
			return s.buildType("REAL", col)
		case core.NullInt64Type:
			return s.buildType("INTEGER", col)
		case core.NullStringType:
			return s.buildType("TEXT", col)
		case core.TimeType:
			return s.buildType("TIMESTAMP", col)
		}
	}

	return "", errUncovert(col.GoType.Name())
}

// l 表示需要取的长度数量
func (s *sqlite3) buildType(typ string, col *core.Column) (string, error) {
	w := core.NewBuilder(typ)

	if col.AI {
		w.WriteString(" PRIMARY KEY AUTOINCREMENT ")
	}

	if !col.Nullable {
		w.WriteString(" NOT NULL")
	}

	if col.HasDefault {
		v, err := s.SQLFormat(col.Default)
		if err != nil {
			return "", err
		}

		w.WriteString(" DEFAULT ").WriteString(v)
	}

	return w.String()
}

func (s *sqlite3) SQLFormat(v interface{}, length ...int) (f string, err error) {
	if vv, ok := v.(driver.Valuer); ok {
		v, err = vv.Value()
		if err != nil {
			return "", err
		}
	}

	if v == nil {
		return "NULL", nil
	}

	switch vv := v.(type) {
	case string:
		return "'" + vv + "'", nil
	case time.Time: // timestamp
		return "'" + vv.In(time.UTC).Format("2006-01-02 15:04:05") + "'", nil
	}

	return fmt.Sprint(v), nil
}
