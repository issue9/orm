// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package model

type conType int8

// 预定的约束类型，方便 Model 中使用。
const (
	none conType = iota
	index
	unique
	fk
	check
)

func (t conType) String() string {
	switch t {
	case none:
		return "<none>"
	case index:
		return "KEY INDEX"
	case unique:
		return "UNIQUE INDEX"
	case fk:
		return "FOREIGN KEY"
	case check:
		return "CHECK"
	default:
		return "<unknown>"
	}
}
