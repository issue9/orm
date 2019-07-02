// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v2/internal/testconfig"
	"github.com/issue9/orm/v2/sqlbuilder"
)

func TestVersion(t *testing.T) {
	a := assert.New(t)
	db := testconfig.NewDB(a)
	defer testconfig.CloseDB(db, a)

	ver, err := sqlbuilder.Version(db, db.Dialect())
	a.NotError(err).NotEmpty(ver)
}
