// SPDX-License-Identifier: MIT

package types

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/issue9/orm/v5/core"
)

type Decimal struct {
	Decimal   decimal.Decimal
	Precision int32
	Valid     bool
}

// FloatDecimal 从浮点数还原 Decimal 对象
//
// precision 表示输出的精度。
func FloatDecimal(f float64, precision int32) Decimal {
	return Decimal{Decimal: decimal.NewFromFloat(f), Precision: precision, Valid: true}
}

// StringDecimal 从字符串还原 Decimal 对象
//
// precision 表示输出的精度。
func StringDecimal(s string, precision int32) (Decimal, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return Decimal{}, err
	}
	return Decimal{Decimal: d, Precision: precision, Valid: true}, nil
}

// StringDecimalWithPrecision 从字符串还原 Decimal 对象
//
// 输出精度从 s 获取，如果 s 不包含小数位，则小数长度为 0
func StringDecimalWithPrecision(s string) (Decimal, error) {
	var p int32
	index := strings.IndexByte(s, '.')
	if index >= 0 {
		p = int32(len(s) - index - 1)
	}

	return StringDecimal(s, p)
}

// Scan implements the Scanner.Scan
func (n *Decimal) Scan(src any) (err error) {
	if src == nil {
		n.Valid = false
		return nil
	}

	switch v := src.(type) {
	case []byte:
		if bytes.Equal(v, nullBytes) {
			n.Valid = false
			return nil
		}
	case string:
		if v == null {
			n.Valid = false
			return nil
		}
	}
	return n.Decimal.Scan(src)
}

func (n Decimal) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Decimal.StringFixed(n.Precision), nil
}

func (n Decimal) PrimitiveType() core.PrimitiveType { return core.Decimal }

func (n Decimal) MarshalText() ([]byte, error) {
	if !n.Valid {
		return nil, nil
	}
	return []byte(n.Decimal.StringFixed(n.Precision)), nil
}

func (n Decimal) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(n.Decimal.StringFixed(n.Precision))
}

func (n *Decimal) UnmarshalText(data []byte) error {
	if n.Valid = len(data) > 0; !n.Valid {
		return nil
	}
	return n.Decimal.UnmarshalText(data)
}

func (n *Decimal) UnmarshalJSON(data []byte) error {
	if n.Valid = len(data) > 0; !n.Valid {
		return nil
	}
	return n.Decimal.UnmarshalJSON(data)
}
