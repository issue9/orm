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

const (
	mysqlEngine  = "mysql_engine"
	mysqlCharset = "mysql_charset"
)

type mysql struct {
	base
	isMariadb bool
}

var (
	_ sqlbuilder.DropConstraintStmtHooker = &mysql{}
	_ sqlbuilder.InsertDefaultValueHooker = &mysql{}
)

// Mysql 返回一个适配 mysql 的 Dialect 接口
//
// 支持以下 meta 属性
//  mysql_charset 字符集，语法为： charset(utf-8)
//  mysql_engine 使用的引擎，语法为： engine(innodb)
func Mysql(driverName, tablePrefix string) core.Dialect {
	return newMysql(false, "mysql", driverName, tablePrefix)
}

// Mariadb 返回一个适配 mariadb 的 Dialect 接口
//
// meta 属性可参考 mysql，大部分内容与 Mysql 相同。
func Mariadb(driverName, tablePrefix string) core.Dialect {
	return newMysql(true, "mariadb", driverName, tablePrefix)
}

func newMysql(isMariadb bool, dbName, driverName, tablePrefix string) core.Dialect {
	return &mysql{
		base:      newBase(dbName, driverName, tablePrefix, "`", "`"),
		isMariadb: isMariadb,
	}
}

func (m *mysql) Fix(query string, args []any) (string, []any, error) {
	query, args, err := fixQueryAndArgs(query, args)
	if err != nil {
		return "", nil, err
	}
	return m.replacer.Replace(query), args, nil
}

func (m *mysql) LastInsertIDSQL(table, col string) (sql string, append bool) {
	return "", false
}

func (m *mysql) VersionSQL() string { return `select version();` }

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

func (m *mysql) LimitSQL(limit any, offset ...any) (string, []any) {
	return mysqlLimitSQL(limit, offset...)
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

func (m *mysql) DropIndexSQL(table, index string) (string, error) {
	if table == "" {
		return "", sqlbuilder.ErrTableIsEmpty
	}
	if index == "" {
		return "", sqlbuilder.ErrColumnsIsEmpty
	}

	return core.NewBuilder("ALTER TABLE ").
		QuoteKey(table).
		WString(" DROP INDEX ").
		QuoteKey(index).
		String()
}

func (m *mysql) TruncateTableSQL(table, ai string) ([]string, error) {
	builder := core.NewBuilder("TRUNCATE TABLE ").QuoteKey(table)

	query, err := builder.String()
	if err != nil {
		return nil, err
	}
	return []string{query}, nil
}

func (m *mysql) CreateViewSQL(replace, temporary bool, name, selectQuery string, cols []string) ([]string, error) {
	builder := core.NewBuilder("CREATE ")

	if replace {
		builder.WString(" OR REPLACE ")
	}

	if temporary {
		builder.WString(" ALGORITHM=TEMPTABLE ")
	}

	q, err := appendViewBody(builder, name, selectQuery, cols)
	if err != nil {
		return nil, err
	}
	return []string{q}, nil
}

func (m *mysql) InsertDefaultValueHook(table string) (string, []any, error) {
	query, err := core.NewBuilder("INSERT INTO").
		QuoteKey(table).
		WString("() VALUES ()").
		String()

	if err != nil {
		return "", nil, err
	}
	return query, nil, nil
}

func (m *mysql) TransactionalDDL() bool { return false }

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
	case core.Float32:
		return m.buildType("FLOAT", col, false, 0)
	case core.Float64:
		return m.buildType("DOUBLE PRECISION", col, false, 0)
	case core.Decimal:
		if len(col.Length) != 2 {
			return "", missLength(col)
		}
		return m.buildType("DECIMAL", col, false, 2)
	case core.String:
		if len(col.Length) == 0 || col.Length[0] == -1 || col.Length[0] > 65533 {
			return m.buildType("LONGTEXT", col, false, 0)
		}
		return m.buildType("VARCHAR", col, false, 1)
	case core.Bytes:
		return m.buildType("BLOB", col, false, 0)
	case core.Time:
		if len(col.Length) == 0 {
			return m.buildType("DATETIME", col, false, 0)
		}
		if col.Length[0] < 0 || col.Length[0] > 6 {
			return "", invalidTimeFractional(col)
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
		v, err := m.formatSQL(col)
		if err != nil {
			return "", err
		}
		w.WString(" DEFAULT ").WString(v)
	}

	return w.String()
}

func (m *mysql) formatSQL(col *core.Column) (f string, err error) {
	v := col.Default
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
		return formatTime(col, vv)
	case sql.NullTime: // datetime
		return formatTime(col, vv.Time)
	}

	return fmt.Sprint(v), nil
}

func formatTime(col *core.Column, t time.Time) (string, error) {
	t = t.In(time.UTC)

	if len(col.Length) == 0 {
		return "'" + t.Format(datetimeLayouts[0]) + "'", nil
	}
	if len(col.Length) > 1 {
		return "", invalidTimeFractional(col)
	}

	index := col.Length[0]
	if index < 0 || index > 6 {
		return "", invalidTimeFractional(col)
	}
	return "'" + t.Format(datetimeLayouts[index]) + "'", nil
}
