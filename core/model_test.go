// SPDX-License-Identifier: MIT

package core

import (
	"testing"

	"github.com/issue9/assert/v2"
)

func TestNewModel(t *testing.T) {
	a := assert.New(t, false)

	m := NewModel(Table, "m1", 10)
	a.NotNil(m)
	a.Equal(m.Name, "m1")
}

func TestModel_SetAutoIncrement(t *testing.T) {
	a := assert.New(t, false)
	m := NewModel(Table, "m1", 10)
	a.NotNil(m)

	// 正常添加
	ai, err := NewColumn(Int)
	a.NotError(err).NotNil(ai)
	ai.Name = "ai"
	a.NotError(m.AddColumns(ai))
	a.NotError(m.SetAutoIncrement(ai))
	a.True(ai.AI)
	a.Equal(m.AutoIncrement.Columns[0], ai)

	// 同类型也不可以多次添加
	a.Error(m.SetAutoIncrement(ai))
	a.True(ai.AI)
	a.Equal(m.AutoIncrement.Columns[0], ai)

	// 已有自增列
	ai2, err := NewColumn(Int64)
	a.NotError(err).NotNil(ai2)
	ai2.Name = "ai2"
	a.NotError(m.AddColumns(ai2))
	a.Error(m.SetAutoIncrement(ai2))

	// 类型错误
	m.Reset()
	ai2, err = NewColumn(String)
	a.NotError(err).NotNil(ai2)
	ai2.Name = "ai2"
	a.NotError(m.AddColumns(ai2))
	a.ErrorString(m.SetAutoIncrement(ai2), "必须是数值类型")

	// 存在主键
	m.Reset()
	a.NotError(m.AddColumns(ai2))
	a.NotError(m.AddPrimaryKey(ai2)) // 主键
	a.ErrorIs(m.SetAutoIncrement(ai), ErrAutoIncrementPrimaryKeyConflict)

	// ai 不存在
	m.Reset()
	a.Error(m.SetAutoIncrement(ai))
}

func TestModel_AddPrimaryKey(t *testing.T) {
	a := assert.New(t, false)
	m := NewModel(Table, "m1", 10)
	a.NotNil(m)

	ai, err := NewColumn(Int)
	a.NotError(err).NotNil(ai)
	ai.Name = "ai"
	ai.AI = true
	a.NotError(m.AddColumns(ai))

	// 与自增冲突
	pk, err := NewColumn(Int)
	a.NotError(err).NotNil(pk)
	pk.Name = "pk"
	a.Error(m.AddPrimaryKey(pk))

	// 列不存在
	m.Reset()
	a.Error(m.AddPrimaryKey(pk))

	// 正常添加
	m.Reset()
	a.NotError(m.AddColumns(pk))
	a.NotError(m.AddPrimaryKey(pk))

	// 多列主键约束
	pk2, err := NewColumn(String)
	a.NotError(err).NotNil(pk2)
	pk2.Name = "pk2"
	a.NotError(m.AddColumns(pk2))
	a.NotError(m.AddPrimaryKey(pk2))

	a.Equal(m.PrimaryKey.Columns, []*Column{pk, pk2})
}

func TestModel_SetOCC(t *testing.T) {
	a := assert.New(t, false)
	m := NewModel(Table, "m1", 10)
	a.NotNil(m)

	// AI 作为 OCC
	ai, err := NewColumn(Int)
	a.NotError(err).NotNil(ai)
	ai.Name = "ai"
	ai.AI = true
	a.Error(m.SetOCC(ai))

	// NULL 在 OCC 列上
	ai.AI = false
	ai.Nullable = true
	a.Error(m.SetOCC(ai))

	// 列不存在
	ai, err = NewColumn(Int)
	a.NotError(err).NotNil(ai)
	ai.Name = "ai"
	a.Error(m.SetOCC(ai))

	// 类型错误
	ai, err = NewColumn(String)
	a.NotError(err).NotNil(ai)
	ai.Name = "ai"
	a.NotError(m.AddColumn(ai))
	a.ErrorString(m.SetOCC(ai), "必须是数值类型")

	// 正常
	ai, err = NewColumn(Int)
	a.NotError(err).NotNil(ai)
	ai.Name = "ai2"
	a.NotError(m.AddColumn(ai))
	a.NotError(m.SetOCC(ai))

	// 多次添加
	col2, err := NewColumn(Int)
	a.NotError(err).NotNil(col2)
	col2.Name = "col2"
	a.NotError(m.AddColumn(col2))
	a.Error(m.SetOCC(col2))
}

func TestModel_AddIndex(t *testing.T) {
	a := assert.New(t, false)
	m := NewModel(Table, "m1", 10)
	a.NotNil(m)

	col, err := NewColumn(Int)
	a.NotError(err).NotNil(col)
	col.Name = "col"

	col2, err := NewColumn(Int)
	a.NotError(err).NotNil(col2)
	col2.Name = "col2"

	a.NotError(m.AddColumns(col, col2))

	a.NotError(m.AddIndex(IndexDefault, "index_0", col))
	a.NotError(m.AddIndex(IndexDefault, "index_0", col2))
	a.NotError(m.AddIndex(IndexDefault, "index_1", col2))

	a.Equal(2, len(m.Indexes))
	i, found := m.Index("index_0")
	a.True(found).Equal(i.Columns, []*Column{col, col2})
	i, found = m.Index("index_1")
	a.True(found).Equal(i.Columns, []*Column{col2})

	// Unique
	a.NotError(m.AddIndex(IndexUnique, "unique_0", col))
	a.Equal(2, len(m.Indexes)).
		Equal(1, len(m.Uniques))

	// unique 列不存在
	col3, err := NewColumn(Int)
	a.NotError(err).NotNil(col3)
	col3.Name = "col3"
	a.Error(m.AddIndex(IndexUnique, "unique_1", col3))

	// index 列不存在
	a.Error(m.AddIndex(IndexDefault, "index_2", col3))
}

func TestModel_NewCheck(t *testing.T) {
	a := assert.New(t, false)
	m := NewModel(Table, "m1", 10)
	a.NotNil(m)

	a.NotError(m.NewCheck("id_great_0", "id>0"))
	a.Error(m.NewCheck("id_great_0", "id>0"))
}

func TestModel_NewForeignKey(t *testing.T) {
	a := assert.New(t, false)
	m := NewModel(Table, "m1", 10)
	a.NotNil(m)

	// 空的 ForeignKey
	a.Error(m.NewForeignKey(&ForeignKey{}))

	fkCol, err := NewColumn(Uint8)
	a.NotError(err).NotNil(fkCol)
	fkCol.Name = "fk_col"

	// 列不存在于当前模型
	a.Error(m.NewForeignKey(&ForeignKey{Name: "fk_0", Column: fkCol, RefTableName: "tbl", RefColName: "col"}))

	// 正常
	a.NotError(m.AddColumn(fkCol))
	a.NotError(m.NewForeignKey(&ForeignKey{Name: "fk_0", Column: fkCol, RefTableName: "tbl", RefColName: "col"}))

	// 重复的约束名
	a.Error(m.NewForeignKey(&ForeignKey{Name: "fk_0", Column: fkCol, RefTableName: "tbl", RefColName: "col"}))
}

func TestModel_Sanitize(t *testing.T) {
	a := assert.New(t, false)

	ai, err := NewColumn(Int)
	a.NotError(err).NotNil(ai)
	ai.AI = true
	ai.Name = "ai"

	pk1, err := NewColumn(Int)
	a.NotError(err).NotNil(pk1)
	pk1.Default = 1
	pk1.HasDefault = true
	pk1.Name = "pk1"

	pk2, err := NewColumn(Int8)
	a.NotError(err).NotNil(pk2)
	pk2.Name = "pk2"

	nullable, err := NewColumn(Int8)
	a.NotError(err).NotNil(nullable)
	nullable.Nullable = true
	nullable.Name = "nullable"

	def, err := NewColumn(Int)
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

	m.Reset()
	m.Name = "m1"
	m.Type = Table
	a.NotError(m.AddColumn(ai))
	a.NotError(m.AddColumns(nullable, def))
	a.NotError(m.AddIndex(IndexDefault, "index_0", nullable))
	a.NotError(m.NewForeignKey(&ForeignKey{Name: "fk_0", Column: def, RefTableName: "tbl", RefColName: "col"}))
	a.NotError(m.NewCheck("id_great_0", "{id}>0"))
	a.NotError(m.Sanitize())

	// 约束重名
	a.NotError(m.AddUnique("fk_0", nullable))
	a.Error(m.Sanitize())
}

func TestConstraint_sanitize(t *testing.T) {
	a := assert.New(t, false)

	var c *Constraint
	a.NotError(c.sanitize())

	c = &Constraint{}
	a.ErrorString(c.sanitize(), "未指定约束名")

	c = &Constraint{Name: "abc"}
	a.ErrorString(c.sanitize(), "并未指定列")
}

func TestForeignKey_sanitize(t *testing.T) {
	a := assert.New(t, false)

	var fk *ForeignKey
	a.NotError(fk.sanitize())

	fk = &ForeignKey{}
	a.ErrorString(fk.sanitize(), "未指定外键的约束名")

	fk = &ForeignKey{Name: "fk"}
	a.ErrorString(fk.sanitize(), "并未指定列")

	fk = &ForeignKey{Name: "fk", Column: &Column{}}
	a.ErrorString(fk.sanitize(), "缺少必要的字段 ref")
}
