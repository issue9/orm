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

	"github.com/issue9/orm/v2"
	"github.com/issue9/orm/v2/sqlbuilder"
)

const postgresName = "postgres"

var postgresInst *postgres

type postgres struct{}

// Postgres 返回一个适配 postgresql 的 Dialect 接口
func Postgres() orm.Dialect {
	if postgresInst == nil {
		postgresInst = &postgres{}
	}

	return postgresInst
}

func (p *postgres) Name() string {
	return postgresName
}

func (p *postgres) QuoteTuple() (byte, byte) {
	return '"', '"'
}

func (p *postgres) VersionSQL() string {
	return `SHOW server_version;`
}

func (p *postgres) LastInsertIDSQL(table, col string) (sql string, append bool) {
	return " RETURNING {" + col + "}", true

	// 在同时插入多列时， RETURNING 会返回每一列的 ID。
	// 最终 sql.Result.LastInsertId() 返回的是最小的那个值，
	// 这在大部分时间不符合要求。
	// return " RETURNING {" + col + "}", true, true
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

func (p *postgres) CreateColumnSQL(buf *sqlbuilder.SQLBuilder, col *sqlbuilder.Column, isAI bool) error {
	buf.WriteByte('{').WriteString(col.Name).WriteByte('}')
	buf.WriteByte(' ')

	if isAI {

		buf.WriteString(" SERIAL ")
	} else {
		buf.WriteString(col.Type).WriteByte(' ')
	}

	if !col.Nullable {
		buf.WriteString(" NOT NULL")
	}

	if col.HasDefault {
		buf.WriteString(" DEFAULT '").
			WriteString(col.Default).
			WriteByte('\'')
	}

	return nil
}

func (p *postgres) CreateTableOptionsSQL(w *sqlbuilder.SQLBuilder, options map[string][]string) error {
	return nil
}

func (p *postgres) LimitSQL(limit interface{}, offset ...interface{}) (string, []interface{}) {
	return mysqlLimitSQL(limit, offset...)
}

func (p *postgres) DropIndexSQL(table, index string) (string, []interface{}) {
	return `DROP INDEX IF EXISTS {` + index + `}`, nil
}

func (p *postgres) TruncateTableSQL(m *orm.Model) []string {
	w := sqlbuilder.New("TRUNCATE TABLE #").WriteString(m.Name)

	if m.AI != nil {
		w.WriteString(" RESTART IDENTITY")
	}

	return []string{w.String()}
}

func (p *postgres) TransactionalDDL() bool {
	return true
}

func (p *postgres) SQLType(col *orm.Column) (string, error) {
	if col == nil {
		return "", errColIsNil
	}

	if col.GoType == nil {
		return "", errGoTypeIsNil
	}

	switch col.GoType.Kind() {
	case reflect.Bool:
		return "BOOLEAN", nil
	case reflect.Int8, reflect.Int16, reflect.Uint8, reflect.Uint16:
		if col.IsAI() {
			return "SERIAL", nil
		}
		return "SMALLINT", nil
	case reflect.Int32, reflect.Uint32:
		if col.IsAI() {
			return "SERIAL", nil
		}
		return "INT", nil
	case reflect.Int64, reflect.Int, reflect.Uint64, reflect.Uint:
		if col.IsAI() {
			return "BIGSERIAL", nil
		}
		return "BIGINT", nil
	case reflect.Float32, reflect.Float64:
		if col.Len1 == 0 || col.Len2 == 0 {
			return "", errMissLength
		}
		return fmt.Sprintf("NUMERIC(%d,%d)", col.Len1, col.Len2), nil
	case reflect.String:
		if col.Len1 == -1 || col.Len1 > 65533 {
			return ("TEXT"), nil
		}
		return (fmt.Sprintf("VARCHAR(%d)", col.Len1)), nil
	case reflect.Slice, reflect.Array:
		if col.GoType.Elem().Kind() == reflect.Uint8 {
			return "BYTEA", nil
		}
	case reflect.Struct:
		switch col.GoType {
		case rawBytes:
			return "BYTEA", nil
		case nullBool:
			return "BOOLEAN", nil
		case nullFloat64:
			if col.Len1 == 0 || col.Len2 == 0 {
				return "", errMissLength
			}
			return fmt.Sprintf("NUMERIC(%d,%d)", col.Len1, col.Len2), nil
		case nullInt64:
			if col.IsAI() {
				return "BIGSERIAL", nil
			}
			return "BIGINT", nil
		case nullString:
			if col.Len1 == -1 || col.Len1 > 65533 {
				return "TEXT", nil
			}
			return fmt.Sprintf("VARCHAR(%d)", col.Len1), nil
		case timeType:
			return fmt.Sprintf("TIMESTAMP(%d)", col.Len1), nil
		}
	}

	return "", errUncovert(col.GoType.Name())
}
