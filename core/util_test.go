// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package core

import (
	"strconv"
	"testing"

	"github.com/issue9/assert"
)

func TestExtractArgs(t *testing.T) {
	a := assert.New(t)

	// 带@参数
	sql, args := ExtractArgs("SELECT * FROM #user where (id=@id and username=@username)and group=@g")
	a.Equal("SELECT * FROM #user where (id=? and username=?)and group=?", sql).
		Equal(args, []string{"id", "username", "g"})

	// 没带@参数
	sql, args = ExtractArgs("SELECT * FROM USER where true")
	a.Equal(sql, "SELECT * FROM USER where true").Equal(len(args), 0)
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

	delete(args, "2")
	val, err = ConvArgs(names, args)
	a.Error(err).Nil(val)
}

func TestAsSQLValue(t *testing.T) {
	a := assert.New(t)

	// bool
	a.Equal(AsSQLValue(true), "true")

	// int
	a.Equal(AsSQLValue(5), "5")

	// uint
	a.Equal(AsSQLValue(uint(5)), "5")

	// float64
	a.Equal(AsSQLValue(7.0), "7")
	a.Equal(AsSQLValue(7.1), "7.1")

	// float32
	a.Equal(AsSQLValue(float32(7.0)), "7")
	a.Equal(AsSQLValue(float32(7.1)), "7.1")

	// string
	a.Equal(AsSQLValue("abc"), "'abc'")
	a.Equal(AsSQLValue("@abc"), "@abc")

	// []byte
	a.Equal(AsSQLValue([]byte("abc")), "'abc'")
	a.Equal(AsSQLValue([]byte("@abc")), "@abc")

	// []rune
	a.Equal(AsSQLValue([]rune("abc")), "'abc'")
	a.Equal(AsSQLValue([]rune("@abc")), "@abc")

	// 自定义
	type INT int
	a.Equal(AsSQLValue(INT(5)), "5")

	a.Equal(AsSQLValue(UINT{val: 5}), "5")

}

type UINT struct {
	val uint64
}

func (i UINT) String() string {
	return strconv.FormatUint(i.val, 10)
}
