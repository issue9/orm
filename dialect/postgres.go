// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"regexp"

	"github.com/issue9/orm/core"
)

type Postgres struct{}

// implement core.Dialect.QuoteStr()
func (p *Postgres) QuoteStr() (l, r string) {
	return `"`, `"`
}

// 匹配dbname=dbname 或是dbname =dbname等格式
var dbnamePrefix = regexp.MustCompile(`\s*=\s*|\s+`)

// implement core.Dialect.GetDBName()
func (p *Postgres) GetDBName(dataSource string) string {
	// dataSource样式：user=user dbname = db password=
	words := dbnamePrefix.Split(dataSource, -1)
	for index, word := range words {
		if word != "dbname" {
			continue
		}

		if index+1 >= len(words) {
			return ""
		}

		return words[index+1]
	}

	return ""
}

// implement core.Dialect.LimitSQL()
func (p *Postgres) LimitSQL(limit int, offset ...int) (string, []interface{}) {
	return mysqlLimitSQL(limit, offset...)
}

// implement core.Dialect.CreateTableSQL()
func (p *Postgres) CreateTableSQL(model *core.Model) (string, error) {
	buf := bytes.NewBufferString("CREATE TABLE IF NOT EXISTS ")
	buf.Grow(300)

	buf.WriteString(model.Name)
	buf.WriteByte('(')

	// 写入字段信息
	for _, col := range model.Cols {
		if err := createColSQL(p, buf, col); err != nil {
			return "", err
		}
		buf.WriteByte(',')
	}

	// PK
	if len(model.PK) > 0 {
		createPKSQL(p, buf, model.PK, pkName)
		buf.WriteByte(',')
	}

	// Unique Index
	for name, index := range model.UniqueIndexes {
		createUniqueSQL(p, buf, index, name)
		buf.WriteByte(',')
	}

	// foreign  key
	for name, fk := range model.FK {
		createFKSQL(p, buf, fk, name)
		buf.WriteByte(',')
	}

	// Check
	for name, chk := range model.Check {
		createCheckSQL(p, buf, chk, name)
		buf.WriteByte(',')
	}

	buf.Truncate(buf.Len() - 1) // 去掉最后的逗号
	buf.WriteByte(')')          // end CreateTable

	return buf.String(), nil
}

// implement base.sqlType
// 将col转换成sql类型，并写入buf中。
func (p *Postgres) sqlType(buf *bytes.Buffer, col *core.Column) error {
	if col == nil {
		return errors.New("col参数是个空值")
	}

	if col.GoType == nil {
		return errors.New("无效的col.GoType值")
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
		buf.WriteString(fmt.Sprintf("DOUBLE(%d,%d)", col.Len1, col.Len2))
	case reflect.String:
		if col.Len1 < 65533 {
			buf.WriteString(fmt.Sprintf("VARCHAR(%d)", col.Len1))
		} else {
			buf.WriteString("TEXT")
		}
	case reflect.Slice, reflect.Array: // []rune,[]byte当作字符串处理
		k := col.GoType.Elem().Kind()
		if (k != reflect.Uint8) && (k != reflect.Int32) {
			return errors.New("不支持数组类型")
		}

		if col.Len1 < 65533 {
			buf.WriteString(fmt.Sprintf("VARCHAR(%d)", col.Len1))
		} else {
			buf.WriteString("TEXT")
		}
	case reflect.Struct:
		switch col.GoType {
		case nullBool:
			buf.WriteString("BOOLEAN")
		case nullFloat64:
			buf.WriteString(fmt.Sprintf("DOUBLE(%d,%d)", col.Len1, col.Len2))
		case nullInt64:
			if col.IsAI() {
				buf.WriteString("BIGSERIAL")
			} else {
				buf.WriteString("BIGINT")
			}
		case nullString:
			if col.Len1 < 65533 {
				buf.WriteString(fmt.Sprintf("VARCHAR(%d)", col.Len1))
			}
			buf.WriteString("TEXT")
		case timeType:
			buf.WriteString("TIME")
		}
	default:
		return fmt.Errorf("不支持的类型:[%v]", col.GoType.Name())
	}

	return nil
}
