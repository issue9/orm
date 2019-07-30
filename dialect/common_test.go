// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect_test

import (
	"database/sql"
	"math"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/core"
	"github.com/issue9/orm/v2/internal/sqltest"
	"github.com/issue9/orm/v2/internal/test"
	"github.com/issue9/orm/v2/sqlbuilder"
)

type sqlFormatTester struct {
	v      interface{} // 格式化的值
	l      []int       // 长度值
	format string      // 格式化之后的值
	err    bool        // 是否有错误
}

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

func testSQLFormat(a *assert.Assertion, d core.Dialect, data []*sqlFormatTester) {
	for index, item := range data {
		f, err := d.SQLFormat(item.v, item.l...)
		if item.err {
			a.Error(err, "not error @%d", index).
				Empty(f)
		} else {
			a.NotError(err, "%v @%d", err, index).
				Equal(f, item.format, "not equal @%d,v1:%s,v2:%s", index, f, item.format)
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
		Column("bool", core.BoolType, false, false, nil).
		Column("int", core.IntType, false, false, nil).
		Column("int8", core.Int8Type, false, false, nil).
		Column("int16", core.Int16Type, false, false, nil).
		Column("int32", core.Int32Type, false, false, nil).
		Column("int64", core.Int64Type, false, false, nil).
		Column("uint", core.UintType, false, false, nil).
		Column("uint8", core.Uint8Type, false, false, nil).
		Column("uint16", core.Uint16Type, false, false, nil).
		Column("uint32", core.Uint32Type, false, false, nil).
		Column("uint64", core.Uint64Type, false, false, nil).
		Column("float32", core.Float32Type, false, false, nil, 5, 3).
		Column("float64", core.Float64Type, false, false, nil, 5, 3).
		Column("string", core.StringType, false, false, nil, 100).
		Column("null_string", core.NullStringType, false, false, nil, 100).
		Column("null_int64", core.NullInt64Type, false, false, nil).
		Column("null_bool", core.NullBoolType, false, false, nil).
		Column("null_float64", core.NullFloat64Type, false, false, nil, 5, 3).
		Column("raw_bytes", core.RawBytesType, false, false, nil).
		Column("time", core.TimeType, false, false, nil).
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
	)
	err = rows.Scan(&Bool, &Int, &Int8, &Int16, &Int32, &Int64,
		&Uint, &Uint8, &Uint16, &Uint32, &Uint64,
		&F32, &F64, &String, &NullString, &NullInt64, &NullBool, &NullF64, &RawBytes, &Time)
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
		Equal(RawBytes, sql.RawBytes("rawBytes"))

	// bug(caixw) lib/pq 处理 time 时有 bug，更换驱动？
	//
	// lib/pq 对 time.Time 的处理有问题，保存时不会考虑其时区，
	// 直接从字面值当作零时区进行保存。
	// https://github.com/lib/pq/issues/329
	if t.DriverName != "postgres" {
		// 部分数据库可能保存时间，会相差 1 秒。
		tt := math.Abs(float64(Time.Unix() - now.Unix()))
		t.True(tt < 2)
	}
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
		Column("bool", core.BoolType, false, true, false).
		Column("int", core.IntType, false, true, -1).
		Column("int8", core.Int8Type, false, true, -8).
		Column("int16", core.Int16Type, false, true, 0).
		Column("int32", core.Int32Type, false, true, 32).
		Column("int64", core.Int64Type, false, true, -64).
		Column("uint", core.UintType, false, true, 0).
		Column("uint8", core.Uint8Type, false, true, 8).
		Column("uint16", core.Uint16Type, false, true, 16).
		Column("uint32", core.Uint32Type, false, true, 32).
		Column("uint64", core.Uint64Type, false, true, 64).
		Column("float32", core.Float32Type, false, true, -3.2, 5, 3).
		Column("float64", core.Float64Type, false, true, 6.654321, 15, 7).
		Column("string", core.StringType, false, true, "str", 100).
		Column("null_string", core.NullStringType, false, true, "null_str", 100).
		Column("null_int64", core.NullInt64Type, true, true, sql.NullInt64{Int64: 64, Valid: false}).
		Column("null_bool", core.NullBoolType, false, true, sql.NullBool{Bool: true, Valid: true}).
		Column("null_float64", core.NullFloat64Type, true, true, nil, 5, 3).
		Column("raw_bytes", core.RawBytesType, true, false, nil).
		Column("time", core.TimeType, false, true, now).
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
	)
	err = rows.Scan(&Bool, &Int, &Int8, &Int16, &Int32, &Int64,
		&Uint, &Uint8, &Uint16, &Uint32, &Uint64,
		&F32, &F64, &String, &NullString, &NullInt64, &NullBool, &NullF64, &RawBytes, &Time)
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
		Equal(Time.Unix(), now.Unix())
}
