// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/issue9/orm"
	"github.com/issue9/orm/sqlbuilder"
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

func (m *mysql) QuoteTuple() (byte, byte) {
	return '`', '`'
}

func (m *mysql) SQL(sql string) (string, error) {
	return sql, nil
}

func (m *mysql) LastInsertID(table, col string) (sql string, append bool) {
	return "", false
}

func (m *mysql) CreateTableSQL(model *orm.Model) ([]string, error) {
	w := sqlbuilder.New("CREATE TABLE IF NOT EXISTS ").
		WriteString("{#").
		WriteString(model.Name).
		WriteString("}(")

	// 自增列
	if model.AI != nil {
		if err := createColSQL(m, w, model.AI); err != nil {
			return nil, err
		}
		w.WriteString(" PRIMARY KEY AUTO_INCREMENT,")
	}

	// 普通列
	for _, col := range model.Cols {
		if col.IsAI() { // 忽略 AI 列
			continue
		}

		if err := createColSQL(m, w, col); err != nil {
			return nil, err
		}
		w.WriteByte(',')
	}

	// 约束
	if len(model.PK) > 0 && !model.PK[0].IsAI() { // PK，若有自增，则已经在上面指定
		createPKSQL(w, model.PK, pkName)
		w.WriteByte(',')
	}
	createConstraints(w, model)

	// index
	m.createIndexSQL(w, model)

	w.TruncateLast(1).WriteByte(')')

	if err := m.createTableOptions(w, model); err != nil {
		return nil, err
	}

	return []string{w.String()}, nil
}

func (m *mysql) createTableOptions(w *sqlbuilder.SQLBuilder, model *orm.Model) error {
	if len(model.Meta["mysql_engine"]) == 1 {
		w.WriteString(" ENGINE=")
		w.WriteString(model.Meta["mysql_engine"][0])
		w.WriteByte(' ')
	} else if len(model.Meta["mysql_engine"]) > 0 {
		return errors.New("无效的属性值 engine")
	}

	if len(model.Meta["mysql_charset"]) == 1 {
		w.WriteString(" CHARACTER SET=")
		w.WriteString(model.Meta["mysql_charset"][0])
		w.WriteByte(' ')
	} else if len(model.Meta["mysql_charset"]) > 0 {
		return errors.New("无效的属性值 mysql_charset")
	}

	return nil
}

func (m *mysql) createIndexSQL(w *sqlbuilder.SQLBuilder, model *orm.Model) {
	for indexName, cols := range model.KeyIndexes {
		// INDEX index_name (id,lastName)
		w.WriteString(" INDEX ").
			WriteString(indexName).
			WriteByte('(')
		for _, col := range cols {
			w.WriteByte('{').WriteString(col.Name).WriteByte('}')
			w.WriteByte(',')
		}
		w.TruncateLast(1) // 去掉最后一个逗号

		w.WriteString("),")
	}
}

func (m *mysql) LimitSQL(limit interface{}, offset ...interface{}) (string, []interface{}) {
	return mysqlLimitSQL(limit, offset...)
}

func (m *mysql) TruncateTableSQL(model *orm.Model) []string {
	return []string{"TRUNCATE TABLE #" + model.Name}
}

func (m *mysql) TransactionalDDL() bool {
	return false
}

func (m *mysql) sqlType(buf *sqlbuilder.SQLBuilder, col *orm.Column) error {
	if col == nil {
		return errors.New("sqlType:col 参数是个空值")
	}

	if col.GoType == nil {
		return errors.New("sqlType:无效的 col.GoType 值")
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
	case reflect.Int64, reflect.Int: // reflect.Int 大小未知，都当作是 BIGINT 处理
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
	case reflect.Slice, reflect.Array:
		if col.GoType.Elem().Kind() == reflect.Uint8 {
			buf.WriteString("BLOB")
		}
	case reflect.Struct:
		switch col.GoType {
		case rawBytes:
			buf.WriteString("BLOB")
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
