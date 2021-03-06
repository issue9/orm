// SPDX-License-Identifier: MIT

package model

import (
	"testing"

	"github.com/issue9/assert"
)

func TestModels(t *testing.T) {
	a := assert.New(t)

	ms := NewModels(nil)
	a.NotNil(ms)

	m, err := ms.New(&User{})
	a.NotError(err).
		NotNil(m).
		Equal(1, len(ms.items))

	// 相同的 model 实例，不会增加数量
	m, err = ms.New(&User{})
	a.NotError(err).
		NotNil(m).
		Equal(1, len(ms.items))

	// 添加新的 model
	m, err = ms.New(&Admin{})
	a.NotError(err).
		NotNil(m).
		Equal(2, len(ms.items))

	ms.Clear()
	a.Equal(0, len(ms.items))
}

func TestModels_addNames(t *testing.T) {
	a := assert.New(t)
	ms := NewModels(nil)
	a.NotNil(ms)

	a.Equal(0, len(ms.names))
	a.NotError(ms.addNames("n1"))
	a.NotError(ms.addNames("n2"))
	a.Equal(2, len(ms.names))

	a.Error(ms.addNames("n1"))
	a.Equal(2, len(ms.names))
}
