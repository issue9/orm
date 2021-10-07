// SPDX-License-Identifier: MIT

package dialect

import (
	"strings"

	"github.com/issue9/orm/v4/core"
)

type base struct {
	driverName  string
	dbName      string
	tablePrefix string
	replacer    *strings.Replacer
}

func newBase(dbName, driverName, tablePrefix, quoteLeft, quoteRight string) base {
	return base{
		dbName:      dbName,
		driverName:  driverName,
		tablePrefix: tablePrefix,
		replacer: strings.NewReplacer(string(core.TablePrefix), tablePrefix,
			string(core.QuoteLeft), quoteLeft,
			string(core.QuoteRight), quoteRight),
	}
}

func (b *base) DBName() string { return b.dbName }

func (b *base) DriverName() string { return b.driverName }

func (b *base) TablePrefix() string { return b.tablePrefix }
