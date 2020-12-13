// SPDX-License-Identifier: MIT

package model

import (
	"reflect"
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/orm/v3/core"
)

func TestParseColumn(t *testing.T) {
	a := assert.New(t)
	m := &core.Model{
		Columns: []*core.Column{},
	}

	// 不存在 struct tag，则以 col.Name 作为键名
	col := &core.Column{
		Name: "xx",
	}
	a.NotError(parseColumn(m, col, ""))
	a.Equal(col.Name, "xx")

	// name 值过多
	col = &core.Column{}
	a.Error(parseColumn(m, col, "name(m1,m2)"))

	// 不存在的属性名称
	col = &core.Column{}
	a.Error(parseColumn(m, col, "not-exists-property(p1)"))
}

func TestSetColumnLen(t *testing.T) {
	a := assert.New(t)
	col := &core.Column{}

	a.NotError(setColumnLen(col, []string{})).Empty(col.Length)
	a.NotError(setColumnLen(col, []string{"1", "2"})).
		Equal(col.Length[0], 1).
		Equal(col.Length[1], 2)
	a.Error(setColumnLen(col, []string{"1", "2", "3"}))
	a.Error(setColumnLen(col, []string{"1", "one"}))
	a.Error(setColumnLen(col, []string{"one", "one"}))
}

func TestSetColumnNullable(t *testing.T) {
	a := assert.New(t)

	col := &core.Column{}

	a.False(col.Nullable)
	a.NotError(setColumnNullable(col, []string{})).True(col.Nullable)
	a.Error(setColumnNullable(col, []string{"false"})).
		True(col.Nullable)
	a.Error(setColumnNullable(col, []string{"1", "2"}))
	a.Error(setColumnNullable(col, []string{"T1"}))

	ms := NewModels(nil)
	a.NotNil(ms)

	// 将 AI 设置为 nullable
	m, err := ms.New(&User{})
	a.NotError(err).NotNil(m)
	col.AI = true
	a.Error(setColumnNullable(col, []string{"true"}))
}

func TestSetDefault(t *testing.T) {
	a := assert.New(t)
	m := core.NewModel(core.Table, "m1", 10)

	// col == int

	col, err := core.NewColumnFromGoType(reflect.TypeOf(1))
	a.NotError(err).NotNil(col)
	col.Name = "def"
	a.NotError(m.AddColumn(col))

	// 未指定参数
	a.Error(setDefault(col, nil))

	// 过多的参数
	a.Error(setDefault(col, []string{"1", "2"}))

	// 正常
	a.NotError(setDefault(col, []string{"1"}))
	a.True(col.HasDefault).
		Equal(col.Default, 1)

	// 可以是主键的一部分
	m.PrimaryKey = []*core.Column{col, col}
	a.NotError(setDefault(col, []string{"1"}))
	a.True(col.HasDefault).
		Equal(col.Default, 1)

	// col == last

	col, err = core.NewColumnFromGoType(reflect.TypeOf(&last{}))
	a.NotError(err).NotNil(col)

	// 格式不正确
	a.Error(setDefault(col, []string{"1"}))

	// 格式正确
	now := time.Now()
	f := now.Format(core.TimeFormatLayout)
	a.NotError(setDefault(col, []string{"192.168.1.1," + f}))
	a.Equal(col.Default, &last{
		IP:      "192.168.1.1",
		Created: now.Unix(),
	})

	// col == time.Time

	col, err = core.NewColumnFromGoType(reflect.TypeOf(time.Time{}))
	a.NotError(err).NotNil(col)

	// 格式不正确
	a.Error(setDefault(col, []string{"1"}))

	// 格式正确
	a.NotError(setDefault(col, []string{f}))
	a.Equal(col.Default.(time.Time).Unix(), now.Unix())

	// col == core.Unix

	col, err = core.NewColumnFromGoType(reflect.TypeOf(core.Unix{}))
	a.NotError(err).NotNil(col)

	// 格式不正确
	a.Error(setDefault(col, []string{"xyz"}))

	// 格式正确，但类型被转换成 *core.Unix，而不初始的 core.Unix
	a.NotError(setDefault(col, []string{f}))
	a.Equal(col.Default.(*core.Unix).AsTime().Unix(), now.Unix())

	// col == &core.Unix

	col, err = core.NewColumnFromGoType(reflect.TypeOf(&core.Unix{}))
	a.NotError(err).NotNil(col)

	// 格式不正确
	a.Error(setDefault(col, []string{"xyz"}))

	// 格式正确
	a.NotError(setDefault(col, []string{f}))
	a.Equal(col.Default.(*core.Unix).AsTime().Unix(), now.Unix())
}
