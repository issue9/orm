// SPDX-License-Identifier: MIT

package types

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/issue9/orm/v5/core"
)

// SliceOf 针对数组存组方式
//
// 最终是以 json 的方式保存在数据库。
type SliceOf[T any] []T

func (n *SliceOf[T]) Scan(value any) (err error) {
	if value == nil {
		return nil
	}

	var j []byte
	switch v := value.(type) {
	case string:
		j = []byte(v)
	case []byte:
		j = v
	default:
		return core.ErrInvalidColumnType
	}

	return json.Unmarshal(j, n)
}

func (n SliceOf[T]) Value() (driver.Value, error) { return json.Marshal(n) }

func (n SliceOf[T]) PrimitiveType() core.PrimitiveType { return core.Bytes }
