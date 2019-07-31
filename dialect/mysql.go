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
	my "github.com/issue9/orm/v2/internal/mysql"
	"github.com/issue9/orm/v2/sqlbuilder"
)

const (
	mysqlName    = "mysql"
	mysqlEngine  = mysqlName + "_engine"
	mysqlCharset = mysqlName + "_charset"
)

var mysqlInst *mysql

type mysql struct {
	replacer *strings.Replacer
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
func Mysql() core.Dialect {
	if mysqlInst == nil {
		mysqlInst = &mysql{
			replacer: strings.NewReplacer("{", "`", "}", "`"),
		}
	}

	return mysqlInst
}

func (m *mysql) Name() string {
	return mysqlName
}

func (m *mysql) SQL(query string, args []interface{}) (string, []interface{}, error) {
	query = replaceNamedArgs(query, args)
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
		w.WriteString(" ENGINE=")
		w.WriteString(options[mysqlEngine][0])
		w.WriteBytes(' ')
	} else if len(options[mysqlEngine]) > 0 {
		return errors.New("无效的属性值：" + mysqlCharset)
	}

	if len(options[mysqlCharset]) == 1 {
		w.WriteString(" CHARACTER SET=")
		w.WriteString(options[mysqlCharset][0])
		w.WriteBytes(' ')
	} else if len(options[mysqlCharset]) > 0 {
		return errors.New("无效的属性值：" + mysqlCharset)
	}

	return nil
}

func (m *mysql) LimitSQL(limit interface{}, offset ...interface{}) (string, []interface{}) {
	return mysqlLimitSQL(limit, offset...)
}

func (m *mysql) DropConstraintStmtHook(stmt *sqlbuilder.DropConstraintStmt) ([]string, error) {
	info, err := my.ParseCreateTable(stmt.TableName, stmt.Engine())
	if err != nil {
		return nil, err
	}

	constraintType, found := info.Constraints[stmt.Name]
	if !found { // 不存在，也返回错误，统一与其它数据的行为
		return nil, fmt.Errorf("不存在的约束:%s", stmt.Name)
	}

	builder := core.NewBuilder("ALTER TABLE ").
		WriteString(stmt.TableName).
		WriteString(" DROP ")
	switch constraintType {
	case core.ConstraintCheck:
		builder.WriteString(" CHECK ").WriteString(stmt.Name)
	case core.ConstraintFK:
		builder.WriteString(" FOREIGN KEY ").WriteString(stmt.Name)
	case core.ConstraintPK:
		builder.WriteString(" PRIMARY KEY")
	case core.ConstraintUnique:
		builder.WriteString(" INDEX ").WriteString(stmt.Name)
	default:
		panic(fmt.Sprintf("不存在的约束类型:%s", constraintType))
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
		WriteString(" DROP INDEX ").
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
		builder.WriteString(" OR REPLACE ")
	}

	if stmt.IsTemporary {
		builder.WriteString(" ALGORITHM=TEMPTABLE ")
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

	builder.WriteString(" AS ").WriteString(stmt.SelectQuery)

	query, err := builder.String()
	if err != nil {
		return nil, err
	}
	return []string{query}, nil
}

func (m *mysql) InsertDefaultValueHook(table string) (string, []interface{}, error) {
	query, err := core.NewBuilder("INSERT INTO").
		QuoteKey(table).
		WriteString("() VALUES ()").
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

	if col.GoType == nil {
		return "", errGoTypeIsNil
	}

	switch col.GoType.Kind() {
	case reflect.Bool:
		return m.buildType("BOOLEAN", col, false, 0)
	case reflect.Int8:
		return m.buildType("SMALLINT", col, false, 1)
	case reflect.Int16:
		return m.buildType("MEDIUMINT", col, false, 1)
	case reflect.Int32:
		return m.buildType("INT", col, false, 1)
	case reflect.Int64, reflect.Int: // reflect.Int 大小未知，都当作是 BIGINT 处理
		return m.buildType("BIGINT", col, false, 1)
	case reflect.Uint8:
		return m.buildType("SMALLINT", col, true, 1)
	case reflect.Uint16:
		return m.buildType("MEDIUMINT", col, true, 1)
	case reflect.Uint32:
		return m.buildType("INT", col, true, 1)
	case reflect.Uint64, reflect.Uint, reflect.Uintptr:
		return m.buildType("BIGINT", col, true, 1)
	case reflect.Float32, reflect.Float64:
		if len(col.Length) != 2 {
			return "", errMissLength
		}
		return m.buildType("DOUBLE", col, false, 2)
	case reflect.String:
		if len(col.Length) == 0 || col.Length[0] == -1 || col.Length[0] > 65533 {
			return m.buildType("LONGTEXT", col, false, 0)
		}
		return m.buildType("VARCHAR", col, false, 1)
	case reflect.Slice, reflect.Array:
		if col.GoType.Elem().Kind() == reflect.Uint8 {
			return m.buildType("BLOB", col, false, 0)
		}
	case reflect.Struct:
		switch col.GoType {
		case core.RawBytesType:
			return m.buildType("BLOB", col, false, 0)
		case core.NullBoolType:
			return m.buildType("BOOLEAN", col, false, 0)
		case core.NullFloat64Type:
			if len(col.Length) != 2 {
				return "", errMissLength
			}
			return m.buildType("DOUBLE", col, false, 2)
		case core.NullInt64Type:
			return m.buildType("BIGINT", col, false, 1)
		case core.NullStringType:
			if len(col.Length) == 0 || col.Length[0] == -1 || col.Length[0] > 65533 {
				return m.buildType("LONGTEXT", col, false, 0)
			}
			return m.buildType("VARCHAR", col, false, 1)
		case core.TimeType:
			if len(col.Length) > 0 && (col.Length[0] < 0 || col.Length[0] > 6) {
				return "", errTimeFractionalInvalid
			}
			return m.buildType("DATETIME", col, false, 0)
		}
	}

	return "", errUncovert(col.GoType.Name())
}

// l 表示需要取的长度数量
func (m *mysql) buildType(typ string, col *core.Column, unsigned bool, l int) (string, error) {
	w := core.NewBuilder(typ)

	switch {
	case l == 1 && len(col.Length) > 0:
		w.Quote(strconv.Itoa(col.Length[0]), '(', ')')
	case l == 2 && len(col.Length) > 1:
		w.WriteBytes('(').
			WriteString(strconv.Itoa(col.Length[0])).
			WriteBytes(',').
			WriteString(strconv.Itoa(col.Length[1])).
			WriteBytes(')')
	}

	if unsigned {
		w.WriteString(" UNSIGNED")
	}

	if col.AI {
		w.WriteString(" PRIMARY KEY AUTO_INCREMENT")
	}

	if !col.Nullable {
		w.WriteString(" NOT NULL")
	}

	if col.HasDefault {
		v, err := m.SQLFormat(col.Default, col.Length...)
		if err != nil {
			return "", err
		}
		w.WriteString(" DEFAULT ").WriteString(v)
	}

	return w.String()
}

func (m *mysql) SQLFormat(v interface{}, length ...int) (f string, err error) {
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
		vv = vv.In(time.UTC)
		if len(length) == 0 {
			return "'" + vv.Format(mysqlDatetimeLayouts[0]) + "'", nil
		}
		if len(length) > 1 {
			return "", errTimeFractionalInvalid
		}

		if length[0] < 0 || length[0] > 6 {
			return "", errTimeFractionalInvalid
		}
		return "'" + vv.Format(mysqlDatetimeLayouts[length[0]]) + "'", nil
	}

	return fmt.Sprint(v), nil
}

var mysqlDatetimeLayouts = []string{
	"2006-01-02 15:04:05",
	"2006-01-02 15:04:05.9",
	"2006-01-02 15:04:05.99",
	"2006-01-02 15:04:05.999",
	"2006-01-02 15:04:05.9999",
	"2006-01-02 15:04:05.99999",
	"2006-01-02 15:04:05.999999",
}
