// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/v2"
	"github.com/issue9/orm/v2/sqlbuilder"

	"github.com/issue9/orm/v2/internal/testconfig"
)

type user struct {
	ID   int64  `orm:"name(id);ai"`
	Name string `orm:"name(name);len(20)"`
}

func createTable(a *assert.Assertion) *orm.DB {
	db := testconfig.NewDB(a)
	a.NotError(db.Create(&user{}))
	return db
}

func TestSQLBuilder(t *testing.T) {
	a := assert.New(t)

	b := sqlbuilder.New("")
	b.WriteByte('1')
	b.WriteString("23")

	a.Equal("123", b.String())
	a.Equal(3, b.Len())

	b.Reset()
	a.Equal(b.String(), "")
	a.Equal(b.Len(), 0)

	b.WriteByte('3').WriteString("21")
	a.Equal(b.String(), "321")

	b.TruncateLast(1)
	a.Equal(b.String(), "32").Equal(2, b.Len())
}
