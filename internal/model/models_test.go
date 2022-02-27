// SPDX-License-Identifier: MIT

package model

import (
	"testing"

	"github.com/issue9/assert/v2"
)

func TestModels(t *testing.T) {
	a := assert.New(t, false)

	ms := NewModels(nil)
	a.NotNil(ms)

	m, err := ms.New(&User{})
	a.NotError(err).
		NotNil(m).
		Equal(1, len(ms.models))

	// 相同的 model 实例，不会增加数量
	m, err = ms.New(&User{})
	a.NotError(err).
		NotNil(m).
		Equal(1, len(ms.models))

	// 添加新的 model
	m, err = ms.New(&Admin{})
	a.NotError(err).
		NotNil(m).
		Equal(2, len(ms.models))

	ms.Clear()
	a.Equal(0, len(ms.models))
}
