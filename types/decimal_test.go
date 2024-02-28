// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/orm/v5/core"
)

var (
	_ sql.Scanner         = &Decimal{}
	_ driver.Valuer       = Decimal{}
	_ core.PrimitiveTyper = &Decimal{}

	_ encoding.TextMarshaler   = Decimal{}
	_ json.Marshaler           = Decimal{}
	_ encoding.TextUnmarshaler = &Decimal{}
	_ json.Unmarshaler         = &Decimal{}
)

func TestStringDecimalWithPrecision(t *testing.T) {
	a := assert.New(t, false)

	d, err := StringDecimalWithPrecision("3.222")
	a.NotError(err).Equal(d.Precision, 3).True(d.Valid)

	d, err = StringDecimalWithPrecision(".222")
	a.NotError(err).Equal(d.Precision, 3).True(d.Valid)

	d, err = StringDecimalWithPrecision("222")
	a.NotError(err).Equal(d.Precision, 0).True(d.Valid)

	d, err = StringDecimalWithPrecision("")
	a.Error(err).False(d.Valid)
}

func TestSQL(t *testing.T) {
	a := assert.New(t, false)

	d := FloatDecimal(2.22, 3)
	a.NotError(d.Scan([]byte("3.3333")))
	v, err := d.Value()
	a.NotError(err).Equal(v, "3.333")

	d = FloatDecimal(2.22, 3)
	a.NotError(d.Scan("3"))
	v, err = d.Value()
	a.NotError(err).Equal(v, "3.000")

	d = FloatDecimal(2.22, 3)
	a.Error(d.Scan(""))
}
