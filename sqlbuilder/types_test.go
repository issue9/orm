// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/test"
	"github.com/issue9/orm/v2/sqlbuilder"
)

var (
	_ sqlbuilder.Engine = &sql.DB{}
	_ sqlbuilder.Engine = &sql.Tx{}
)

func quoteColumns(stmt *sqlbuilder.SelectStmt, col ...string) {
	for _, c := range col {
		stmt.Column("{" + c + "}")
	}
}

func TestTypes(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Test) {
		e := t.DB.DB
		d := t.DB.Dialect()
		tableName := "test_type_read_write"
		now := time.Now()

		creator := sqlbuilder.CreateTable(e, d).
			Column("bool", sqlbuilder.BoolType, false, false, nil).
			Column("int", sqlbuilder.IntType, false, false, nil).
			Column("int8", sqlbuilder.Int8Type, false, false, nil).
			Column("int16", sqlbuilder.Int16Type, false, false, nil).
			Column("int32", sqlbuilder.Int32Type, false, false, nil).
			Column("int64", sqlbuilder.Int64Type, false, false, nil).
			Column("uint", sqlbuilder.UintType, false, false, nil).
			Column("uint8", sqlbuilder.Uint8Type, false, false, nil).
			Column("uint16", sqlbuilder.Uint16Type, false, false, nil).
			Column("uint32", sqlbuilder.Uint32Type, false, false, nil).
			Column("uint64", sqlbuilder.Uint64Type, false, false, nil).
			Column("float32", sqlbuilder.Float32Type, false, false, nil, 5, 3).
			Column("float64", sqlbuilder.Float64Type, false, false, nil, 5, 3).
			Column("string", sqlbuilder.StringType, false, false, nil, 100).
			Column("null_string", sqlbuilder.NullStringType, false, false, nil, 100).
			Column("null_int64", sqlbuilder.NullInt64Type, false, false, nil).
			Column("null_bool", sqlbuilder.NullBoolType, false, false, nil).
			Column("null_float64", sqlbuilder.NullFloat64Type, false, false, nil, 5, 3).
			Column("raw_bytes", sqlbuilder.RawBytesType, false, false, nil).
			Column("time", sqlbuilder.TimeType, false, false, nil, 0).
			Table(tableName)
		t.NotError(creator.Exec())
		defer func() {
			t.NotError(sqlbuilder.DropTable(e, d).Table(tableName).Exec())
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

		r, err := sqlbuilder.Insert(e, d).
			Table(tableName).
			Columns(cols...).
			Values(vals...).
			Exec()
		t.NotError(err).NotNil(r)

		selStmt := sqlbuilder.Select(e, d).
			From(tableName)
		quoteColumns(selStmt, cols...)
		rows, err := selStmt.Query()
		t.NotError(err).NotNil(rows)
		defer func() {
			t.NotError(rows.Close())
		}()

		a.True(rows.Next())
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
		a.NotError(err)
		a.True(Bool).
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
			t.Equal(Time.Unix(), now.Unix())
		}
	})
}
