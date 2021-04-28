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

	"github.com/issue9/orm/v3/core"
	"github.com/issue9/orm/v3/internal/createtable"
	"github.com/issue9/orm/v3/sqlbuilder"
)

const (
	mysqlEngine  = "mysql_engine"
	mysqlCharset = "mysql_charset"
)

type mysql struct {
	isMariadb  bool
	dbName     string
	driverName string
	replacer   *strings.Replacer
}

var (
	_ sqlbuilder.TruncateTableStmtHooker  = &mysql{}
	_ sqlbuilder.DropIndexStmtHooker      = &mysql{}
	_ sqlbuilder.DropConstraintStmtHooker = &mysql{}
	_ sqlbuilder.CreateViewStmtHooker     = &mysql{}
	_ sqlbuilder.InsertDefaultValueHooker = &mysql{}
)

// Mysql 返回一个适配 mysql 的 Dialect 接口
//
// 支持以下 meta 属性
//  charset 字符集，语法为： charset(utf-8)
//  engine 使用的引擎，语法为： engine(innodb)
func Mysql(driverName string) core.Dialect {
	return newMysql(false, "mysql", driverName)
}

// Mariadb 返回一个适配 mariadb 的 Dialect 接口
//
// meta 属性可参考 mysql，大部分内容增多与 Mysql 相同。
func Mariadb(driverName string) core.Dialect {
	return newMysql(true, "mariadb", driverName)
}

func newMysql(isMariadb bool, name, driverName string) core.Dialect {
	return &mysql{
		isMariadb:  isMariadb,
		dbName:     name,
		driverName: driverName,
		replacer:   strings.NewReplacer("{", "`", "}", "`"),
	}
}

func (m *mysql) DBName() string {
	return m.dbName
}

func (m *mysql) DriverName() string {
	return m.driverName
}

func (m *mysql) Fix(query string, args []interface{}) (string, []interface{}, error) {
	query = ReplaceNamedArgs(query, args)
	return m.replacer.Replace(query), args, nil
}

func (m *mysql) LastInsertIDSQL(table, col string) (sql string, append bool) {
	return "", false
}

func (m *mysql) VersionSQL() string {
	return `select version();`
}

func (m *mysql) Prepare(query string) (string, map[string]int, error) {
	query, orders, err := PrepareNamedArgs(query)
	if err != nil {
		return "", nil, err
	}
	return m.replacer.Replace(query), orders, nil
}

func (m *mysql) CreateTableOptionsSQL(w *core.Builder, options map[string][]string) error {
	if len(options[mysqlEngine]) == 1 {
		w.WString(" ENGINE=")
		w.WString(options[mysqlEngine][0])
		w.WBytes(' ')
	} else if len(options[mysqlEngine]) > 0 {
		return errors.New("无效的属性值：" + mysqlCharset)
	}

	if len(options[mysqlCharset]) == 1 {
		w.WString(" CHARACTER SET=")
		w.WString(options[mysqlCharset][0])
		w.WBytes(' ')
	} else if len(options[mysqlCharset]) > 0 {
		return errors.New("无效的属性值：" + mysqlCharset)
	}

	return nil
}

func (m *mysql) LimitSQL(limit interface{}, offset ...interface{}) (string, []interface{}) {
	return MysqlLimitSQL(limit, offset...)
}

func (m *mysql) DropConstraintStmtHook(stmt *sqlbuilder.DropConstraintStmt) ([]string, error) {
	info, err := createtable.ParseMysqlCreateTable(stmt.TableName, stmt.Engine())
	if err != nil {
		return nil, err
	}

	builder := core.NewBuilder("ALTER TABLE ").
		WString(stmt.TableName).
		WString(" DROP ")
	if stmt.IsPK {
		query, err := builder.WString(" PRIMARY KEY").String()
		if err != nil {
			return nil, err
		}
		return []string{query}, nil
	}

	name := strings.Replace(stmt.Name, "#", stmt.Engine().TablePrefix(), 1)
	constraintType, found := info.Constraints[name]
	if !found { // 不存在，也返回错误，统一与其它数据的行为
		return nil, fmt.Errorf("不存在的约束:%s", name)
	}

	switch constraintType {
	case core.ConstraintFK:
		builder.WString(" FOREIGN KEY ").WString(stmt.Name)
	case core.ConstraintPK:
		builder.WString(" PRIMARY KEY")
	case core.ConstraintUnique:
		builder.WString(" INDEX ").WString(stmt.Name)
	default:
		if constraintType == core.ConstraintCheck {
			if m.isMariadb {
				builder.WString(" CONSTRAINT ").WString(stmt.Name)
			} else {
				builder.WString(" CHECK ").WString(stmt.Name)
			}
		} else {
			panic(fmt.Sprintf("不存在的约束类型:%d", constraintType))
		}
	}

	query, err := builder.String()
	if err != nil {
		return nil, err
	}
	return []string{query}, nil
}

func (m *mysql) DropIndexStmtHook(stmt *sqlbuilder.DropIndexStmt) ([]string, error) {
	builder := core.NewBuilder("ALTER TABLE ").
		QuoteKey(stmt.TableName).
		WString(" DROP INDEX ").
		QuoteKey(stmt.IndexName)

	query, err := builder.String()
	if err != nil {
		return nil, err
	}
	return []string{query}, nil
}

func (m *mysql) TruncateTableStmtHook(stmt *sqlbuilder.TruncateTableStmt) ([]string, error) {
	builder := core.NewBuilder("TRUNCATE TABLE ").QuoteKey(stmt.TableName)

	query, err := builder.String()
	if err != nil {
		return nil, err
	}
	return []string{query}, nil
}

func (m *mysql) CreateViewStmtHook(stmt *sqlbuilder.CreateViewStmt) ([]string, error) {
	builder := core.NewBuilder("CREATE ")

	if stmt.IsReplace {
		builder.WString(" OR REPLACE ")
	}

	if stmt.IsTemporary {
		builder.WString(" ALGORITHM=TEMPTABLE ")
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

	builder.WString(" AS ").WString(stmt.SelectQuery)

	query, err := builder.String()
	if err != nil {
		return nil, err
	}
	return []string{query}, nil
}

func (m *mysql) InsertDefaultValueHook(table string) (string, []interface{}, error) {
	query, err := core.NewBuilder("INSERT INTO").
		QuoteKey(table).
		WString("() VALUES ()").
		String()

	if err != nil {
		return "", nil, err
	}
	return query, nil, nil
}

func (m *mysql) TransactionalDDL() bool {
	return false
}

func (m *mysql) SQLType(col *core.Column) (string, error) {
	if col == nil {
		return "", errColIsNil
	}

	switch col.PrimitiveType {
	case core.Bool:
		return m.buildType("BOOLEAN", col, false, 0)
	case core.Int8:
		return m.buildType("SMALLINT", col, false, 1)
	case core.Int16:
		return m.buildType("MEDIUMINT", col, false, 1)
	case core.Int32:
		return m.buildType("INT", col, false, 1)
	case core.Int64, core.Int: // reflect.Int 大小未知，都当作是 BIGINT 处理
		return m.buildType("BIGINT", col, false, 1)
	case core.Uint8:
		return m.buildType("SMALLINT", col, true, 1)
	case core.Uint16:
		return m.buildType("MEDIUMINT", col, true, 1)
	case core.Uint32:
		return m.buildType("INT", col, true, 1)
	case core.Uint64, core.Uint:
		return m.buildType("BIGINT", col, true, 1)
	case core.Float32, core.Float64:
		if len(col.Length) != 2 {
			return "", errMissLength
		}
		return m.buildType("DOUBLE", col, false, 2)
	case core.String:
		if len(col.Length) == 0 || col.Length[0] == -1 || col.Length[0] > 65533 {
			return m.buildType("LONGTEXT", col, false, 0)
		}
		return m.buildType("VARCHAR", col, false, 1)
	case core.RawBytes, core.Bytes:
		return m.buildType("BLOB", col, false, 0)
	case core.NullBool:
		return m.buildType("BOOLEAN", col, false, 0)
	case core.NullFloat64:
		if len(col.Length) != 2 {
			return "", errMissLength
		}
		return m.buildType("DOUBLE", col, false, 2)
	case core.NullInt64:
		return m.buildType("BIGINT", col, false, 1)
	case core.NullString:
		if len(col.Length) == 0 || col.Length[0] == -1 || col.Length[0] > 65533 {
			return m.buildType("LONGTEXT", col, false, 0)
		}
		return m.buildType("VARCHAR", col, false, 1)
	case core.Time, core.NullTime:
		if len(col.Length) == 0 {
			return m.buildType("DATETIME", col, false, 0)
		}
		if col.Length[0] < 0 || col.Length[0] > 6 {
			return "", errTimeFractionalInvalid
		}
		return m.buildType("DATETIME", col, false, 1)
	}

	return "", errUncovert(col)
}

// l 表示需要取的长度数量
func (m *mysql) buildType(typ string, col *core.Column, unsigned bool, l int) (string, error) {
	w := core.NewBuilder(typ)

	switch {
	case l == 1 && len(col.Length) > 0:
		w.Quote(strconv.Itoa(col.Length[0]), '(', ')')
	case l == 2 && len(col.Length) > 1:
		w.WBytes('(').
			WString(strconv.Itoa(col.Length[0])).
			WBytes(',').
			WString(strconv.Itoa(col.Length[1])).
			WBytes(')')
	}

	if unsigned {
		w.WString(" UNSIGNED")
	}

	if col.AI {
		w.WString(" PRIMARY KEY AUTO_INCREMENT")
	}

	if !col.Nullable {
		w.WString(" NOT NULL")
	}

	if col.HasDefault {
		v, err := m.formatSQL(col.Default, col.Length...)
		if err != nil {
			return "", err
		}
		w.WString(" DEFAULT ").WString(v)
	}

	return w.String()
}

func (m *mysql) formatSQL(v interface{}, length ...int) (f string, err error) {
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
	case bool:
		if vv {
			return "1", nil
		}
		return "0", nil
	case string:
		return "'" + vv + "'", nil
	case time.Time: // datetime
		return m.formatTime(vv, length...)
	case sql.NullTime: // datetime
		return m.formatTime(vv.Time, length...)
	}

	return fmt.Sprint(v), nil
}

func (m *mysql) formatTime(t time.Time, length ...int) (string, error) {
	t = t.In(time.UTC)

	if len(length) == 0 {
		return "'" + t.Format(datetimeLayouts[0]) + "'", nil
	}
	if len(length) > 1 {
		return "", errTimeFractionalInvalid
	}

	index := length[0]
	if index < 0 || index > 6 {
		return "", errTimeFractionalInvalid
	}
	return "'" + t.Format(datetimeLayouts[index]) + "'", nil
}
