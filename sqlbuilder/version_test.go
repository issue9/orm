// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v3/internal/test"
	"github.com/issue9/orm/v3/sqlbuilder"
)

func TestVersion(t *testing.T) {
	a := assert.New(t)

	s := test.NewSuite(a)
	defer s.Close()

	s.ForEach(func(t *test.Driver) {
		ver, err := sqlbuilder.Version(t.DB)
		t.NotError(err).
			NotEmpty(ver)
	})
}
