// SPDX-License-Identifier: MIT

package dialect_test

import (
	"database/sql"
	"math"
	"reflect"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v3/core"
	"github.com/issue9/orm/v3/internal/sqltest"
	"github.com/issue9/orm/v3/internal/test"
	"github.com/issue9/orm/v3/sqlbuilder"
)

type sqlTypeTester struct {
	col     *core.Column
	err     bool
	SQLType string
}

func testSQLType(a *assert.Assertion, d core.Dialect, data []*sqlTypeTester) {
	for index, item := range data {
		typ, err := d.SQLType(item.col)
		if item.err {
			a.Error(err, "not error @%d", index)
		} else {
			a.NotError(err, "%v @%d", err, index)
			sqltest.Equal(a, typ, item.SQLType)
		}
	}
}

func testDialectVersionSQL(t *test.Driver) {
	rows, err := t.DB.Query(t.DB.Dialect().VersionSQL())
	t.NotError(err).NotNil(rows)

	defer func() {
		t.NotError(rows.Close())
	}()

	t.True(rows.Next())
	var ver string
	t.NotError(rows.Scan(&ver))
	t.NotEmpty(ver)
}

func testDialectDropConstraintStmtHook(t *test.Driver) {
	db := t.DB

	// 不存在的约束，出错
	stmt := sqlbuilder.DropConstraint(db).
		Table("fk_table").
		Constraint("id_great_zero")

	t.Error(stmt.Exec())

	err := sqlbuilder.AddConstraint(db).
		Table("fk_table").
		Check("id_great_zero", "id>0").
		Exec()
	t.NotError(err)

	// 约束已经添加，可以正常删除
	// check
	stmt.Reset()
	err = stmt.Table("fk_table").Constraint("id_great_zero").Exec()
	t.NotError(err)

	// fk
	stmt.Reset()
	err = stmt.Table("usr").Constraint("xxx_fk").Exec()
	t.NotError(err)

	// unique
	stmt.Reset()
	err = stmt.Table("usr").Constraint("u_user_xx1").Exec()
	t.NotError(err)

	// pk
	stmt.Reset()
	err = stmt.Table("usr").Constraint(core.PKName("usr")).Exec()
	t.NotError(err)
}

func testTypes(t *test.Driver) {
	tableName := "test_type_read_write"
	now := time.Now()

	creator := sqlbuilder.CreateTable(t.DB).
		Column("bool", reflect.TypeOf(true), false, false, false, nil).
		Column("int", reflect.TypeOf(1), false, false, false, nil).
		Column("int8", reflect.TypeOf(int8(1)), false, false, false, nil).
		Column("int16", reflect.TypeOf(int16(1)), false, false, false, nil).
		Column("int32", reflect.TypeOf(int32(1)), false, false, false, nil).
		Column("int64", reflect.TypeOf(int64(1)), false, false, false, nil).
		Column("uint", reflect.TypeOf(uint(1)), false, false, false, nil).
		Column("uint8", reflect.TypeOf(uint8(1)), false, false, false, nil).
		Column("uint16", reflect.TypeOf(uint16(1)), false, false, false, nil).
		Column("uint32", reflect.TypeOf(uint32(1)), false, false, false, nil).
		Column("uint64", reflect.TypeOf(uint64(1)), false, false, false, nil).
		Column("float32", reflect.TypeOf(float32(1)), false, false, false, nil, 5, 3).
		Column("float64", reflect.TypeOf(float64(1)), false, false, false, nil, 5, 3).
		Column("string", reflect.TypeOf(""), false, false, false, nil, 100).
		Column("null_string", reflect.TypeOf(sql.NullString{}), false, false, false, nil, 100).
		Column("null_int64", reflect.TypeOf(sql.NullInt64{}), false, false, false, nil).
		Column("null_bool", reflect.TypeOf(sql.NullBool{}), false, false, false, nil).
		Column("null_float64", reflect.TypeOf(sql.NullFloat64{}), false, false, false, nil, 5, 3).
		Column("raw_bytes", reflect.TypeOf(sql.RawBytes{}), false, false, false, nil).
		Column("time", reflect.TypeOf(time.Time{}), false, false, false, nil).
		Column("null_time", reflect.TypeOf(sql.NullTime{}), false, false, false, nil, 5).
		Column("unix", reflect.TypeOf(core.Unix{}), false, false, false, nil).
		Table(tableName)
	t.NotError(creator.Exec())
	defer func() {
		t.NotError(sqlbuilder.DropTable(t.DB).Table(tableName).Exec())
	}()

	cols := []string{
		"bool",
		"int",
		"int8",
		"int16",
		"int32",
		"int64",
		"uint",
		"uint8",
		"uint16",
		"uint32",
		"uint64",
		"float32",
		"float64",
		"string",
		"null_string",
		"null_int64",
		"null_bool",
		"null_float64",
		"raw_bytes",
		"time",
		"null_time",
		"unix",
	}
	vals := []interface{}{ // 与 cols 一一对应
		true,
		-1,
		-8,
		-16,
		-32,
		-64,
		1,
		8,
		16,
		32,
		64,
		-1.32,
		1.64,
		"str",
		"null_str",
		164,
		true,
		.64,
		sql.RawBytes("rawBytes"),
		now,
		now,
		core.Unix(now),
	}

	r, err := sqlbuilder.Insert(t.DB).
		Table(tableName).
		Columns(cols...).
		Values(vals...).
		Exec()
	t.NotError(err).NotNil(r)

	selStmt := sqlbuilder.Select(t.DB).
		From(tableName)
	quoteColumns(selStmt, cols...)
	rows, err := selStmt.Query()
	t.NotError(err).NotNil(rows)
	defer func() {
		t.NotError(rows.Close())
	}()

	t.True(rows.Next())
	var (
		Bool       bool
		Int        int
		Int8       int8
		Int16      int16
		Int32      int32
		Int64      int64
		Uint       uint
		Uint8      uint8
		Uint16     uint16
		Uint32     uint32
		Uint64     uint64
		F32        float32
		F64        float64
		String     string
		NullString sql.NullString
		NullInt64  sql.NullInt64
		NullBool   sql.NullBool
		NullF64    sql.NullFloat64
		RawBytes   sql.RawBytes
		Time       time.Time
		NullTime   time.Time
		Unix       core.Unix
	)
	err = rows.Scan(&Bool, &Int, &Int8, &Int16, &Int32, &Int64,
		&Uint, &Uint8, &Uint16, &Uint32, &Uint64,
		&F32, &F64, &String, &NullString, &NullInt64, &NullBool, &NullF64, &RawBytes, &Time, &NullTime, &Unix)
	t.NotError(err)
	t.True(Bool).
		Equal(Int, -1).
		Equal(Int8, -8).
		Equal(Int16, -16).
		Equal(Int32, -32).
		Equal(Int64, -64).
		Equal(Uint, 1).
		Equal(Uint8, 8).
		Equal(Uint16, 16).
		Equal(Uint32, 32).
		Equal(Uint64, 64).
		Equal(F32, float32(-1.32)).
		Equal(F64, 1.64).
		Equal(String, "str").
		True(NullString.Valid).Equal(NullString.String, "null_str").
		True(NullInt64.Valid).Equal(NullInt64.Int64, 164).
		True(NullBool.Valid).True(NullBool.Bool).
		True(NullF64.Valid).Equal(NullF64.Float64, .64).
		Equal(RawBytes, sql.RawBytes("rawBytes")).
		// bug(caixw): mysql 对无精度的保存会取整
		True(math.Abs(float64(Time.Unix()-now.Unix())) < 2, "Time not true\n%v:%d\n%v:%d", Time, Time.Unix(), now, now.Unix()).
		Equal(NullTime.Unix(), now.Unix(), "NullTime not equal\n%v:%d\n%v:%d", Time, Time.Unix(), now, now.Unix()).
		Equal(Unix.AsTime().Unix(), now.Unix(), "Unix not equal\n%v:%d\n%v:%d", Unix.AsTime(), Unix.AsTime().Unix(), now, now.Unix())

	//	fmt.Printf("\n%s,%v,%v", t.DriverName, Time, now)
}

func quoteColumns(stmt *sqlbuilder.SelectStmt, col ...string) {
	for _, c := range col {
		stmt.Column("{" + c + "}")
	}
}

func testTypesDefault(t *test.Driver) {
	tableName := "test_type_read_write"
	now := time.Now()

	creator := sqlbuilder.CreateTable(t.DB).
		Column("bool", reflect.TypeOf(true), false, false, true, false).
		Column("int", reflect.TypeOf(1), false, false, true, -1).
		Column("int8", reflect.TypeOf(int8(1)), false, false, true, -8).
		Column("int16", reflect.TypeOf(int16(1)), false, false, true, 0).
		Column("int32", reflect.TypeOf(int32(1)), false, false, true, 32).
		Column("int64", reflect.TypeOf(int64(1)), false, false, true, -64).
		Column("uint", reflect.TypeOf(uint(1)), false, false, true, 0).
		Column("uint8", reflect.TypeOf(uint8(1)), false, false, true, 8).
		Column("uint16", reflect.TypeOf(uint16(1)), false, false, true, 16).
		Column("uint32", reflect.TypeOf(uint32(1)), false, false, true, 32).
		Column("uint64", reflect.TypeOf(uint64(1)), false, false, true, 64).
		Column("float32", reflect.TypeOf(float32(1)), false, false, true, -3.2, 5, 3).
		Column("float64", reflect.TypeOf(float64(1)), false, false, true, 6.654321, 15, 7).
		Column("string", reflect.TypeOf(""), false, false, true, "str", 100).
		Column("null_string", reflect.TypeOf(sql.NullString{}), false, false, true, "null_str", 100).
		Column("null_int64", reflect.TypeOf(sql.NullInt64{}), false, true, true, sql.NullInt64{Int64: 64, Valid: false}).
		Column("null_bool", reflect.TypeOf(sql.NullBool{}), false, false, true, sql.NullBool{Bool: true, Valid: true}).
		Column("null_float64", reflect.TypeOf(sql.NullFloat64{}), false, true, true, nil, 5, 3).
		Column("raw_bytes", reflect.TypeOf(sql.RawBytes{}), false, true, false, nil).
		Column("time", reflect.TypeOf(time.Time{}), false, false, true, now).
		Column("time_with_len", reflect.TypeOf(time.Time{}), false, false, true, now, 5).
		Column("unix", reflect.TypeOf(&core.Unix{}), false, false, true, now.Format(core.TimeFormatLayout)).
		Table(tableName)
	t.NotError(creator.Exec())
	defer func() {
		t.NotError(sqlbuilder.DropTable(t.DB).Table(tableName).Exec())
	}()

	cols := []string{
		"bool",
		"int",
		"int8",
		"int16",
		"int32",
		"int64",
		"uint",
		"uint8",
		"uint16",
		"uint32",
		"uint64",
		"float32",
		"float64",
		"string",
		"null_string",
		"null_int64",
		"null_bool",
		"null_float64",
		"raw_bytes",
		"time",
		"time_with_len",
		"unix",
	}
	r, err := sqlbuilder.Insert(t.DB).
		Table(tableName).
		Exec()
	t.NotError(err).NotNil(r)

	selStmt := sqlbuilder.Select(t.DB).
		From(tableName)
	quoteColumns(selStmt, cols...)
	rows, err := selStmt.Query()
	t.NotError(err).NotNil(rows)
	defer func() {
		t.NotError(rows.Close())
	}()

	t.True(rows.Next())
	var (
		Bool        bool
		Int         int
		Int8        int8
		Int16       int16
		Int32       int32
		Int64       int64
		Uint        uint
		Uint8       uint8
		Uint16      uint16
		Uint32      uint32
		Uint64      uint64
		F32         float32
		F64         float64
		String      string
		NullString  sql.NullString
		NullInt64   sql.NullInt64
		NullBool    sql.NullBool
		NullF64     sql.NullFloat64
		RawBytes    sql.RawBytes
		Time        time.Time
		TimeWithLen time.Time
		Unix        core.Unix
	)
	err = rows.Scan(&Bool, &Int, &Int8, &Int16, &Int32, &Int64,
		&Uint, &Uint8, &Uint16, &Uint32, &Uint64,
		&F32, &F64, &String, &NullString, &NullInt64, &NullBool, &NullF64, &RawBytes, &Time, &TimeWithLen, &Unix)
	t.NotError(err)
	t.False(Bool).
		Equal(Int, -1).
		Equal(Int8, -8).
		Equal(Int16, 0).
		Equal(Int32, 32).
		Equal(Int64, -64).
		Equal(Uint, 0).
		Equal(Uint8, 8).
		Equal(Uint16, 16).
		Equal(Uint32, 32).
		Equal(Uint64, 64).
		Equal(F32, float32(-3.2)).
		Equal(F64, 6.654321).
		Equal(String, "str").
		True(NullString.Valid).Equal(NullString.String, "null_str").
		False(NullInt64.Valid).
		True(NullBool.Valid).True(NullBool.Bool).
		False(NullF64.Valid).
		Nil(RawBytes).
		Equal(Time.Unix(), now.Unix()).
		Equal(TimeWithLen.Unix(), now.Unix()).
		Equal(Unix.AsTime().Unix(), now.Unix())
	//fmt.Printf("\n%s,%v,%v", t.DriverName, Time, now)
}
