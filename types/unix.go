// SPDX-License-Identifier: MIT

package types

import (
	"database/sql/driver"
	"encoding/json"
	"strconv"
	"time"

	"github.com/issue9/orm/v4/core"
)

// Unix 以 unix 时间戳保存的 time.Time 数据格式
type Unix struct {
	time.Time
	IsNull bool
}

// Scan implements the Scanner.Scan
func (n *Unix) Scan(src interface{}) (err error) {
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
		n.IsNull = true
		return nil
	}

	unix := int64(0)
	switch v := src.(type) {
	case int64:
		unix = v
	case []byte:
		if unix, err = strconv.ParseInt(string(v), 10, 64); err != nil {
			return err
		}
	case string:
		if unix, err = strconv.ParseInt(v, 10, 64); err != nil {
			return err
		}
	default:
		return core.ErrInvalidColumnType
	}

	n.Time = time.Unix(unix, 0)
	return nil
}

func (n Unix) Value() (driver.Value, error) {
	if n.IsNull {
		return nil, nil
	}

	return n.Time.Unix(), nil
}

// FromTime 从 time.Time 转换而来
func (n *Unix) FromTime(t time.Time) {
	n.IsNull = false
	n.Time = t
}

// ParseDefault 实现 DefaultParser 接口
func (n *Unix) ParseDefault(v string) error {
	n.IsNull = len(v) == 0
	if n.IsNull {
		return nil
	}

	t, err := time.Parse(core.TimeFormatLayout, v)
	if err != nil {
		n.IsNull = false
		return err
	}

	n.FromTime(t)
	return nil
}

func (n Unix) PrimitiveType() core.PrimitiveType { return core.Int64 }

func (n Unix) MarshalText() ([]byte, error) {
	if n.IsNull {
		return nil, nil
	}
	return n.Time.MarshalText()
}

func (n Unix) MarshalJSON() ([]byte, error) {
	if n.IsNull {
		return json.Marshal(nil)
	}
	return n.Time.MarshalJSON()
}

func (n *Unix) UnmarshalText(data []byte) error {
	n.IsNull = len(data) == 0
	if n.IsNull {
		return nil
	}
	return n.Time.UnmarshalText(data)
}

func (n *Unix) UnmarshalJSON(data []byte) error {
	n.IsNull = len(data) == 0
	if n.IsNull {
		return nil
	}
	return n.Time.UnmarshalJSON(data)
}
