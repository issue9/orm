// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package model

import (
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/orm/v6/internal/model/testdata"
)

func (ms *Models) len() (cnt int) {
	ms.models.Range(func(key, value any) bool {
		cnt++
		return true
	})
	return
}

func TestModels(t *testing.T) {
	a := assert.New(t, false)

	ms := NewModels()
	a.NotNil(ms)

	m, err := ms.New(&User{})
	a.NotError(err).
		NotNil(m).
		Equal(1, ms.len())

	// 相同的 model 实例，不会增加数量
	m, err = ms.New(&User{})
	a.NotError(err).
		NotNil(m).
		Equal(1, ms.len())

	// 相同的表名，但是类型不同
	m, err = ms.New(&testdata.User{})
	a.NotError(err).
		NotNil(m).
		Equal(2, ms.len())

	// 添加新的 model
	m, err = ms.New(&Admin{})
	a.NotError(err).
		NotNil(m).
		Equal(3, ms.len())

	ms.Clear()
	a.Equal(0, ms.len())
}
