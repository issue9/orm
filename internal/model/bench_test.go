// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package model_test

import (
	"testing"

	"github.com/issue9/assert/v4"
)

func BenchmarkNewModelNoCached(b *testing.B) {
	a := assert.New(b, false)
	ms := newModules(a)

	for i := 0; i < b.N; i++ {
		m, err := ms.New(&User{})
		a.NotError(err).NotNil(m)
		ms.Close()
	}
}

func BenchmarkNewModelCached(b *testing.B) {
	a := assert.New(b, false)
	ms := newModules(a)

	for i := 0; i < b.N; i++ {
		m, err := ms.New(&User{})
		a.NotError(err).NotNil(m)
	}
}
