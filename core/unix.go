// SPDX-License-Identifier: MIT

package core

import (
	"database/sql/driver"
	"strconv"
	"time"
)

// TimeFormatLayout 时间如果需要转换成字符串采用此格式
const TimeFormatLayout = time.RFC3339

// Unix 以 unix 时间戳保存的 time.Time 数据格式
type Unix time.Time

// Scan implements the Scanner.Scan
func (n *Unix) Scan(value interface{}) (err error) {
	unix := int64(0)
	if value != nil {
		switch v := value.(type) {
		case int64:
			unix = v
		case int:
			unix = int64(v)
		case []byte:
			if unix, err = strconv.ParseInt(string(v), 10, 64); err != nil {
				return err
			}
		case string:
			if unix, err = strconv.ParseInt(v, 10, 64); err != nil {
				return err
			}
		case []rune:
			if unix, err = strconv.ParseInt(string(v), 10, 64); err != nil {
				return err
			}
		default:
			return ErrInvalidColumnType
		}
	}
	*n = Unix(time.Unix(unix, 0))
	return nil
}

// Value implements the driver.Valuer
func (n Unix) Value() (driver.Value, error) {
	return n.AsTime().Unix(), nil
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
	t, err := time.Parse(TimeFormatLayout, v)
	if err != nil {
		return err
	}
	n.FromTime(t)
	return nil
}

// PrimitiveType 实现 PrimitiveTyper 接口
func (n Unix) PrimitiveType() PrimitiveType {
	return Int64
}

// MarshalText encoding.TextMarshaler
func (n Unix) MarshalText() ([]byte, error) {
	return n.AsTime().MarshalText()
}

// MarshalJSON implements json.Marshaler
func (n Unix) MarshalJSON() ([]byte, error) {
	return n.AsTime().MarshalJSON()
}

// MarshalBinary implements encoding.BinaryMarshaler
func (n Unix) MarshalBinary() ([]byte, error) {
	return n.AsTime().MarshalBinary()
}

// UnmarshalText encoding.TextUnmarshaler
func (n *Unix) UnmarshalText(data []byte) error {
	t := (*time.Time)(n)
	return t.UnmarshalText(data)
}

// UnmarshalJSON implements json.Unmarshaler
func (n *Unix) UnmarshalJSON(data []byte) error {
	t := (*time.Time)(n)
	return t.UnmarshalJSON(data)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (n *Unix) UnmarshalBinary(data []byte) error {
	t := (*time.Time)(n)
	return t.UnmarshalBinary(data)
}
