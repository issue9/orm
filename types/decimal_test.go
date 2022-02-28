// SPDX-License-Identifier: MIT

package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/shopspring/decimal"

	"github.com/issue9/orm/v5/core"
)

var (
	_ sql.Scanner         = &Decimal{}
	_ driver.Valuer       = Decimal{}
	_ core.DefaultParser  = &Decimal{}
	_ core.PrimitiveTyper = &Decimal{}

	_ encoding.TextMarshaler   = Decimal{}
	_ json.Marshaler           = Decimal{}
	_ encoding.TextUnmarshaler = &Decimal{}
	_ json.Unmarshaler         = &Decimal{}
)

func TestStringDecimalWithPrecision(t *testing.T) {
	a := assert.New(t, false)

	d, err := StringDecimalWithPrecision("3.222")
	a.NotError(err).Equal(d.Precision, 3).False(d.IsNull)

	d, err = StringDecimalWithPrecision(".222")
	a.NotError(err).Equal(d.Precision, 3).False(d.IsNull)

	d, err = StringDecimalWithPrecision("222")
	a.NotError(err).Equal(d.Precision, 0).False(d.IsNull)

	d, err = StringDecimalWithPrecision("")
	a.Error(err).False(d.IsNull)
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

func TestParseDefault(t *testing.T) {
	a := assert.New(t, false)

	d := Decimal{Decimal: decimal.New(1, 2), Precision: 1}
	a.NotError(d.ParseDefault("3.333"))
	val, err := d.MarshalText()
	a.NotError(err).Equal(string(val), "3.3")

	dd := &Decimal{Decimal: decimal.New(1, 2), Precision: 1}
	a.NotError(dd.ParseDefault("3.333"))
	val, err = dd.MarshalText()
	a.NotError(err).Equal(string(val), "3.3")
}
