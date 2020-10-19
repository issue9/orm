// SPDX-License-Identifier: MIT

package test

import (
	"testing"

	"github.com/issue9/assert"
)

func TestSuite_ForEach(t *testing.T) {
	a := assert.New(t)

	s := NewSuite(a)
	defer s.Close()

	var size int
	s.ForEach(func(t *Driver) {
		a.NotNil(t).
			NotNil(t.DB).
			NotNil(t.DB.Dialect()).
			NotNil(t.DB.DB).
			Equal(t.Assertion, a)
		size++
	})
	a.Equal(size, len(cases))

	// 指定了不存在的 driverName
	a.Panic(func() {
		s.ForEach(func(t *Driver) {
			size++
		}, Mysql, nil)
	})
}
