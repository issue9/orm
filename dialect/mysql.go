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

func (m *mysql) CreateColumnSQL(buf *sqlbuilder.SQLBuilder, col *sqlbuilder.Column, isAI bool) error {
	buf.WriteString(col.Name)
	buf.WriteByte(' ')

	buf.WriteString(col.Type).WriteByte(' ')

	if !col.Nullable {
		buf.WriteString(" NOT NULL")
	}

	if isAI {
		buf.WriteString(" PRIMARY KEY AUTO_INCREMENT ")
	}

	if col.HasDefault {
		buf.WriteString(" DEFAULT '").
			WriteString(col.Default).
			WriteByte('\'')
	}

	return nil
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

func (m *mysql) DropIndexSQL(table, index string) (string, []interface{}) {
	return `ALTER TABLE {` + table + `} DROP INDEX {` + index + `}`, nil
}

func (m *mysql) TruncateTableSQL(model *orm.Model) []string {
	return []string{"TRUNCATE TABLE #" + model.Name}
}

func (m *mysql) TransactionalDDL() bool {
	return false
}

func (m *mysql) SQLType(col *orm.Column) (string, error) {
	if col == nil {
		return "", errColIsNil
	}

	if col.GoType == nil {
		return "", errGoTypeIsNil
	}

	intLen := func(typ string) string {
		if col.Len1 > 0 {
			return typ + "(" + strconv.Itoa(col.Len1) + ")"
		}
		return typ
	}

	switch col.GoType.Kind() {
	case reflect.Bool:
		return "BOOLEAN", nil
	case reflect.Int8:
		return intLen("SMALLINT"), nil
	case reflect.Int16:
		return intLen("MEDIUMINT"), nil
	case reflect.Int32:
		return intLen("INT"), nil
	case reflect.Int64, reflect.Int: // reflect.Int 大小未知，都当作是 BIGINT 处理
		return intLen("BIGINT"), nil
	case reflect.Uint8:
		return intLen("SMALLINT") + " UNSIGNED", nil
	case reflect.Uint16:
		return intLen("MEDIUMINT") + " UNSIGNED", nil
	case reflect.Uint32:
		return intLen("INT") + " UNSIGNED", nil
	case reflect.Uint64, reflect.Uint, reflect.Uintptr:
		return intLen("BIGINT") + " UNSIGNED", nil
	case reflect.Float32, reflect.Float64:
		if col.Len1 == 0 || col.Len2 == 0 {
			return "", errMissLength
		}
		return fmt.Sprintf("DOUBLE(%d,%d)", col.Len1, col.Len2), nil
	case reflect.String:
		if col.Len1 == -1 || col.Len1 > 65533 {
			return "LONGTEXT", nil
		}
		return fmt.Sprintf("VARCHAR(%d)", col.Len1), nil
	case reflect.Slice, reflect.Array:
		if col.GoType.Elem().Kind() == reflect.Uint8 {
			return "BLOB", nil
		}
	case reflect.Struct:
		switch col.GoType {
		case rawBytes:
			return "BLOB", nil
		case nullBool:
			return "BOOLEAN", nil
		case nullFloat64:
			if col.Len1 == 0 || col.Len2 == 0 {
				return "", errMissLength
			}
			return fmt.Sprintf("DOUBLE(%d,%d)", col.Len1, col.Len2), nil
		case nullInt64:
			return intLen("BIGINT"), nil
		case nullString:
			if col.Len1 == -1 || col.Len1 > 65533 {
				return "LONGTEXT", nil
			}
			return fmt.Sprintf("VARCHAR(%d)", col.Len1), nil
		case timeType:
			return "DATETIME", nil
		}
	}

	return "", errUncovert(col.GoType.Name())
}
