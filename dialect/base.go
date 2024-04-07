// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package dialect

type base struct {
	driverName     string
	name           string
	quoteL, quoteR byte
}

func newBase(name, driverName string, quoteLeft, quoteRight byte) base {
	return base{
		name:       name,
		driverName: driverName,
		quoteL:     quoteLeft,
		quoteR:     quoteRight,
	}
}

func (b *base) Name() string { return b.name }

func (b *base) DriverName() string { return b.driverName }

func (b *base) Quotes() (byte, byte) { return b.quoteL, b.quoteR }
