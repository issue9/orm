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
	if value == nil {
		n = &Unix{}
		return nil
	}

	switch v := value.(type) {
	case int64:
		vv := time.Unix(v, 0)
		n = (*Unix)(&vv)
		return nil
	default:
		return ErrInvalidColumnType
	}
}

// Value implements the driver.Valuer
func (n Unix) Value() (driver.Value, error) {
	return int64(time.Time(n).Unix()), nil
}
