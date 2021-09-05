// SPDX-License-Identifier: MIT

package types

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/shopspring/decimal"
)

func TestStringDecimalWithPercision(t *testing.T) {
	a := assert.New(t)

	d, err := StringDecimalWithPercision("3.222")
	a.NotError(err).Equal(d.Precision, 3).True(d.Valid)

	d, err = StringDecimalWithPercision(".222")
	a.NotError(err).Equal(d.Precision, 3).True(d.Valid)

	d, err = StringDecimalWithPercision("222")
	a.NotError(err).Equal(d.Precision, 0).True(d.Valid)

	d, err = StringDecimalWithPercision("")
	a.Error(err).False(d.Valid)
}

func TestSQL(t *testing.T) {
	a := assert.New(t)

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
	a := assert.New(t)

	d := Decimal{Decimal: decimal.New(1, 2), Valid: true, Precision: 1}
	d.ParseDefault("3.333")
	val, err := d.MarshalText()
	a.NotError(err).Equal(string(val), "3.3")

	dd := &Decimal{Decimal: decimal.New(1, 2), Valid: true, Precision: 1}
	dd.ParseDefault("3.333")
	val, err = dd.MarshalText()
	a.NotError(err).Equal(string(val), "3.3")
}
