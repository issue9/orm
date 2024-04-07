// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/orm/v6/core"
)

var (
	_ core.PrimitiveTyper = &Rat{}
	_ driver.Valuer       = &Rat{}
	_ sql.Scanner         = &Rat{}

	_ encoding.TextMarshaler   = &Rat{}
	_ encoding.TextUnmarshaler = &Rat{}
)

func TestRational(t *testing.T) {
	a := assert.New(t, false)

	r := Rational(3, 4)
	a.False(r.IsNull())
	val, err := r.Value()
	a.NotError(err).Equal(val, "3/4")

	r = Rat{}
	a.True(r.IsNull())
}

func TestRat_SQL(t *testing.T) {
	a := assert.New(t, false)

	r := &Rat{}
	a.NotError(r.Scan("1/3"))
	a.Equal(r.Rat().String(), "1/3")
	val, err := r.Value()
	a.Equal(val, "1/3").NotError(err)

	r = &Rat{}
	a.NotError(r.Scan(nil))
	a.Nil(r.Rat())
	val, err = r.Value()
	a.Nil(val).NotError(err)

	r = &Rat{}
	a.ErrorIs(r.Scan(1), core.ErrInvalidColumnType)
	val, err = r.Value()
	a.Nil(val).NotError(err)

	r2 := Rational(3, 4)
	a.NotError(r2.Scan("1/3"))
	a.Equal(r2.Rat().String(), "1/3")
	val, err = r2.Value()
	a.Equal(val, "1/3").NotError(err)
}
