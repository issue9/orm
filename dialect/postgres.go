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
	"github.com/issue9/orm/v2/sqlbuilder"
)

const postgresName = "postgres"

var postgresInst *postgres

type postgres struct {
	replacer *strings.Replacer
}

var (
	_ sqlbuilder.TruncateTableStmtHooker = &postgres{}
)

// Postgres 返回一个适配 postgresql 的 Dialect 接口
func Postgres() core.Dialect {
	if postgresInst == nil {
		postgresInst = &postgres{
			replacer: strings.NewReplacer("{", `"`, "}", `"`),
		}
	}

	return postgresInst
}

func (p *postgres) Name() string {
	return postgresName
}

func (p *postgres) VersionSQL() string {
	return `SHOW server_version;`
}

func (p *postgres) Prepare(query string) (string, map[string]int, error) {
	query, orders, err := PrepareNamedArgs(query)
	if err != nil {
		return "", nil, err
	}

	query, err = p.replace(query)
	if err != nil {
		return "", nil, err
	}

	return query, orders, nil
}

func (p *postgres) LastInsertIDSQL(table, col string) (sql string, append bool) {
	return " RETURNING " + col, true
}

// 在有 ? 占位符的情况下，语句中不能包含 $ 字符串
func (p *postgres) SQL(query string, args []interface{}) (string, []interface{}, error) {
	query = replaceNamedArgs(query, args)
	query, err := p.replace(query)
	if err != nil {
		return "", nil, err
	}

	return query, args, nil
}

var errInvalidDollar = errors.New("语句中包含非法的字符串:$")

func (p *postgres) replace(query string) (string, error) {
	query = p.replacer.Replace(query)

	if strings.IndexByte(query, '?') < 0 {
		return query, nil
	}

	num := 1
	build := core.NewBuilder("")
	for _, c := range query {
		switch c {
		case '?':
			build.WriteBytes('$').WriteString(strconv.Itoa(num))
			num++
		case '$':
			return "", errInvalidDollar
		default:
			build.WriteRunes(c)
		}
	}

	return build.String()
}

func (p *postgres) CreateTableOptionsSQL(w *core.Builder, options map[string][]string) error {
	return nil
}

func (p *postgres) LimitSQL(limit interface{}, offset ...interface{}) (string, []interface{}) {
	return mysqlLimitSQL(limit, offset...)
}

func (p *postgres) TruncateTableStmtHook(stmt *sqlbuilder.TruncateTableStmt) ([]string, error) {
	builder := core.NewBuilder("TRUNCATE TABLE ").
		QuoteKey(stmt.TableName)

	if stmt.AIColumnName != "" {
		builder.WriteString(" RESTART IDENTITY")
	}

	query, err := builder.String()
	if err != nil {
		return nil, err
	}
	return []string{query}, nil
}

func (p *postgres) TransactionalDDL() bool {
	return true
}

func (p *postgres) SQLType(col *core.Column) (string, error) {
	if col == nil {
		return "", errColIsNil
	}

	if col.GoType == nil {
		return "", errGoTypeIsNil
	}

	switch col.GoType.Kind() {
	case reflect.Bool:
		return p.buildType("BOOLEAN", col, 0)
	case reflect.Int8, reflect.Int16, reflect.Uint8, reflect.Uint16:
		if col.AI {
			return p.buildType("SERIAL", col, 0)
		}
		return p.buildType("SMALLINT", col, 0)
	case reflect.Int32, reflect.Uint32:
		if col.AI {
			return p.buildType("SERIAL", col, 0)
		}
		return p.buildType("INT", col, 0)
	case reflect.Int64, reflect.Int, reflect.Uint64, reflect.Uint:
		if col.AI {
			return p.buildType("BIGSERIAL", col, 0)
		}
		return p.buildType("BIGINT", col, 0)
	case reflect.Float32, reflect.Float64:
		if len(col.Length) != 2 {
			return "", errMissLength
		}
		return p.buildType("NUMERIC", col, 2)
	case reflect.String:
		if len(col.Length) == 0 || (col.Length[0] == -1 || col.Length[0] > 65533) {
			return p.buildType("TEXT", col, 0)
		}
		return p.buildType("VARCHAR", col, 1)
	case reflect.Slice, reflect.Array:
		if col.GoType.Elem().Kind() == reflect.Uint8 {
			return p.buildType("BYTEA", col, 0)
		}
	case reflect.Struct:
		switch col.GoType {
		case core.RawBytesType:
			return p.buildType("BYTEA", col, 0)
		case core.NullBoolType:
			return p.buildType("BOOLEAN", col, 0)
		case core.NullFloat64Type:
			if len(col.Length) != 2 {
				return "", errMissLength
			}
			return p.buildType("NUMERIC", col, 2)
		case core.NullInt64Type:
			if col.AI {
				return p.buildType("BIGSERIAL", col, 0)
			}
			return p.buildType("BIGINT", col, 0)
		case core.NullStringType:
			if len(col.Length) == 0 || (col.Length[0] == -1 || col.Length[0] > 65533) {
				return p.buildType("TEXT", col, 0)
			}
			return p.buildType("VARCHAR", col, 1)
		case core.TimeType:
			if len(col.Length) > 0 && (col.Length[0] < 0 || col.Length[0] > 6) {
				return "", errTimeFractionalInvalid
			}
			return p.buildType("TIMESTAMP", col, 1)
		}
	}

	return "", errUncovert(col.GoType.Name())
}

// l 表示需要取的长度数量
func (p *postgres) buildType(typ string, col *core.Column, l int) (string, error) {
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

	if !col.Nullable {
		w.WriteString(" NOT NULL")
	}

	if col.HasDefault {
		v, err := p.SQLFormat(col.Default, col.Length...)
		if err != nil {
			return "", err
		}
		w.WriteString(" DEFAULT ").WriteString(v)
	}

	return w.String()
}

func (p *postgres) SQLFormat(v interface{}, length ...int) (f string, err error) {
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
	case string:
		return "'" + vv + "'", nil
	case time.Time: // timestamp
		if len(length) == 0 {
			return "'" + vv.Format(postgresDatetimeLayouts[0]) + "'", nil
		}
		if len(length) > 1 {
			return "", errTimeFractionalInvalid
		}

		if length[0] < 0 || length[0] > 6 {
			return "", errTimeFractionalInvalid
		}
		// layout 中带了时区信息
		return "'" + vv.Format(postgresDatetimeLayouts[length[0]]) + "'", nil
	}

	return fmt.Sprint(v), nil
}

var postgresDatetimeLayouts = []string{
	"2006-01-02 15:04:05Z07:00",
	"2006-01-02 15:04:05.9Z07:00",
	"2006-01-02 15:04:05.99Z07:00",
	"2006-01-02 15:04:05.999Z07:00",
	"2006-01-02 15:04:05.9999Z07:00",
	"2006-01-02 15:04:05.99999Z07:00",
	"2006-01-02 15:04:05.999999Z07:00",
}
