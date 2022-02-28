// SPDX-License-Identifier: MIT

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/orm/v5/internal/test"
	"github.com/issue9/orm/v5/sqlbuilder"
)

func TestVersion(t *testing.T) {
	a := assert.New(t, false)

	s := test.NewSuite(a)
	defer s.Close()

	s.ForEach(func(t *test.Driver) {
		ver, err := sqlbuilder.Version(t.DB)
		t.NotError(err).
			NotEmpty(ver)
	})
}
