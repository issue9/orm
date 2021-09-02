// SPDX-License-Identifier: MIT

package orm_test

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v3"
	"github.com/issue9/orm/v3/core"
	"github.com/issue9/orm/v3/internal/test"
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

func (o *beforeObject1) Meta(m *core.Model) error {
	m.Name = "#objects1"
	return nil
}

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

func (o *beforeObject2) Meta(m *core.Model) error {
	m.Name = "#objects2"
	return nil
}

func (o *beforeObject2) BeforeInsert() error {
	o.Name = "insert-" + o.Name
	return nil
}

func (o *beforeObject2) BeforeUpdate() error {
	o.Name = "update-" + o.Name
	return nil
}

func TestBeforeInsertUpdate(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		// create
		t.NotError(t.DB.Create(&beforeObject1{}))
		defer func() {
			t.NotError(t.DB.Drop(&beforeObject1{}))
		}()

		// insert
		o := &beforeObject1{ID: 1, Name: "name1"}
		t.NotError(t.DB.Insert(o))
		o = &beforeObject1{ID: 1}
		t.NotError(t.DB.Select(o))
		t.Equal(o.Name, "insert-name1")

		// update
		o = &beforeObject1{ID: 1, Name: "name11"}
		t.NotError(t.DB.Update(o))
		o = &beforeObject1{ID: 1}
		t.NotError(t.DB.Select(o))
		t.Equal(o.Name, "update-name11")
	})
}

func TestBeforeInsertUpdate_Mult(t *testing.T) {
	a := assert.New(t)
	suite := test.NewSuite(a)
	defer suite.Close()

	suite.ForEach(func(t *test.Driver) {
		// create
		t.NotError(t.DB.MultCreate(&beforeObject1{}, &beforeObject2{}))
		defer func() {
			t.NotError(t.DB.MultDelete(&beforeObject1{ID: 1}, &beforeObject1{ID: 2}, &beforeObject2{ID: 3}))
			t.NotError(t.DB.MultDrop(&beforeObject1{}, &beforeObject2{}))
		}()

		// insert
		oo := []interface{}{
			&beforeObject1{ID: 1, Name: "name1"},
			&beforeObject1{ID: 2, Name: "name2"},
			&beforeObject2{ID: 3, Name: "name3"},
		}
		t.NotError(t.DB.MultInsert(oo...))
		oo = []interface{}{
			&beforeObject1{ID: 1},
			&beforeObject1{ID: 2},
			&beforeObject2{ID: 3},
		}
		t.NotError(t.DB.MultSelect(oo...))
		o0, ok := oo[0].(*beforeObject1)
		t.True(ok).Equal(o0.Name, "insert-name1")
		o1, ok := oo[1].(*beforeObject1)
		t.True(ok).Equal(o1.Name, "insert-name2")
		o2, ok := oo[2].(*beforeObject2)
		t.True(ok).Equal(o2.Name, "insert-name3")

		// update
		oo = []interface{}{
			&beforeObject1{ID: 1, Name: "name11"},
			&beforeObject1{ID: 2, Name: "name22"},
			&beforeObject2{ID: 3, Name: "name33"},
		}
		t.NotError(t.DB.MultUpdate(oo...))
		oo = []interface{}{
			&beforeObject1{ID: 1},
			&beforeObject1{ID: 2},
			&beforeObject2{ID: 3},
		}
		t.NotError(t.DB.MultSelect(oo...))
		o0, ok = oo[0].(*beforeObject1)
		t.True(ok).Equal(o0.Name, "update-name11")
		o1, ok = oo[1].(*beforeObject1)
		t.True(ok).Equal(o1.Name, "update-name22")
		o2, ok = oo[2].(*beforeObject2)
		t.True(ok).Equal(o2.Name, "update-name33")
	})
}
