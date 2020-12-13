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

	"github.com/issue9/orm/v3/core"
	"github.com/issue9/orm/v3/sqlbuilder"
)

type postgres struct {
	driverName string
	replacer   *strings.Replacer
}

var (
	_ sqlbuilder.TruncateTableStmtHooker = &postgres{}
)

// Postgres 返回一个适配 postgresql 的 Dialect 接口
func Postgres(driverName string) core.Dialect {
	return &postgres{
		driverName: driverName,
		replacer:   strings.NewReplacer("{", `"`, "}", `"`),
	}
}

func (p *postgres) DBName() string {
	return "postgres"
}

func (p *postgres) DriverName() string {
	return p.driverName
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
func (p *postgres) Fix(query string, args []interface{}) (string, []interface{}, error) {
	query = ReplaceNamedArgs(query, args)
	query, err := p.replace(query)
	if err != nil {
		return "", nil, err
	}

	// lib/pq 对 time.Time 的处理有问题，保存时不会考虑其时区，
	// 直接从字面值当作零时区进行保存。
	// https://github.com/lib/pq/issues/329
	for index, arg := range args {
		switch a := arg.(type) {
		case time.Time:
			args[index] = a.In(time.UTC)
		case *time.Time:
			args[index] = a.In(time.UTC)
		case sql.NullTime:
			args[index] = sql.NullTime{Valid: a.Valid, Time: a.Time.In(time.UTC)}
		case *sql.NullTime:
			args[index] = &sql.NullTime{Valid: a.Valid, Time: a.Time.In(time.UTC)}
		}
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
			build.WBytes('$').WString(strconv.Itoa(num))
			num++
		case '$':
			return "", errInvalidDollar
		default:
			build.WRunes(c)
		}
	}

	return build.String()
}

func (p *postgres) CreateTableOptionsSQL(w *core.Builder, options map[string][]string) error {
	return nil
}

func (p *postgres) LimitSQL(limit interface{}, offset ...interface{}) (string, []interface{}) {
	return MysqlLimitSQL(limit, offset...)
}

func (p *postgres) TruncateTableStmtHook(stmt *sqlbuilder.TruncateTableStmt) ([]string, error) {
	builder := core.NewBuilder("TRUNCATE TABLE ").
		QuoteKey(stmt.TableName)

	if stmt.AIColumnName != "" {
		builder.WString(" RESTART IDENTITY")
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

	switch col.PrimitiveType {
	case core.Bool:
		return p.buildType("BOOLEAN", col, 0)
	case core.Int8, core.Int16, core.Uint8, core.Uint16:
		if col.AI {
			return p.buildType("SERIAL", col, 0)
		}
		return p.buildType("SMALLINT", col, 0)
	case core.Int32, core.Uint32:
		if col.AI {
			return p.buildType("SERIAL", col, 0)
		}
		return p.buildType("INT", col, 0)
	case core.Int64, core.Int, core.Uint64, core.Uint:
		if col.AI {
			return p.buildType("BIGSERIAL", col, 0)
		}
		return p.buildType("BIGINT", col, 0)
	case core.Float32, core.Float64:
		if len(col.Length) != 2 {
			return "", errMissLength
		}
		return p.buildType("NUMERIC", col, 2)
	case core.String:
		if len(col.Length) == 0 || (col.Length[0] == -1 || col.Length[0] > 65533) {
			return p.buildType("TEXT", col, 0)
		}
		return p.buildType("VARCHAR", col, 1)
	case core.RawBytes:
		return p.buildType("BYTEA", col, 0)
	case core.NullBool:
		return p.buildType("BOOLEAN", col, 0)
	case core.NullFloat64:
		if len(col.Length) != 2 {
			return "", errMissLength
		}
		return p.buildType("NUMERIC", col, 2)
	case core.NullInt64:
		if col.AI {
			return p.buildType("BIGSERIAL", col, 0)
		}
		return p.buildType("BIGINT", col, 0)
	case core.NullString:
		if len(col.Length) == 0 || (col.Length[0] == -1 || col.Length[0] > 65533) {
			return p.buildType("TEXT", col, 0)
		}
		return p.buildType("VARCHAR", col, 1)
	case core.Time, core.NullTime:
		if len(col.Length) == 0 {
			return p.buildType("TIMESTAMP", col, 0)
		}
		if col.Length[0] < 0 || col.Length[0] > 6 {
			return "", errTimeFractionalInvalid
		}
		return p.buildType("TIMESTAMP", col, 1)
	}

	return "", errUncovert(col)
}

// l 表示需要取的长度数量
func (p *postgres) buildType(typ string, col *core.Column, l int) (string, error) {
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

	if !col.Nullable {
		w.WString(" NOT NULL")
	}

	if col.HasDefault {
		v, err := p.formatSQL(col.Default, col.Length...)
		if err != nil {
			return "", err
		}
		w.WString(" DEFAULT ").WString(v)
	}

	return w.String()
}

func (p *postgres) formatSQL(v interface{}, length ...int) (f string, err error) {
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
		return p.formatTime(vv, length...)
	case sql.NullTime: // timestamp
		return p.formatTime(vv.Time, length...)
	}

	return fmt.Sprint(v), nil
}

func (p *postgres) formatTime(t time.Time, length ...int) (string, error) {
	t = t.In(time.UTC)

	if len(length) == 0 {
		return "'" + t.Format(datetimeLayouts[6]) + "'", nil
	}
	if len(length) > 1 {
		return "", errTimeFractionalInvalid
	}

	index := length[0]
	if index < 0 || index > 6 {
		return "", errTimeFractionalInvalid
	}
	return "'" + t.Format(datetimeLayouts[index]) + "'", nil
}
