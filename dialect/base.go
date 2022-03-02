// SPDX-License-Identifier: MIT

package dialect

type base struct {
	driverName     string
	dbName         string
	quoteL, quoteR string
}

func newBase(dbName, driverName, quoteLeft, quoteRight string) base {
	return base{
		dbName:     dbName,
		driverName: driverName,
		quoteL:     quoteLeft,
		quoteR:     quoteRight,
	}
}

func (b *base) DBName() string { return b.dbName }

func (b *base) DriverName() string { return b.driverName }

func (b *base) Quotes() (string, string) { return b.quoteL, b.quoteR }
