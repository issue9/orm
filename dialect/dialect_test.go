// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"reflect"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/core"
	"github.com/issue9/orm/internal/sqltest"
)

func TestCreatColSQL(t *testing.T) {
	a := assert.New(t)
	dialect := &mysql{}
	buf := core.NewStringBuilder("")
	col := &core.Column{}

	col.Name = "id"
	col.GoType = reflect.TypeOf(1)
	col.Len1 = 5
	createColSQL(dialect, buf, col)
	wont := "{id} BIGINT(5) NOT NULL"
	sqltest.Equal(a, buf.String(), wont)

	buf.Reset()
	col.Len1 = 0
	col.GoType = reflect.TypeOf(int8(1))
	col.HasDefault = true
	col.Default = "1"
	createColSQL(dialect, buf, col)
	wont = "{id} SMALLINT NOT NULL DEFAULT '1'"
	sqltest.Equal(a, buf.String(), wont)

	buf.Reset()
	col.HasDefault = false
	col.Nullable = true
	createColSQL(dialect, buf, col)
	wont = "{id} SMALLINT"
	sqltest.Equal(a, buf.String(), wont)
}

func TestCreatePKSQL(t *testing.T) {
	a := assert.New(t)
	dialect := &mysql{}
	buf := core.NewStringBuilder("")
	col1 := &core.Column{Name: "id"}
	col2 := &core.Column{Name: "username"}
	cols := []*core.Column{col1, col2}

	createPKSQL(dialect, buf, cols, "pkname")
	wont := "CONSTRAINT pkname PRIMARY KEY({id},{username})"
	sqltest.Equal(a, buf.String(), wont)

	buf.Reset()
	createPKSQL(dialect, buf, cols[:1], "pkname")
	wont = "CONSTRAINT pkname PRIMARY KEY({id})"
	sqltest.Equal(a, buf.String(), wont)
}

func TestCreateUniqueSQL(t *testing.T) {
	a := assert.New(t)
	dialect := &mysql{}
	buf := core.NewStringBuilder("")
	col1 := &core.Column{Name: "id"}
	col2 := &core.Column{Name: "username"}
	cols := []*core.Column{col1, col2}

	createUniqueSQL(dialect, buf, cols, "pkname")
	wont := "CONSTRAINT pkname UNIQUE({id},{username})"
	sqltest.Equal(a, buf.String(), wont)

	buf.Reset()
	createUniqueSQL(dialect, buf, cols[:1], "pkname")
	wont = "CONSTRAINT pkname UNIQUE({id})"
	sqltest.Equal(a, buf.String(), wont)
}

func TestCreateFKSQL(t *testing.T) {
	a := assert.New(t)
	dialect := &mysql{}
	buf := core.NewStringBuilder("")
	fk := &core.ForeignKey{
		Col:          &core.Column{Name: "id"},
		RefTableName: "refTable",
		RefColName:   "refCol",
		UpdateRule:   "NO ACTION",
	}

	createFKSQL(dialect, buf, fk, "fkname")
	wont := "CONSTRAINT fkname FOREIGN KEY({id}) REFERENCES refTable({refCol}) ON UPDATE NO ACTION"
	sqltest.Equal(a, buf.String(), wont)
}

func TestCreateCheckSQL(t *testing.T) {
	a := assert.New(t)
	dialect := &mysql{}
	buf := core.NewStringBuilder("")

	createCheckSQL(dialect, buf, "id>5", "chkname")
	wont := "CONSTRAINT chkname CHECK(id>5)"
	sqltest.Equal(a, buf.String(), wont)
}

func TestMysqlLimitSQL(t *testing.T) {
	a := assert.New(t)

	query, ret := mysqlLimitSQL(5, 0)
	a.Equal(ret, []int{5, 0})
	sqltest.Equal(a, query, " LIMIT ? OFFSET ? ")

	query, ret = mysqlLimitSQL(5)
	a.Equal(ret, []int{5})
	sqltest.Equal(a, query, "LIMIT ?")
}

func TestOracleLimitSQL(t *testing.T) {
	a := assert.New(t)

	query, ret := oracleLimitSQL(5, 0)
	a.Equal(ret, []int{0, 5})
	sqltest.Equal(a, query, " OFFSET ? ROWS FETCH NEXT ? ROWS ONLY ")

	query, ret = oracleLimitSQL(5)
	a.Equal(ret, []int{5})
	sqltest.Equal(a, query, "FETCH NEXT ? ROWS ONLY ")
}
