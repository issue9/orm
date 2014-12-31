// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package dialect

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/core"
)

var style = assert.StyleTrim | assert.StyleSpace | assert.StyleCase

func TestCreatColSQL(t *testing.T) {
	a := assert.New(t)
	dialect := &Mysql{}
	buf := bytes.NewBufferString("")
	col := &core.Column{}

	col.Name = "id"
	col.GoType = reflect.TypeOf(1)
	col.Len1 = 5
	createColSQL(dialect, buf, col)
	wont := "id BIGINT(5) NOT NULL"
	a.StringEqual(buf.String(), wont, style)

	buf.Reset()
	col.Len1 = 0
	col.GoType = reflect.TypeOf(int8(1))
	col.HasDefault = true
	col.Default = "1"
	createColSQL(dialect, buf, col)
	wont = "id SMALLINT NOT NULL DEFAULT '1'"
	a.StringEqual(buf.String(), wont, style)

	buf.Reset()
	col.HasDefault = false
	col.Nullable = true
	createColSQL(dialect, buf, col)
	wont = "id SMALLINT NULL"
}

func TestCreatePKSQL(t *testing.T) {
	a := assert.New(t)
	dialect := &Mysql{}
	buf := bytes.NewBufferString("")
	col1 := &core.Column{Name: "id"}
	col2 := &core.Column{Name: "username"}
	cols := []*core.Column{col1, col2}

	createPKSQL(dialect, buf, cols, "pkname")
	wont := "CONSTRAINT pkname PRIMARY KEY(id,username)"
	a.StringEqual(buf.String(), wont, style)

	buf.Reset()
	createPKSQL(dialect, buf, cols[:1], "pkname")
	wont = "CONSTRAINT pkname PRIMARY KEY(id)"
	a.StringEqual(buf.String(), wont, style)
}

func TestCreateUniqueSQL(t *testing.T) {
	a := assert.New(t)
	dialect := &Mysql{}
	buf := bytes.NewBufferString("")
	col1 := &core.Column{Name: "id"}
	col2 := &core.Column{Name: "username"}
	cols := []*core.Column{col1, col2}

	createUniqueSQL(dialect, buf, cols, "pkname")
	wont := "CONSTRAINT pkname UNIQUE(id,username)"
	a.StringEqual(buf.String(), wont, style)

	buf.Reset()
	createUniqueSQL(dialect, buf, cols[:1], "pkname")
	wont = "CONSTRAINT pkname UNIQUE(id)"
	a.StringEqual(buf.String(), wont, style)
}

func TestCreateFKSQL(t *testing.T) {
	a := assert.New(t)
	dialect := &Mysql{}
	buf := bytes.NewBufferString("")
	fk := &core.ForeignKey{
		Col:          &core.Column{Name: "id"},
		RefTableName: "refTable",
		RefColName:   "refCol",
		UpdateRule:   "NO ACTION",
	}

	createFKSQL(dialect, buf, fk, "fkname")
	wont := "CONSTRAINT fkname FOREIGN KEY(id) REFERENCES refTable(refCol) ON UPDATE NO ACTION"
	a.StringEqual(buf.String(), wont, style)
}

func TestCreateCheckSQL(t *testing.T) {
	a := assert.New(t)
	dialect := &Mysql{}
	buf := bytes.NewBufferString("")

	createCheckSQL(dialect, buf, "id>5", "chkname")
	wont := "CONSTRAINT chkname CHECK(id>5)"
	a.StringEqual(wont, buf.String(), style)
}
