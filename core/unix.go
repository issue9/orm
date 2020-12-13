// SPDX-License-Identifier: MIT

package core

import (
	"database/sql/driver"
	"time"
)

// Unix 以 unix 时间戳保存的 time.Time 数据格式
type Unix time.Time

// Scan implements the Scanner.Scan
func (n *Unix) Scan(value interface{}) error {
	if value != nil {
		switch v := value.(type) {
		case int64:
			vv := time.Unix(v, 0)
			*n = Unix(vv)
		case int:
			vv := time.Unix(int64(v), 0)
			*n = Unix(vv)
		default:
			return ErrInvalidColumnType
		}
	}
	return nil
}

// Value implements the driver.Valuer
func (n Unix) Value() (driver.Value, error) {
	return int64(time.Time(n).Unix()), nil
}

// AsTime 转换成 time.Time
func (n Unix) AsTime() time.Time {
	return time.Time(n)
}

// FromTime 从 time.Time 转换而来
func (n *Unix) FromTime(t time.Time) {
	*n = Unix(t)
}

// ParseDefault 实现 DefaultParser 接口
func (n *Unix) ParseDefault(v string) error {
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return err
	}
	n.FromTime(t)
	return nil
}

// PrimitiveType 实现 PrimitiveTyper 接口
func (n *Unix) PrimitiveType() PrimitiveType {
	return Int64
}
