// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package core

import (
	"testing"

	"github.com/issue9/assert"
)

func TestIndex_String(t *testing.T) {
	a := assert.New(t)

	a.Equal(IndexUnique.String(), "UNIQUE INDEX")
	a.Equal(IndexDefault.String(), "INDEX")
	a.Equal(Index(3).String(), "<unknown>")
	a.Equal(Index(-1).String(), "<unknown>")
}

func TestConstraint_String(t *testing.T) {
	a := assert.New(t)

	a.NotEqual(ConstraintAI.String(), "<unknown>")
	a.NotEqual(ConstraintFK.String(), "<unknown>")
	a.NotEqual(ConstraintPK.String(), "<unknown>")
	a.NotEqual(ConstraintUnique.String(), "<unknown>")
	a.NotEqual(ConstraintCheck.String(), "<unknown>")
	a.Equal(Constraint(-1).String(), "<unknown>")
	a.Equal(Constraint(100).String(), "<unknown>")
}
