// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/issue9/orm/core"
)

// Mysql 返回一个适配 mysql 的 core.Dialect 接口
func Mysql() core.Dialect {
	return &mysql{}
}

type mysql struct{}

func (m *mysql) Name() string {
	return "mysql"
}

func (m *mysql) SupportInsertMany() bool {
	return true
}

func (m *mysql) QuoteTuple() (byte, byte) {
	return '`', '`'
}

func (m *mysql) SQL(sql string) (string, error) {
	return sql, nil
}

func (m *mysql) CreateTableSQL(w *core.StringBuilder, model *core.Model) error {
	w.WriteString("CREATE TABLE IF NOT EXISTS ").
		WriteString("{#").
		WriteString(model.Name).
		WriteString("}(")

	// 自增列
	if model.AI != nil {
		if err := createColSQL(m, w, model.AI); err != nil {
			return err
		}
		w.WriteString(" PRIMARY KEY AUTO_INCREMENT,")
	}

	// 普通列
	for _, col := range model.Cols {
		if col.IsAI() { // 忽略AI列
			continue
		}

		if err := createColSQL(m, w, col); err != nil {
			return err
		}
		w.WriteByte(',')
	}

	// 约束
	if len(model.PK) > 0 && !model.PK[0].IsAI() { // PK，若有自增，则已经在上面指定
		createPKSQL(m, w, model.PK, pkName)
		w.WriteByte(',')
	}
	createConstraints(m, w, model)

	w.TruncateLast(1).WriteByte(')')
	return nil
}

func (m *mysql) LimitSQL(limit int, offset ...int) (string, []interface{}) {
	return mysqlLimitSQL(limit, offset...)
}

func (m *mysql) TruncateTableSQL(w *core.StringBuilder, tableName, aiColumn string) {
	w.WriteString("TRUNCATE TABLE ").WriteString(tableName)
}

func (m *mysql) sqlType(buf *core.StringBuilder, col *core.Column) error {
	if col == nil {
		return errors.New("sqlType:col参数是个空值")
	}

	if col.GoType == nil {
		return errors.New("sqlType:无效的col.GoType值")
	}

	addIntLen := func() {
		if col.Len1 > 0 {
			buf.WriteByte('(').
				WriteString(strconv.Itoa(col.Len1)).
				WriteByte(')')
		}
	}

	switch col.GoType.Kind() {
	case reflect.Bool:
		buf.WriteString("BOOLEAN")
	case reflect.Int8:
		buf.WriteString("SMALLINT")
		addIntLen()
	case reflect.Int16:
		buf.WriteString("MEDIUMINT")
		addIntLen()
	case reflect.Int32:
		buf.WriteString("INT")
		addIntLen()
	case reflect.Int64, reflect.Int: // reflect.Int大小未知，都当作是BIGINT处理
		buf.WriteString("BIGINT")
		addIntLen()
	case reflect.Uint8:
		buf.WriteString("SMALLINT")
		addIntLen()
		buf.WriteString(" UNSIGNED")
	case reflect.Uint16:
		buf.WriteString("MEDIUMINT")
		addIntLen()
		buf.WriteString(" UNSIGNED")
	case reflect.Uint32:
		buf.WriteString("INT")
		addIntLen()
		buf.WriteString(" UNSIGNED")
	case reflect.Uint64, reflect.Uint, reflect.Uintptr:
		buf.WriteString("BIGINT")
		addIntLen()
		buf.WriteString(" UNSIGNED")
	case reflect.Float32, reflect.Float64:
		if col.Len1 == 0 || col.Len2 == 0 {
			return errors.New("请指定长度")
		}
		buf.WriteString(fmt.Sprintf("DOUBLE(%d,%d)", col.Len1, col.Len2))
	case reflect.String:
		if col.Len1 == -1 || col.Len1 > 65533 {
			buf.WriteString("LONGTEXT")
		} else {
			buf.WriteString(fmt.Sprintf("VARCHAR(%d)", col.Len1))
		}
	case reflect.Slice, reflect.Array: // []rune,[]byte当作字符串处理
		k := col.GoType.Elem().Kind()
		if (k != reflect.Uint8) && (k != reflect.Int32) {
			return fmt.Errorf("sqlType:不支持[%v]类型的数组", k)
		}

		if col.Len1 == -1 || col.Len1 > 65533 {
			buf.WriteString("LONGTEXT")
		} else {
			buf.WriteString(fmt.Sprintf("VARCHAR(%d)", col.Len1))
		}
	case reflect.Struct:
		switch col.GoType {
		case nullBool:
			buf.WriteString("BOOLEAN")
		case nullFloat64:
			if col.Len1 == 0 || col.Len2 == 0 {
				return errors.New("请指定长度")
			}
			buf.WriteString(fmt.Sprintf("DOUBLE(%d,%d)", col.Len1, col.Len2))
		case nullInt64:
			buf.WriteString("BIGINT")
			addIntLen()
		case nullString:
			if col.Len1 == -1 || col.Len1 > 65533 {
				buf.WriteString("LONGTEXT")
			} else {
				buf.WriteString(fmt.Sprintf("VARCHAR(%d)", col.Len1))
			}
		case timeType:
			buf.WriteString("DATETIME")
		}
	default:
		return fmt.Errorf("sqlType:不支持的类型:[%v]", col.GoType.Name())
	}

	return nil
}
