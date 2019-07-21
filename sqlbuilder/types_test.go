// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"database/sql"
	"math"
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/core"
	"github.com/issue9/orm/v2/internal/test"
	"github.com/issue9/orm/v2/sqlbuilder"
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
			Column("time", core.TimeType, false, false, nil, 0).
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
			// 部分数据库可能保存时间，会相差 1 秒。
			tt := math.Abs(float64(Time.Unix() - now.Unix()))
			t.True(tt < 2)
		}
	})
}
