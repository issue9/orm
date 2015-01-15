// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package core

import (
	"testing"

	"github.com/issue9/assert"
)

func TestExtractArgs(t *testing.T) {
	a := assert.New(t)

	sql, args := ExtractArgs("SELECT * FROM #user where (id=@id and username=@username)and group=@g")
	a.Equal("SELECT * FROM #user where (id=? and username=?)and group=?", sql).
		Equal(args, []string{"id", "username", "g"})
}

func TestConvArgs(t *testing.T) {
	a := assert.New(t)

	names := []string{"1", "2", "3"}
	args := map[string]interface{}{
		"1": 1,
		"2": "two",
		"3": "三",
	}

	val, err := ConvArgs(names, args)
	a.NotError(err).Equal(val, []interface{}{1, "two", "三"})
}

func TestAsSQLValue(t *testing.T) {
	a := assert.New(t)

	a.Equal(AsSQLValue(5), "5")
	a.Equal(AsSQLValue(7.0), "7")
	a.Equal(AsSQLValue(7.1), "7.1")
	a.Equal(AsSQLValue("abc"), "'abc'")
	a.Equal(AsSQLValue("@abc"), "@abc")
}
