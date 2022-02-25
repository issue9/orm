// SPDX-License-Identifier: MIT

package model

import (
	"reflect"
	"testing"
	"time"

	"github.com/issue9/assert/v2"

	"github.com/issue9/orm/v4/core"
	"github.com/issue9/orm/v4/types"
)

func TestNewColumn(t *testing.T) {
	a := assert.New(t, false)

	c, err := newColumn(reflect.StructField{Name: "name", Type: reflect.TypeOf(5)})
	a.NotError(err).NotNil(c).
		Equal(c.Name, "name").Equal(c.GoName, "name").
		Equal(c.GoType, reflect.TypeOf(1)).
		Equal(c.PrimitiveType, core.Int)

	c, err = newColumn(reflect.StructField{Name: "Name", Type: reflect.TypeOf(&last{})})
	a.NotError(err).NotNil(c).
		Equal(c.Name, "Name").Equal(c.GoName, "Name").
		Equal(c.GoType, reflect.TypeOf(last{})).
		Equal(c.PrimitiveType, (&last{}).PrimitiveType())

	c, err = newColumn(reflect.StructField{Name: "Name", Type: reflect.TypeOf([]byte{'1', '2'})})
	a.NotError(err).NotNil(c).
		Equal(c.Name, "Name").Equal(c.GoName, "Name").
		Equal(c.GoType, reflect.TypeOf([]byte{})).
		Equal(c.PrimitiveType, core.Bytes)

	// 自定义类型，但是未实现 PrimitiveTyper 接口
	type T int16
	c, err = newColumn(reflect.StructField{Name: "Name", Type: reflect.TypeOf(T(1))})
	a.NotError(err).NotNil(c).
		Equal(c.Name, "Name").Equal(c.GoName, "Name").
		Equal(c.GoType, reflect.TypeOf(T(1))).
		Equal(c.PrimitiveType, core.Int16)
}

func TestColumn_parseTags(t *testing.T) {
	a := assert.New(t, false)
	m := &core.Model{
		Columns: []*core.Column{},
	}

	// 不存在 struct tag，则以 col.Name 作为键名
	col := &column{
		Column: &core.Column{Name: "xx"},
	}
	a.NotError(col.parseTags(m, ""))
	a.Equal(col.Name, "xx")

	// name 值过多
	col = &column{Column: &core.Column{}}
	a.Error(col.parseTags(m, "name(m1,m2)"))

	// 不存在的属性名称
	col = &column{Column: &core.Column{}}
	a.Error(col.parseTags(m, "not-exists-property(p1)"))
}

func TestColumn_SetLen(t *testing.T) {
	a := assert.New(t, false)
	col := &column{Column: &core.Column{}}

	a.NotError(col.setLen([]string{})).Empty(col.Length)
	a.NotError(col.setLen([]string{"1", "2"})).
		Equal(col.Length[0], 1).
		Equal(col.Length[1], 2)
	a.Error(col.setLen([]string{"1", "2", "3"}))
	a.Error(col.setLen([]string{"1", "one"}))
	a.Error(col.setLen([]string{"one", "one"}))
}

func TestColumn_setNullable(t *testing.T) {
	a := assert.New(t, false)

	col := &column{Column: &core.Column{}}

	a.False(col.Nullable)
	a.NotError(col.setNullable([]string{})).True(col.Nullable)
	a.Error(col.setNullable([]string{"false"})).
		True(col.Nullable)
	a.Error(col.setNullable([]string{"1", "2"}))
	a.Error(col.setNullable([]string{"T1"}))

	ms := NewModels(nil)
	a.NotNil(ms)

	// 将 AI 设置为 nullable
	m, err := ms.New(&User{})
	a.NotError(err).NotNil(m)
	col.AI = true
	a.Error(col.setNullable([]string{"true"}))
}

func TestColumn_setDefault(t *testing.T) {
	a := assert.New(t, false)
	m := core.NewModel(core.Table, "m1", 10)

	// col == int

	col, err := newColumn(reflect.StructField{Name: "def", Type: reflect.TypeOf(1)})
	a.NotError(err).NotNil(col).Equal(col.GoType.Kind(), reflect.Int)
	a.NotError(m.AddColumn(col.Column))

	// 未指定参数
	a.Error(col.setDefault(nil))

	// 过多的参数
	a.Error(col.setDefault([]string{"1", "2"}))

	// 正常
	a.NotError(col.setDefault([]string{"1"}))
	a.True(col.HasDefault).
		Equal(col.Default, 1)

	// 可以是主键的一部分
	m.PrimaryKey = &core.Constraint{Columns: []*core.Column{col.Column, col.Column}, Name: core.PKName(m.Name)}
	a.NotError(col.setDefault([]string{"1"}))
	a.True(col.HasDefault).
		Equal(col.Default, 1)

	// col == []byte

	col, err = newColumn(reflect.StructField{Name: "def", Type: reflect.TypeOf([]byte{'1', '2'})})
	a.NotError(err).NotNil(col).Equal(col.GoType, reflect.TypeOf([]byte{}))

	// 空格
	a.NotError(col.setDefault([]string{""}))
	a.Equal(col.Default, []byte(""))

	a.NotError(col.setDefault([]string{"192.168.1.1,"}))
	a.Equal(col.Default, []byte("192.168.1.1,"))

	// col == last

	col, err = newColumn(reflect.StructField{Name: "def", Type: reflect.TypeOf(&last{})})
	a.NotError(err).NotNil(col).Equal(col.GoType, reflect.TypeOf(last{}))

	// 格式不正确
	a.Error(col.setDefault([]string{"1"}))

	// 格式正确
	now := time.Now()
	f := now.Format(core.TimeFormatLayout)
	a.NotError(col.setDefault([]string{"192.168.1.1," + f}))
	a.Equal(col.Default, &last{
		IP:      "192.168.1.1",
		Created: now.Unix(),
	})

	// col == time.Time

	col, err = newColumn(reflect.StructField{Name: "def", Type: reflect.TypeOf(time.Time{})})
	a.NotError(err).NotNil(col)

	// 格式不正确
	a.Error(col.setDefault([]string{"1"}))

	// 格式正确
	a.NotError(col.setDefault([]string{f}))
	a.Equal(col.Default.(time.Time).Unix(), now.Unix())

	// col == core.Unix

	col, err = newColumn(reflect.StructField{Name: "def", Type: reflect.TypeOf(types.Unix{})})
	a.NotError(err).NotNil(col)

	// 格式不正确
	a.Error(col.setDefault([]string{"xyz"}))

	// 格式正确，但类型被转换成 *core.Unix，而不初始的 core.Unix
	a.NotError(col.setDefault([]string{f}))
	a.Equal(col.Default.(*types.Unix).Time.Unix(), now.Unix())

	// col == &core.Unix

	col, err = newColumn(reflect.StructField{Name: "def", Type: reflect.TypeOf(&types.Unix{})})
	a.NotError(err).NotNil(col)

	// 格式不正确
	a.Error(col.setDefault([]string{"xyz"}))

	// 格式正确
	a.NotError(col.setDefault([]string{f}))
	a.Equal(col.Default.(*types.Unix).Time.Unix(), now.Unix())

	// col == &&core.Unix

	u := &types.Unix{}
	col, err = newColumn(reflect.StructField{Name: "def", Type: reflect.TypeOf(&u)})
	a.NotError(err).NotNil(col)

	// 格式不正确
	a.Error(col.setDefault([]string{"xyz"}))

	// 格式正确
	a.NotError(col.setDefault([]string{f}))
	a.Equal(col.Default.(*types.Unix).Time.Unix(), now.Unix())
}
