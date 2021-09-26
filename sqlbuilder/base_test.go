// SPDX-License-Identifier: MIT

package sqlbuilder_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v4/core"
	"github.com/issue9/orm/v4/internal/test"
	"github.com/issue9/orm/v4/sqlbuilder"
)

func TestMergeDDL(t *testing.T) {
	a := assert.New(t)

	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		initDB(t)
		defer clearDB(t)

		ddl1 := sqlbuilder.CreateIndex(t.DB).
			Table("users").
			Name("index_key").
			Columns("id", "name")

		ddl2 := sqlbuilder.AddColumn(t.DB).
			Table("users").
			Column("id", core.Int, true, false, false, nil)

		ddl3 := sqlbuilder.AddColumn(t.DB).
			Table("users").
			Column("name", core.String, false, true, false, nil)

		ddl := sqlbuilder.MergeDDL(ddl1, ddl2, ddl3)
		queries, err := ddl.DDLSQL()
		a.NotError(err).NotEmpty(queries)
	})
}
