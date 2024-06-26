// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package types

import (
	"bytes"
	"database/sql/driver"
	"math/big"

	"github.com/issue9/orm/v6/core"
)

// Rat 有理数
//
// 这是对 [big.Rat] 的扩展，提供了 orm 需要的接口支持。
//
// 在数据库中以分数的形式保存至字符串类型的列，所以需要指定长度。
type Rat struct {
	rat *big.Rat
}

func Rational(a, b int64) Rat { return Rat{rat: big.NewRat(a, b)} }

// Scan implements the sql.Scanner
func (n *Rat) Scan(src any) (err error) {
	// The src value will be of one of the following types:
	//
	//    int64
	//    float64
	//    bool
	//    []byte
	//    string
	//    time.Time
	//    nil - for NULL values
	if src == nil {
		n.rat = nil
		return nil
	}

	switch v := src.(type) {
	case []byte:
		if bytes.Equal(v, nullBytes) {
			n.rat = nil
			return nil
		}
		return n.UnmarshalText(v)
	case string:
		if v == null {
			n.rat = nil
			return nil
		}
		return n.UnmarshalText([]byte(v))
	default:
		return core.ErrInvalidColumnType()
	}
}

func (n Rat) Value() (driver.Value, error) {
	if n.IsNull() {
		return nil, nil
	}
	return n.Rat().String(), nil
}

// Rat 返回标准库中 [big.Rat] 的实例
func (n Rat) Rat() *big.Rat { return n.rat }

func (n Rat) PrimitiveType() core.PrimitiveType { return core.String }

func (n Rat) MarshalText() ([]byte, error) {
	if n.IsNull() {
		return []byte{}, nil
	}
	return n.Rat().MarshalText()
}

func (n *Rat) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		n.rat = nil
		return nil
	}

	if n.IsNull() {
		n.rat = new(big.Rat)
	}
	return n.Rat().UnmarshalText(data)
}

func (n Rat) IsNull() bool { return n.Rat() == nil }
