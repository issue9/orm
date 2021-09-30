// SPDX-License-Identifier: MIT

package core

import (
	"errors"
	"fmt"
)

// ErrInvalidColumnType 无效的列类型
//
// 作为列类型，该数据类型必须是可序列化的。
// 像 reflect.Func 和 reflect.Chan 等都将返回该错误。
var ErrInvalidColumnType = errors.New("无效的列类型")

// DefaultParser 提供了 ParseDefault 函数
//
// 在 struct tag 中可以通过 default 指定默认值，
// 该值的表示可能与数据库中的表示不尽相同，
// 所以自定义的数据类型，需要实现该接口，以便能正确转换成该类型的值。
//
// 如果用户不提供该接口实现，那么默认情况下，
// 系统会采用 github.com/issue9/conv.Value() 函数作默认转换。
type DefaultParser interface {
	// ParseDefault 将默认值从字符串解析成 t 类型的值
	ParseDefault(v string) error
}

// Column 列结构
type Column struct {
	Name       string // 数据库的字段名
	AI         bool
	Nullable   bool
	HasDefault bool
	Default    interface{}
	Length     []int

	PrimitiveType PrimitiveType
	GoName        string // Go 中的字段名
}

// NewColumn 从 Go 类型中生成 Column
func NewColumn(p PrimitiveType) (*Column, error) {
	if p <= Auto || p >= maxPrimitiveType {
		return nil, ErrInvalidColumnType
	}

	return &Column{
		PrimitiveType: p,
	}, nil
}

// Clone 复制 Column
func (c *Column) Clone() *Column {
	cc := &Column{}
	*cc = *c

	return cc
}

// Check 检测 Column 内容是否合法
func (c *Column) Check() error {
	if c.AI && c.HasDefault {
		return fmt.Errorf("AutoIncrement 列 %s 不能同时包含默认值", c.Name)
	}

	if c.AI && c.Nullable {
		return fmt.Errorf("AutoIncrement 列 %s 不能同时带 NULL 约束", c.Name)
	}

	if c.PrimitiveType == String {
		if len(c.Length) > 0 && (c.Length[0] < -1 || c.Length[0] == 0) {
			return fmt.Errorf("列 %s 的长度只能是 -1 或是 >0", c.Name)
		}
	} else {
		for _, v := range c.Length {
			if v < 0 {
				return fmt.Errorf("列 %s 的长度不能小于 0", c.Name)
			}
		}
	}

	return nil
}

// AddColumns 添加新列
func (m *Model) AddColumns(col ...*Column) error {
	for _, c := range col {
		if err := m.AddColumn(c); err != nil {
			return err
		}
	}

	return nil
}

// AddColumn 添加新列
//
// 按添加顺序确定位置，越早添加的越在前。
func (m *Model) AddColumn(col *Column) error {
	if col.Name == "" {
		return errors.New("列必须存在名称")
	}

	if m.FindColumn(col.Name) != nil {
		return fmt.Errorf("列 %s 已经存在", col.Name)
	}

	m.Columns = append(m.Columns, col)

	if col.AI {
		return m.SetAutoIncrement(col)
	}
	return nil
}

// FindColumn 查找指定名称的列
//
// 不存在该列则返回 nil
func (m *Model) FindColumn(name string) *Column {
	for _, col := range m.Columns {
		if col.Name == name {
			return col
		}
	}
	return nil
}

func (m *Model) columnExists(col *Column) bool {
	for _, c := range m.Columns {
		if c == col {
			return true
		}
	}

	return false
}
