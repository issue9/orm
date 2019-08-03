// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package core

import (
	"testing"

	"github.com/issue9/assert"
)

func (m *Model) reset() {
	m.Type = Table
	m.Columns = m.Columns[:0]
	m.Indexes = map[string][]*Column{}
	m.Uniques = map[string][]*Column{}
	m.ForeignKeys = map[string]*ForeignKey{}
	m.PrimaryKey = []*Column{}
	m.Checks = map[string]string{}
	m.Meta = map[string][]string{}
	m.AutoIncrement = nil
}

func TestNewModel(t *testing.T) {
	a := assert.New(t)

	m := NewModel(Table, "m1", 10)
	a.NotNil(m)
	a.Equal(m.FullName, "#m1")

	a.Equal(m.AIName(), "#m1_ai")
	a.Equal(m.PKName(), "#m1_pk")
}

func TestModel_AddColumns(t *testing.T) {
	a := assert.New(t)
	m := NewModel(Table, "m1", 10)
	a.NotNil(m)

	ai, err := NewColumnFromGoType(IntType)
	a.NotError(err).NotNil(ai)
	ai.AI = true
	a.Error(m.AddColumns(ai)) // 没有名称

	col, err := NewColumnFromGoType(IntType)
	a.NotError(err).NotNil(col)

	// 同名
	ai.Name = "ai"
	col.Name = "ai"
	a.Error(m.AddColumns(ai, col))

	// 正常
	m.reset()
	col.Name = "col"
	a.NotError(m.AddColumns(ai, col))
}

func TestModel_SetAutoIncrement(t *testing.T) {
	a := assert.New(t)
	m := NewModel(Table, "m1", 10)
	a.NotNil(m)

	// 正常添加
	ai, err := NewColumnFromGoType(IntType)
	a.NotError(err).NotNil(ai)
	ai.Name = "ai"
	a.NotError(m.AddColumns(ai))
	a.NotError(m.SetAutoIncrement(ai))
	a.True(ai.AI)
	a.Equal(m.AutoIncrement, ai)

	// 同类型可以多次添加
	a.NotError(m.SetAutoIncrement(ai))
	a.True(ai.AI)
	a.Equal(m.AutoIncrement, ai)

	// 已有自增列
	ai2, err := NewColumnFromGoType(Int64Type)
	a.NotError(err).NotNil(ai2)
	ai2.Name = "ai2"
	a.NotError(m.AddColumns(ai2))
	a.Error(m.SetAutoIncrement(ai2))

	// 类型错误
	m.reset()
	ai2, err = NewColumnFromGoType(StringType)
	a.NotError(err).NotNil(ai2)
	ai2.Name = "ai2"
	a.NotError(m.AddColumns(ai2))
	a.ErrorType(m.SetAutoIncrement(ai2), ErrColumnTypeError)

	// 存在主键
	m.reset()
	a.NotError(m.AddColumns(ai2))
	a.NotError(m.AddPrimaryKey(ai2)) // 主键
	a.ErrorType(m.SetAutoIncrement(ai), ErrAutoIncrementPrimaryKeyConflict)

	// ai 不存在
	m.reset()
	a.Error(m.SetAutoIncrement(ai))
}

func TestModel_AddPrimaryKey(t *testing.T) {
	a := assert.New(t)
	m := NewModel(Table, "m1", 10)
	a.NotNil(m)

	ai, err := NewColumnFromGoType(IntType)
	a.NotError(err).NotNil(ai)
	ai.Name = "ai"
	ai.AI = true
	a.NotError(m.AddColumns(ai))

	// 与自增冲突
	pk, err := NewColumnFromGoType(IntType)
	a.NotError(err).NotNil(pk)
	pk.Name = "pk"
	a.Error(m.AddPrimaryKey(pk))

	// 列不存在
	m.reset()
	a.Error(m.AddPrimaryKey(pk))

	// 正常添加
	m.reset()
	a.NotError(m.AddColumns(pk))
	a.NotError(m.AddPrimaryKey(pk))

	// 多列主键约束
	pk2, err := NewColumnFromGoType(StringType)
	a.NotError(err).NotNil(pk2)
	pk2.Name = "pk2"
	a.NotError(m.AddColumns(pk2))
	a.NotError(m.AddPrimaryKey(pk2))

	a.Equal(m.PrimaryKey, []*Column{pk, pk2})
}

func TestModel_SetOCC(t *testing.T) {
	a := assert.New(t)
	m := NewModel(Table, "m1", 10)
	a.NotNil(m)

	// AI 作为 OCC
	ai, err := NewColumnFromGoType(IntType)
	a.NotError(err).NotNil(ai)
	ai.Name = "ai"
	ai.AI = true
	a.Error(m.SetOCC(ai))

	// NULL 在 OCC 列上
	ai.AI = false
	ai.Nullable = true
	a.Error(m.SetOCC(ai))

	// 列不存在
	ai, err = NewColumnFromGoType(IntType)
	a.NotError(err).NotNil(ai)
	ai.Name = "ai"
	a.Error(m.SetOCC(ai))

	// 类型错误
	ai, err = NewColumnFromGoType(StringType)
	a.NotError(err).NotNil(ai)
	ai.Name = "ai"
	a.NotError(m.AddColumn(ai))
	a.ErrorType(m.SetOCC(ai), ErrColumnTypeError)

	// 正常
	ai, err = NewColumnFromGoType(IntType)
	a.NotError(err).NotNil(ai)
	ai.Name = "ai2"
	a.NotError(m.AddColumn(ai))
	a.NotError(m.SetOCC(ai))

	// 多次添加
	col2, err := NewColumnFromGoType(IntType)
	a.NotError(err).NotNil(col2)
	col2.Name = "col2"
	a.NotError(m.AddColumn(col2))
	a.Error(m.SetOCC(col2))
}

func TestModel_AddIndex(t *testing.T) {
	a := assert.New(t)
	m := NewModel(Table, "m1", 10)
	a.NotNil(m)

	col, err := NewColumnFromGoType(IntType)
	a.NotError(err).NotNil(col)
	col.Name = "col"

	col2, err := NewColumnFromGoType(IntType)
	a.NotError(err).NotNil(col2)
	col2.Name = "col2"

	a.NotError(m.AddColumns(col, col2))

	a.NotError(m.AddIndex(IndexDefault, "index_0", col))
	a.NotError(m.AddIndex(IndexDefault, "index_0", col2))
	a.NotError(m.AddIndex(IndexDefault, "index_1", col2))

	a.Equal(2, len(m.Indexes)).
		Equal(m.Indexes["index_0"], []*Column{col, col2}).
		Equal(m.Indexes["index_1"], []*Column{col2})

	// Unique
	a.NotError(m.AddIndex(IndexUnique, "unique_0", col))
	a.Equal(2, len(m.Indexes)).
		Equal(1, len(m.Uniques))

	// unique 列不存在
	col3, err := NewColumnFromGoType(IntType)
	a.NotError(err).NotNil(col3)
	col3.Name = "col3"
	a.Error(m.AddIndex(IndexUnique, "unique_1", col3))

	// index 列不存在
	a.Error(m.AddIndex(IndexDefault, "index_2", col3))
}

func TestModel_NewCheck(t *testing.T) {
	a := assert.New(t)
	m := NewModel(Table, "m1", 10)
	a.NotNil(m)

	a.NotError(m.NewCheck("id_great_0", "id>0"))
	a.Error(m.NewCheck("id_great_0", "id>0"))
}

func TestModel_NewForeignKey(t *testing.T) {
	a := assert.New(t)
	m := NewModel(Table, "m1", 10)
	a.NotNil(m)

	// 空的 ForeignKey
	a.Error(m.NewForeignKey("fk_0", &ForeignKey{}))

	fkCol, err := NewColumnFromGoType(Uint8Type)
	a.NotError(err).NotNil(fkCol)
	fkCol.Name = "fk_col"

	// 列不存在于当前模型
	a.Error(m.NewForeignKey("fk_0", &ForeignKey{Column: fkCol, RefTableName: "tbl", RefColName: "col"}))

	// 正常
	a.NotError(m.AddColumn(fkCol))
	a.NotError(m.NewForeignKey("fk_0", &ForeignKey{Column: fkCol, RefTableName: "tbl", RefColName: "col"}))

	// 重复的约束名
	a.Error(m.NewForeignKey("fk_0", &ForeignKey{Column: fkCol, RefTableName: "tbl", RefColName: "col"}))
}

func TestModel_Sanitize(t *testing.T) {
	a := assert.New(t)

	ai, err := NewColumnFromGoType(IntType)
	a.NotError(err).NotNil(ai)
	ai.AI = true
	ai.Name = "ai"

	pk1, err := NewColumnFromGoType(IntType)
	a.NotError(err).NotNil(pk1)
	pk1.SetDefault(1)
	pk1.Name = "pk1"

	pk2, err := NewColumnFromGoType(IntType)
	a.NotError(err).NotNil(pk2)
	pk2.Name = "pk2"

	nullable, err := NewColumnFromGoType(IntType)
	a.NotError(err).NotNil(nullable)
	nullable.Nullable = true
	nullable.Name = "nullable"

	def, err := NewColumnFromGoType(IntType)
	a.NotError(err).NotNil(def)
	def.HasDefault = true
	def.Default = 1
	def.Name = "def"

	m := NewModel(Table, "m1", 10)
	a.NotNil(m)
	a.NotError(m.AddColumns(pk1, pk2, nullable, def))
	a.NotError(m.Sanitize())

	// 单列主键，且带默认值
	a.NotError(m.AddPrimaryKey(pk1))
	a.Error(m.Sanitize())

	// 多列主键，要以带默认值值
	a.NotError(m.AddPrimaryKey(pk2))
	a.NotError(m.Sanitize())

	m.reset()
	a.NotError(m.AddColumn(ai))
	a.NotError(m.AddColumns(nullable, def))
	a.NotError(m.AddIndex(IndexDefault, "index_0", nullable))
	a.NotError(m.NewForeignKey("fk_0", &ForeignKey{Column: def, RefTableName: "tbl", RefColName: "col"}))
	a.NotError(m.NewCheck("id_great_0", "{id}>0"))
	a.NotError(m.Sanitize())

	// 约束重名
	a.NotError(m.AddUnique("fk_0", nullable))
	a.Error(m.Sanitize())
}
