// SPDX-License-Identifier: MIT

package core

import "fmt"

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
//
// 以下定义了一些常用的约束类型，但是并不是所有的数据都支持这些约束类型，
// 比如 mysql<8.0.16 和 mariadb<10.2.1 不支持 check 约束。
const (
	ConstraintNone   Constraint = iota
	ConstraintUnique            // 唯一约束
	ConstraintFK                // 外键约束
	ConstraintCheck             // Check 约束
	ConstraintPK                // 主键约束
	ConstraintAI                // 自增
)

// ErrConstraintExists 返回约束名已经存在的错误
func ErrConstraintExists(c string) error {
	return fmt.Errorf("约束 %s 已经存在", c)
}
