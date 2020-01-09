// SPDX-License-Identifier: MIT

package model

import (
	"testing"

	"github.com/issue9/assert"
)

// go1.10 BenchmarkNewModelNoCached-4   	  200000	      8161 ns/op
func BenchmarkNewModelNoCached(b *testing.B) {
	a := assert.New(b)
	ms := NewModels(nil)
	a.NotNil(ms)

	for i := 0; i < b.N; i++ {
		m, err := ms.New(&User{})
		a.NotError(err).NotNil(m)
		ms.Clear()
	}
}

// go1.10 BenchmarkNewModelCached-4     	10000000	       187 ns/op
func BenchmarkNewModelCached(b *testing.B) {
	a := assert.New(b)
	ms := NewModels(nil)
	a.NotNil(ms)

	for i := 0; i < b.N; i++ {
		m, err := ms.New(&User{})
		a.NotError(err).NotNil(m)
	}
}
