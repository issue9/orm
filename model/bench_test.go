// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/orm/internal/modeltest"
)

// BenchmarkNewModelNoCached	  100000	     23724 ns/op
func BenchmarkNewModelNoCached(b *testing.B) {
	ClearModels()
	a := assert.New(b)

	for i := 0; i < b.N; i++ {
		m, err := New(&modeltest.User{})
		ClearModels()
		a.NotError(err).NotNil(m)
	}
}

// BenchmarkNewModelCached	 3000000	       480 ns/op
func BenchmarkNewModelCached(b *testing.B) {
	ClearModels()
	a := assert.New(b)

	for i := 0; i < b.N; i++ {
		m, err := New(&modeltest.User{})
		a.NotError(err).NotNil(m)
	}
}
