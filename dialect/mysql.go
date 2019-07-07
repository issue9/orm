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
	"github.com/issue9/orm/v2/internal/parser"
	"github.com/issue9/orm/v2/sqlbuilder"
)

const (
	mysqlName    = "mysql"
	mysqlEngine  = mysqlName + "_engine"
	mysqlCharset = mysqlName + "_charset"
)

var mysqlInst *mysql

type mysql struct{}

// Mysql 返回一个适配 mysql 的 Dialect 接口
//
// 支持以下 meta 属性
//  charset 字符集，语法为： charset(utf-8)
//  engine 使用的引擎，语法为： engine(innodb)
func Mysql() orm.Dialect {
	if mysqlInst == nil {
		mysqlInst = &mysql{}
	}

	return mysqlInst
}

func (m *mysql) Name() string {
	return mysqlName
}

func (m *mysql) QuoteTuple() (byte, byte) {
	return '`', '`'
}

func (m *mysql) SQL(sql string) (string, error) {
	return sql, nil
}

func (m *mysql) LastInsertIDSQL(table, col string) (sql string, append bool) {
	return "", false
}

func (m *mysql) VersionSQL() string {
	return `select version();`
}

func (m *mysql) CreateTableOptionsSQL(w *sqlbuilder.SQLBuilder, options map[string][]string) error {
	if len(options[mysqlEngine]) == 1 {
		w.WriteString(" ENGINE=")
		w.WriteString(options[mysqlEngine][0])
		w.WriteByte(' ')
	} else if len(options[mysqlEngine]) > 0 {
		return errors.New("无效的属性值：" + mysqlCharset)
	}

	if len(options[mysqlCharset]) == 1 {
		w.WriteString(" CHARACTER SET=")
		w.WriteString(options[mysqlCharset][0])
		w.WriteByte(' ')
	} else if len(options[mysqlCharset]) > 0 {
		return errors.New("无效的属性值：" + mysqlCharset)
	}

	return nil
}

func (m *mysql) LimitSQL(limit interface{}, offset ...interface{}) (string, []interface{}) {
	return mysqlLimitSQL(limit, offset...)
}

func (m *mysql) DropConstraintStmtHook(stmt *sqlbuilder.DropConstraintStmt) ([]string, error) {
	info, err := parser.ParseCreateTable("mysql", stmt.TableName, stmt.Engine())
	if err != nil {
		return nil, err
	}

	constraintType, found := info.Constraints[stmt.Name]
	if !found { // 不存在，也返回错误，统一与其它数据的行为
		return nil, fmt.Errorf("不存在的约束:%s", stmt.Name)
	}

	builder := sqlbuilder.New("ALTER TABLE ").
		WriteString(stmt.TableName).
		WriteString(" DROP ")
	switch constraintType {
	case sqlbuilder.ConstraintCheck:
		builder.WriteString(" CHECK ").WriteString(stmt.Name)
	case sqlbuilder.ConstraintFK:
		builder.WriteString(" FOREIGN KEY ").WriteString(stmt.Name)
	case sqlbuilder.ConstraintPK:
		builder.WriteString(" PRIMARY ")
	case sqlbuilder.ConstraintUnique:
		builder.WriteString(" INDEX ").WriteString(stmt.Name)
	default:
		panic(fmt.Sprintf("不存在的约束类型:%s", constraintType))
	}

	return []string{builder.String()}, nil
}

func (m *mysql) DropIndexStmtHook(stmt *sqlbuilder.DropIndexStmt) ([]string, error) {
	return []string{"ALTER TABLE " + stmt.TableName + " DROP INDEX " + stmt.IndexName}, nil
}

func (m *mysql) TruncateTableStmtHook(stmt *sqlbuilder.TruncateTableStmt) ([]string, error) {
	return []string{"TRUNCATE TABLE " + stmt.TableName}, nil
}

func (m *mysql) TransactionalDDL() bool {
	return false
}

func (m *mysql) SQLType(col *sqlbuilder.Column) (string, error) {
	if col == nil {
		return "", errColIsNil
	}

	if col.GoType == nil {
		return "", errGoTypeIsNil
	}

	switch col.GoType.Kind() {
	case reflect.Bool:
		return buildMysqlType("BOOLEAN", col, false, 0), nil
	case reflect.Int8:
		return buildMysqlType("SMALLINT", col, false, 1), nil
	case reflect.Int16:
		return buildMysqlType("MEDIUMINT", col, false, 1), nil
	case reflect.Int32:
		return buildMysqlType("INT", col, false, 1), nil
	case reflect.Int64, reflect.Int: // reflect.Int 大小未知，都当作是 BIGINT 处理
		return buildMysqlType("BIGINT", col, false, 1), nil
	case reflect.Uint8:
		return buildMysqlType("SMALLINT", col, true, 1), nil
	case reflect.Uint16:
		return buildMysqlType("MEDIUMINT", col, true, 1), nil
	case reflect.Uint32:
		return buildMysqlType("INT", col, true, 1), nil
	case reflect.Uint64, reflect.Uint, reflect.Uintptr:
		return buildMysqlType("BIGINT", col, true, 1), nil
	case reflect.Float32, reflect.Float64:
		if len(col.Length) != 2 {
			return "", errMissLength
		}
		return buildMysqlType("DOUBLE", col, false, 2), nil
	case reflect.String:
		if len(col.Length) == 0 || col.Length[0] == -1 || col.Length[0] > 65533 {
			return buildMysqlType("LONGTEXT", col, false, 0), nil
		}
		return buildMysqlType("VARCHAR", col, false, 1), nil
	case reflect.Slice, reflect.Array:
		if col.GoType.Elem().Kind() == reflect.Uint8 {
			return buildMysqlType("BLOB", col, false, 0), nil
		}
	case reflect.Struct:
		switch col.GoType {
		case rawBytes:
			return buildMysqlType("BLOB", col, false, 0), nil
		case nullBool:
			return buildMysqlType("BOOLEAN", col, false, 0), nil
		case nullFloat64:
			if len(col.Length) != 2 {
				return "", errMissLength
			}
			return buildMysqlType("DOUBLE", col, false, 2), nil
		case nullInt64:
			return buildMysqlType("BIGINT", col, false, 1), nil
		case nullString:
			if len(col.Length) == 0 || col.Length[0] == -1 || col.Length[0] > 65533 {
				return buildMysqlType("LONGTEXT", col, false, 0), nil
			}
			return buildMysqlType("VARCHAR", col, false, 1), nil
		case timeType:
			return buildMysqlType("DATETIME", col, false, 0), nil
		}
	}

	return "", errUncovert(col.GoType.Name())
}

// l 表示需要取的长度数量
func buildMysqlType(typ string, col *sqlbuilder.Column, unsigned bool, l int) string {
	w := sqlbuilder.New(typ)

	switch {
	case l == 1 && len(col.Length) > 0:
		w.WriteByte('(')
		w.WriteString(strconv.Itoa(col.Length[0]))
		w.WriteByte(')')
	case l == 2 && len(col.Length) > 1:
		w.WriteByte('(')
		w.WriteString(strconv.Itoa(col.Length[0]))
		w.WriteByte(',')
		w.WriteString(strconv.Itoa(col.Length[1]))
		w.WriteByte(')')
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
		w.WriteString(" DEFAULT '")
		w.WriteString(fmt.Sprint(col.Default)) // TODO: 是否需要和专门的转换？
		w.WriteByte('\'')
	}

	return w.String()
}
