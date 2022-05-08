// SPDX-License-Identifier: MIT

package orm_test

import (
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/orm/v5"
	"github.com/issue9/orm/v5/internal/test"
)

var (
	_ orm.Engine = &orm.DB{}
	_ orm.Engine = &orm.Tx{}
)

type beforeObject1 struct {
	ID   int64  `orm:"name(id);ai"`
	Name string `orm:"name(name);len(24)"`
}

type beforeObject2 struct {
	ID   int64  `orm:"name(id);ai"`
	Name string `orm:"name(name);len(24)"`
}

var (
	_ orm.BeforeInserter = &beforeObject1{}
	_ orm.BeforeUpdater  = &beforeObject1{}
)

func (o *beforeObject1) TableName() string { return "objects1" }

func (o *beforeObject1) BeforeInsert() error {
	o.Name = "insert-" + o.Name
	return nil
}

func (o *beforeObject1) BeforeUpdate() error {
	o.Name = "update-" + o.Name
	return nil
}

var (
	_ orm.BeforeInserter = &beforeObject1{}
	_ orm.BeforeUpdater  = &beforeObject1{}
)

func (o *beforeObject2) TableName() string { return "objects2" }

func (o *beforeObject2) BeforeInsert() error {
	o.Name = "insert-" + o.Name
	return nil
}

func (o *beforeObject2) BeforeUpdate() error {
	o.Name = "update-" + o.Name
	return nil
}

func TestBeforeCreateUpdate(t *testing.T) {
	a := assert.New(t, false)
	suite := test.NewSuite(a)

	suite.Run(func(t *test.Driver) {
		// create
		t.NotError(t.DB.Create(&beforeObject1{}))
		defer func() {
			t.NotError(t.DB.Drop(&beforeObject1{}))
		}()

		// insert
		o := &beforeObject1{ID: 1, Name: "name1"}
		_, err := t.DB.Insert(o)
		t.NotError(err)
		o = &beforeObject1{ID: 1}
		found, err := t.DB.Select(o)
		t.NotError(err).True(found)
		t.Equal(o.Name, "insert-name1")

		// update
		o = &beforeObject1{ID: 1, Name: "name11"}
		_, err = t.DB.Update(o)
		t.NotError(err)
		o = &beforeObject1{ID: 1}
		found, err = t.DB.Select(o)
		t.NotError(err).True(found)
		t.Equal(o.Name, "update-name11")
	})
}
