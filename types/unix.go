// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package types

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"strconv"
	"time"

	"github.com/issue9/orm/v5/core"
)

// Unix 以 unix 时间戳保存的 time.Time 数据格式
type Unix struct {
	time.Time
	Valid bool
}

func (n *Unix) Scan(src any) (err error) {
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
		n.Valid = false
		return nil
	}

	unix := int64(0)
	switch v := src.(type) {
	case int64:
		unix = v
	case []byte:
		if bytes.Equal(v, nullBytes) {
			n.Valid = false
			return nil
		}
		if unix, err = strconv.ParseInt(string(v), 10, 64); err != nil {
			return err
		}
	case string:
		if v == null {
			n.Valid = false
			return nil
		}
		if unix, err = strconv.ParseInt(v, 10, 64); err != nil {
			return err
		}
	default:
		return core.ErrInvalidColumnType
	}

	n.Time = time.Unix(unix, 0)
	n.Valid = true
	return nil
}

func (n Unix) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}

	return n.Time.Unix(), nil
}

// FromTime 从 time.Time 转换而来
func (n *Unix) FromTime(t time.Time) {
	n.Valid = true
	n.Time = t
}

func (n Unix) PrimitiveType() core.PrimitiveType { return core.Int64 }

func (n Unix) MarshalText() ([]byte, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Time.MarshalText()
}

func (n Unix) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return json.Marshal(nil)
	}
	return n.Time.MarshalJSON()
}

func (n *Unix) UnmarshalText(data []byte) error {
	if n.Valid = len(data) > 0; !n.Valid {
		return nil
	}
	return n.Time.UnmarshalText(data)
}

func (n *Unix) UnmarshalJSON(data []byte) error {
	if n.Valid = len(data) > 0; !n.Valid {
		return nil
	}
	return n.Time.UnmarshalJSON(data)
}
