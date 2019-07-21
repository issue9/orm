// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package core

// Index 索引的类型
type Index int8

// Constraint 表示约束类型
type Constraint int8

// 索引的类型
const (
	IndexDefault Index = iota // 普通的索引
	IndexUnique               // 唯一索引
)

// 约束类型
const (
	ConstraintNone   Constraint = iota
	ConstraintUnique            // 唯一约束
	ConstraintFK                // 外键约束
	ConstraintCheck             // Check 约束
	ConstraintPK                // 主键约束
	ConstraintAI                // 自增
)

func (t Index) String() string {
	switch t {
	case IndexDefault:
		return "INDEX"
	case IndexUnique:
		return "UNIQUE INDEX"
	default:
		return "<unknown>"
	}
}

func (t Constraint) String() string {
	switch t {
	case ConstraintUnique:
		return "UNIQUE"
	case ConstraintFK:
		return "FOREIGN KEY"
	case ConstraintPK:
		return "PRIMARY KEY"
	case ConstraintCheck:
		return "CHECK"
	case ConstraintAI:
		return "AUTO INCREMENT"
	default:
		return "<unknown>"
	}
}
