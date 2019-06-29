// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"reflect"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/sqlbuilder"
)

func TestColumn(t *testing.T) {
	a := assert.New(t)

	db := initDB(a)
	defer clearDB(a, db)

	r, err := sqlbuilder.AddColumn(db, db.Dialect()).
		Table("#user").
		Column("col1", reflect.TypeOf(1), true, true, nil).
		Exec()
	a.NotError(err).NotNil(r)

	r, err = sqlbuilder.DropColumn(db).
		Table("#user").
		Column("col1").
		Exec()
	a.NotError(err).NotNil(r)
}
