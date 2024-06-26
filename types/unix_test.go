// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"testing"
	"time"

	"github.com/issue9/assert/v4"

	"github.com/issue9/orm/v6/core"
)

var (
	_ sql.Scanner         = &Unix{}
	_ driver.Valuer       = Unix{}
	_ core.PrimitiveTyper = &Unix{}

	_ encoding.BinaryMarshaler   = Unix{}
	_ encoding.TextMarshaler     = Unix{}
	_ json.Marshaler             = Unix{}
	_ encoding.BinaryUnmarshaler = &Unix{}
	_ encoding.TextUnmarshaler   = &Unix{}
	_ json.Unmarshaler           = &Unix{}
)

func TestUnix_Scan(t *testing.T) {
	a := assert.New(t, false)

	u := &Unix{}
	a.NotError(u.Scan(int64(1))).
		Equal(1, u.Time.Unix())

	u = &Unix{}
	a.NotError(u.Scan("123")).
		Equal(123, u.Time.Unix())

	u = &Unix{}
	a.NotError(u.Scan([]byte("123"))).
		Equal(123, u.Time.Unix()).
		True(u.Valid)

	u = &Unix{}
	a.NotError(u.Scan(nil)).
		False(u.Valid)

	// 无法解析的值
	u = &Unix{}
	a.Error(u.Scan(int32(1)))
	u = &Unix{}
	a.Error(u.Scan("123x"))

	// 无效的类型
	u = &Unix{}
	a.Error(u.Scan(int32(1)))
	u = &Unix{}
	a.Error(u.Scan(&struct{ X int }{X: 5}))

	u = &Unix{}
	a.NotError(u.Scan(nil))
}

func TestUnix_Unmarshal(t *testing.T) {
	a := assert.New(t, false)

	now := time.Now()
	format := now.Format(time.RFC3339)
	j := `{"u":"` + format + `"}`

	obj := struct {
		U Unix `json:"u"`
	}{}
	a.NotError(json.Unmarshal([]byte(j), &obj))
	a.Equal(now.Unix(), obj.U.Unix())

	jj, err := json.Marshal(obj)
	a.NotError(err).Equal(string(jj), j)
}
