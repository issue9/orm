// SPDX-License-Identifier: MIT

package model

import (
	"testing"

	"github.com/issue9/assert/v2"
)

func BenchmarkNewModelNoCached(b *testing.B) {
	a := assert.New(b, false)
	ms := NewModels(nil)
	a.NotNil(ms)

	for i := 0; i < b.N; i++ {
		m, err := ms.New(&User{})
		a.NotError(err).NotNil(m)
		ms.Clear()
	}
}

func BenchmarkNewModelCached(b *testing.B) {
	a := assert.New(b, false)
	ms := NewModels(nil)
	a.NotNil(ms)

	for i := 0; i < b.N; i++ {
		m, err := ms.New(&User{})
		a.NotError(err).NotNil(m)
	}
}
