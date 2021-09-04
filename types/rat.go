// SPDX-License-Identifier: MIT

package types

import (
	"database/sql/driver"
	"math/big"

	"github.com/issue9/orm/v4/core"
)

// Rat 有理数
//
// 这是对 bit.Rat 的扩展，提供了 orm 需要的接口支持。
type Rat struct {
	rat *big.Rat
}

// Scan implements the Scanner.Scan
func (n *Rat) Scan(src interface{}) (err error) {
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
		return n.rat.UnmarshalText(v)
	case string:
		return n.rat.UnmarshalText([]byte(v))
	default:
		return core.ErrInvalidColumnType
	}
}

func (n *Rat) Value() (driver.Value, error) { return n.Rat(), nil }

// Rat 返回标准库中 bit.Rat 的实例
func (n *Rat) Rat() *big.Rat { return n.rat }

// ParseDefault 实现 DefaultParser 接口
func (n *Rat) ParseDefault(v string) error { return n.UnmarshalText([]byte(v)) }

func (n Rat) PrimitiveType() core.PrimitiveType { return core.Bytes }

func (n Rat) MarshalText() ([]byte, error) { return n.Rat().MarshalText() }

func (n *Rat) UnmarshalText(data []byte) error { return n.Rat().UnmarshalText(data) }
