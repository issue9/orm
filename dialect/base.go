// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package dialect

type base struct {
	driverName     string
	dbName         string
	quoteL, quoteR byte
}

func newBase(dbName, driverName string, quoteLeft, quoteRight byte) base {
	return base{
		dbName:     dbName,
		driverName: driverName,
		quoteL:     quoteLeft,
		quoteR:     quoteRight,
	}
}

func (b *base) DBName() string { return b.dbName }

func (b *base) DriverName() string { return b.driverName }

func (b *base) Quotes() (byte, byte) { return b.quoteL, b.quoteR }
