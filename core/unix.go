// SPDX-License-Identifier: MIT

package core

import (
	"bytes"
	"database/sql/driver"
	"strconv"
	"time"
)

// TimeFormatLayout 时间如果需要转换成字符串采用此格式
const TimeFormatLayout = time.RFC3339

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
	if src != nil {
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
			return ErrInvalidColumnType
		}
	}
	n.Time = time.Unix(unix, 0)
	return nil
}

// Value implements the driver.Valuer
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
	if bytes.Equal(unixNull, bytes.ToLower([]byte(v))) {
		n.IsNull = true
		return nil
	}

	t, err := time.Parse(TimeFormatLayout, v)
	if err != nil {
		return err
	}
	n.FromTime(t)
	n.IsNull = false
	return nil
}

// PrimitiveType 实现 PrimitiveTyper 接口
func (n Unix) PrimitiveType() PrimitiveType {
	return Int64
}

var unixNull = []byte("null")

// MarshalText encoding.TextMarshaler
func (n Unix) MarshalText() ([]byte, error) {
	if n.IsNull {
		return unixNull, nil
	}
	return n.Time.MarshalText()
}

// MarshalJSON implements json.Marshaler
func (n Unix) MarshalJSON() ([]byte, error) {
	if n.IsNull {
		return unixNull, nil
	}
	return n.Time.MarshalJSON()
}

// MarshalBinary implements encoding.BinaryMarshaler
func (n Unix) MarshalBinary() ([]byte, error) {
	if n.IsNull {
		return unixNull, nil
	}
	return n.Time.MarshalBinary()
}

// UnmarshalText encoding.TextUnmarshaler
func (n *Unix) UnmarshalText(data []byte) error {
	if bytes.Equal(unixNull, bytes.ToLower(data)) {
		n.IsNull = true
		return nil
	}

	n.IsNull = false
	return n.Time.UnmarshalText(data)
}

// UnmarshalJSON implements json.Unmarshaler
func (n *Unix) UnmarshalJSON(data []byte) error {
	if bytes.Equal(unixNull, bytes.ToLower(data)) {
		n.IsNull = true
		return nil
	}

	n.IsNull = false
	return n.Time.UnmarshalJSON(data)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (n *Unix) UnmarshalBinary(data []byte) error {
	if bytes.Equal(unixNull, bytes.ToLower(data)) {
		n.IsNull = true
		return nil
	}

	n.IsNull = false
	return n.Time.UnmarshalBinary(data)
}
