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

	"github.com/issue9/orm"
	"github.com/issue9/orm/model"
	"github.com/issue9/orm/sqlbuilder"
)

// Postgres 返回一个适配 postgresql 的 Dialect 接口
func Postgres() orm.Dialect {
	return &postgres{}
}

type postgres struct{}

func (p *postgres) QuoteTuple() (byte, byte) {
	return '"', '"'
}

// 在有 ? 占位符的情况下，语句中不能包含$字符串
func (p *postgres) SQL(sql string) (string, error) {
	if strings.IndexByte(sql, '?') < 0 {
		return sql, nil
	}

	num := 1
	ret := make([]rune, 0, len(sql))
	for _, c := range sql {
		switch c {
		case '?':
			ret = append(ret, '$')
			ret = append(ret, []rune(strconv.Itoa(num))...)
			num++
		case '$':
			return "", errors.New("语句中包含非法的字符串:$")
		default:
			ret = append(ret, c)
		}
	}

	return string(ret), nil
}

func (p *postgres) CreateTableSQL(model *model.Model) (string, error) {
	w := sqlbuilder.New("CREATE TABLE IF NOT EXISTS ").
		WriteString("{#").
		WriteString(model.Name).
		WriteString("}(")

	// 自增和普通列输出是相同的，自增列仅是类型名不相同
	for _, col := range model.Cols {
		if err := createColSQL(p, w, col); err != nil {
			return "", err
		}
		w.WriteByte(',')
	}

	if len(model.PK) > 0 {
		createPKSQL(w, model.PK, model.Name+pkName) // postgres 主键名需要全局唯一？
		w.WriteByte(',')
	}
	createConstraints(w, model)
	w.TruncateLast(1).WriteByte(')')

	// TODO meta
	return w.String(), nil
}

func (p *postgres) LimitSQL(limit interface{}, offset ...interface{}) (string, []interface{}) {
	return mysqlLimitSQL(limit, offset...)
}

func (p *postgres) TruncateTableSQL(table, ai string) string {
	w := sqlbuilder.New("TRUNCATE TABLE ").WriteString(table)

	if ai != "" {
		w.WriteString(" RESTART IDENTITY")
	}

	return w.String()
}

func (p *postgres) TransactionalDDL() bool {
	return true
}

// implement base.sqlType
// 将col转换成sql类型，并写入buf中。
func (p *postgres) sqlType(buf *sqlbuilder.SQLBuilder, col *model.Column) error {
	if col == nil {
		return errors.New("sqlType:col参数是个空值")
	}

	if col.GoType == nil {
		return errors.New("sqlType:无效的col.GoType值")
	}

	switch col.GoType.Kind() {
	case reflect.Bool:
		buf.WriteString("BOOLEAN")
	case reflect.Int8, reflect.Int16, reflect.Uint8, reflect.Uint16:
		if col.IsAI() {
			buf.WriteString("SERIAL")
		} else {
			buf.WriteString("SMALLINT")
		}
	case reflect.Int32, reflect.Uint32:
		if col.IsAI() {
			buf.WriteString("SERIAL")
		} else {
			buf.WriteString("INT")
		}
	case reflect.Int64, reflect.Int, reflect.Uint64, reflect.Uint:
		if col.IsAI() {
			buf.WriteString("BIGSERIAL")
		} else {
			buf.WriteString("BIGINT")
		}
	case reflect.Float32, reflect.Float64:
		if col.Len1 == 0 || col.Len2 == 0 {
			return errors.New("请指定长度")
		}
		buf.WriteString(fmt.Sprintf("DOUBLE(%d,%d)", col.Len1, col.Len2))
	case reflect.String:
		if col.Len1 == -1 || col.Len1 > 65533 {
			buf.WriteString("TEXT")
		} else {
			buf.WriteString(fmt.Sprintf("VARCHAR(%d)", col.Len1))
		}
	case reflect.Slice, reflect.Array: // []rune,[]byte当作字符串处理
		k := col.GoType.Elem().Kind()
		if (k != reflect.Uint8) && (k != reflect.Int32) {
			return errors.New("sqlType:不支持数组类型")
		}

		if col.Len1 == -1 || col.Len1 > 65533 {
			buf.WriteString("TEXT")
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
			if col.IsAI() {
				buf.WriteString("BIGSERIAL")
			} else {
				buf.WriteString("BIGINT")
			}
		case nullString:
			if col.Len1 == -1 || col.Len1 > 65533 {
				buf.WriteString("TEXT")
			} else {
				buf.WriteString(fmt.Sprintf("VARCHAR(%d)", col.Len1))
			}
		case timeType:
			buf.WriteString("TIME")
		}
	default:
		return fmt.Errorf("sqlType:不支持的类型:[%v]", col.GoType.Name())
	}

	return nil
}
