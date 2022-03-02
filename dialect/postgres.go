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

	"github.com/issue9/orm/v5/core"
)

type postgres struct {
	base
}

// Postgres 返回一个适配 postgresql 的 Dialect 接口
func Postgres(driverName string) core.Dialect {
	return &postgres{
		base: newBase("postgres", driverName, `"`, `"`),
	}
}

func (p *postgres) VersionSQL() string { return `SHOW server_version;` }

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

// Fix 在有 ? 占位符的情况下，语句中不能包含 $ 字符串
func (p *postgres) Fix(query string, args []any) (string, []any, error) {
	query, args, err := fixQueryAndArgs(query, args)
	if err != nil {
		return "", nil, err
	}

	query, err = p.replace(query)
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

func (p *postgres) LimitSQL(limit any, offset ...any) (string, []any) {
	return mysqlLimitSQL(limit, offset...)
}

func (p *postgres) TruncateTableSQL(table, ai string) ([]string, error) {
	builder := core.NewBuilder("TRUNCATE TABLE ").
		QuoteKey(table)

	if ai != "" {
		builder.WString(" RESTART IDENTITY")
	}

	query, err := builder.String()
	if err != nil {
		return nil, err
	}
	return []string{query}, nil
}

func (p *postgres) CreateViewSQL(replace, temporary bool, name, selectQuery string, cols []string) ([]string, error) {
	builder := core.NewBuilder("CREATE ")

	if replace {
		builder.WString(" OR REPLACE ")
	}

	if temporary {
		builder.WString(" TEMPORARY ")
	}

	q, err := appendViewBody(builder, name, selectQuery, cols)
	if err != nil {
		return nil, err
	}
	return []string{q}, nil
}

func (p *postgres) TransactionalDDL() bool { return true }

func (p *postgres) DropIndexSQL(table, index string) (string, error) {
	return stdDropIndex(index)
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
	case core.Float32:
		return p.buildType("REAL", col, 0)
	case core.Float64:
		return p.buildType("DOUBLE PRECISION", col, 0)
	case core.Decimal:
		if len(col.Length) != 2 {
			return "", missLength(col)
		}
		return p.buildType("DECIMAL", col, 2)
	case core.String:
		if len(col.Length) == 0 || (col.Length[0] == -1 || col.Length[0] > 65533) {
			return p.buildType("TEXT", col, 0)
		}
		return p.buildType("VARCHAR", col, 1)
	case core.Bytes:
		return p.buildType("BYTEA", col, 0)
	case core.Time:
		if len(col.Length) == 0 {
			return p.buildType("TIMESTAMP", col, 0)
		}
		if col.Length[0] < 0 || col.Length[0] > 6 {
			return "", invalidTimeFractional(col)
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
		v, err := p.formatSQL(col)
		if err != nil {
			return "", err
		}
		w.WString(" DEFAULT ").WString(v)
	}

	return w.String()
}

func (p *postgres) formatSQL(col *core.Column) (f string, err error) {
	v := col.Default
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
		return formatTime(col, vv)
	case sql.NullTime: // timestamp
		return formatTime(col, vv.Time)
	}

	return fmt.Sprint(v), nil
}
