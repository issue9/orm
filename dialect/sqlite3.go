// SPDX-License-Identifier: MIT

package dialect

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/issue9/orm/v4/core"
	"github.com/issue9/orm/v4/internal/createtable"
	"github.com/issue9/orm/v4/sqlbuilder"
)

const sqlite3RowID = "sqlite3_rowid"

type sqlite3 struct {
	driverName string
	replacer   *strings.Replacer
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
func Sqlite3(driverName string) core.Dialect {
	return &sqlite3{
		driverName: driverName,
		replacer:   strings.NewReplacer("{", "`", "}", "`"),
	}
}

func (s *sqlite3) DBName() string {
	return "sqlite3"
}

func (s *sqlite3) DriverName() string {
	return s.driverName
}

func (s *sqlite3) Fix(query string, args []interface{}) (string, []interface{}, error) {
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
			w.WString("WITHOUT ROWID")
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
		WString(stmt.Name)
	switch stmt.Type {
	case core.ConstraintUnique:
		builder.WString(" UNIQUE(")
		for _, col := range stmt.Data {
			builder.WString(col).WBytes(',')
		}
		builder.TruncateLast(1).
			WBytes(')')
	case core.ConstraintPK:
		builder.WString(" PRIMARY KEY(")
		for _, col := range stmt.Data {
			builder.WString(col).WBytes(',')
		}
		builder.TruncateLast(1).
			WBytes(')')
	case core.ConstraintCheck:
		builder.WString(" CHECK(").
			WString(stmt.Data[0]).
			WBytes(')')
	case core.ConstraintFK:
		builder.WString(" FOREIGN KEY(").
			WString(stmt.Data[0]).
			WString(") REFERENCES ").
			WString(stmt.Data[1]).
			WBytes('(').
			WString(stmt.Data[2]).
			WBytes(')')
		if len(stmt.Data) >= 3 && stmt.Data[3] != "" {
			builder.WString(" ON UPDATE ").WString(stmt.Data[3])
		}
		if len(stmt.Data) >= 4 && stmt.Data[4] != "" {
			builder.WString(" ON DELETE ").WString(stmt.Data[4])
		}
	default:
		return nil, fmt.Errorf("未知的约束类型：%d", stmt.Type)
	}

	info, err := createtable.ParseSqlite3CreateTable(stmt.TableName, stmt.Engine())
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
	info.Constraints[stmt.Name] = &createtable.Sqlite3Constraint{
		Type: stmt.Type,
		SQL:  query,
	}

	return s.buildSQLS(stmt.Engine(), info, stmt.TableName)
}

// https://www.sqlite.org/lang_altertable.html
// BUG: 可能会让视图失去关联
func (s *sqlite3) DropConstraintStmtHook(stmt *sqlbuilder.DropConstraintStmt) ([]string, error) {
	info, err := createtable.ParseSqlite3CreateTable(stmt.TableName, stmt.Engine())
	if err != nil {
		return nil, err
	}

	name := strings.Replace(stmt.Name, "#", stmt.Engine().TablePrefix(), 1)
	if _, found := info.Constraints[name]; !found {
		return nil, fmt.Errorf("不存在的约束:%s", name)
	}

	delete(info.Constraints, name)

	return s.buildSQLS(stmt.Engine(), info, stmt.TableName)
}

// https://www.sqlite.org/lang_altertable.html
// BUG: 可能会让视图失去关联
func (s *sqlite3) DropColumnStmtHook(stmt *sqlbuilder.DropColumnStmt) ([]string, error) {
	info, err := createtable.ParseSqlite3CreateTable(stmt.TableName, stmt.Engine())
	if err != nil {
		return nil, err
	}

	if _, found := info.Columns[stmt.ColumnName]; !found {
		return nil, fmt.Errorf("列 %s 不存在", stmt.ColumnName)
	}

	delete(info.Columns, stmt.ColumnName)

	return s.buildSQLS(stmt.Engine(), info, stmt.TableName)
}

func (s *sqlite3) buildSQLS(e core.Engine, table *createtable.Sqlite3Table, tableName string) ([]string, error) {
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

	ret[1], err = builder.Reset().WString("DELETE FROM SQLITE_SEQUENCE WHERE name=").
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
		builder.WString(" TEMPORARY ")
	}

	builder.WString(" VIEW ").QuoteKey(stmt.ViewName)

	if len(stmt.Columns) > 0 {
		builder.WBytes('(')
		for _, col := range stmt.Columns {
			builder.QuoteKey(col).
				WBytes(',')
		}
		builder.TruncateLast(1).WBytes(')')
	}

	query, err := builder.WString(" AS ").
		WString(stmt.SelectQuery).
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

	switch col.PrimitiveType {
	case core.Bool:
		return s.buildType("INTEGER", col)
	case core.String:
		return s.buildType("TEXT", col)
	case core.Int, core.Int8, core.Int16, core.Int32, core.Int64,
		core.Uint, core.Uint8, core.Uint16, core.Uint32, core.Uint64:
		return s.buildType("INTEGER", col)
	case core.Float32, core.Float64:
		return s.buildType("REAL", col)
	case core.RawBytes, core.Bytes:
		return s.buildType("BLOB", col)
	case core.NullBool:
		return s.buildType("INTEGER", col)
	case core.NullFloat64:
		return s.buildType("REAL", col)
	case core.NullInt64:
		return s.buildType("INTEGER", col)
	case core.NullString:
		return s.buildType("TEXT", col)
	case core.Time, core.NullTime:
		return s.buildType("TIMESTAMP", col)
	}

	return "", errUncovert(col)
}

// l 表示需要取的长度数量
func (s *sqlite3) buildType(typ string, col *core.Column) (string, error) {
	w := core.NewBuilder(typ)

	if col.AI {
		w.WString(" PRIMARY KEY AUTOINCREMENT ")
	}

	if !col.Nullable {
		w.WString(" NOT NULL")
	}

	if col.HasDefault {
		v, err := s.formatSQL(col.Default)
		if err != nil {
			return "", err
		}

		w.WString(" DEFAULT ").WString(v)
	}

	return w.String()
}

func (s *sqlite3) formatSQL(v interface{}, length ...int) (f string, err error) {
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
		return "'" + vv.In(time.UTC).Format(datetimeLayouts[0]) + "'", nil
	case sql.NullTime: // timestamp
		return "'" + vv.Time.In(time.UTC).Format(datetimeLayouts[0]) + "'", nil
	}

	return fmt.Sprint(v), nil
}
