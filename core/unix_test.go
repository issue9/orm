// SPDX-License-Identifier: MIT

package core

import (
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/issue9/assert"
)

var (
	_ sql.Scanner    = &Unix{}
	_ driver.Valuer  = Unix{}
	_ DefaultParser  = &Unix{}
	_ PrimitiveTyper = &Unix{}
)

func TestUnix_ParseDefault(t *testing.T) {
	a := assert.New(t)

	u := &Unix{}
	a.Error(u.ParseDefault("2020"))

	now := time.Now()
	a.NotError(u.ParseDefault(now.Format(time.RFC3339))).
		Equal(now.Unix(), u.AsTime().Unix())
}

func TestUnix_Scan(t *testing.T) {
	a := assert.New(t)

	u := &Unix{}
	a.NotError(u.Scan(int64(1))).
		Equal(1, u.AsTime().Unix())

	u = &Unix{}
	a.NotError(u.Scan(1)).
		Equal(1, u.AsTime().Unix())

	// 无效的类型
	u = &Unix{}
	a.Error(u.Scan(int32(1)))
	u = &Unix{}
	a.Error(u.Scan("123"))

	u = &Unix{}
	a.NotError(u.Scan(nil))
}
