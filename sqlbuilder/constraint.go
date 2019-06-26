// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"context"
	"database/sql"
)

// Constraint 表示约束类型
type Constraint int8

// 约束类型
const (
	ConstraintUnique Constraint = iota // 唯一约束
	ConstraintFK                       // 外键约束
	ConstraintCheck                    // Check 约束
	ConstraintPK                       // 主键约束
	ConstraintAI                       // 自增
	constraintNone
)

// AddConstraintStmt 添加约束
type AddConstraintStmt struct {
	engine Engine
	table  string
	name   string
	typ    Constraint
	cols   []string
	fk     *foreignKey
}

// AddConstraint 声明添加约束的语句
func AddConstraint(e Engine) *AddConstraintStmt {
	return &AddConstraintStmt{
		engine: e,
	}
}

// Reset 重置内容
func (stmt *AddConstraintStmt) Reset() {
	stmt.table = ""
	stmt.name = ""
	stmt.typ = constraintNone
	stmt.cols = stmt.cols[:0]
}

// Table 指定表名
func (stmt *AddConstraintStmt) Table(t string) *AddConstraintStmt {
	stmt.table = t
	return stmt
}

// Unique 指定唯一约束
func (stmt *AddConstraintStmt) Unique(name string, col ...string) *AddConstraintStmt {
	if stmt.typ != constraintNone {
		panic(ErrConstraintType)
	}

	stmt.typ = ConstraintUnique
	stmt.name = name
	stmt.cols = col

	return stmt
}

// PK 指定主键约束
func (stmt *AddConstraintStmt) PK(name string, col ...string) *AddConstraintStmt {
	if stmt.typ != constraintNone {
		panic(ErrConstraintType)
	}

	stmt.typ = ConstraintPK
	stmt.name = name
	stmt.cols = col

	return stmt
}

// Check Check 约束
func (stmt *AddConstraintStmt) Check(name, expr string) *AddConstraintStmt {
	if stmt.typ != constraintNone {
		panic(ErrConstraintType)
	}

	stmt.typ = ConstraintCheck
	stmt.name = name
	stmt.cols = []string{expr}

	return stmt
}

// FK 外键约束
func (stmt *AddConstraintStmt) FK(name, col, refTable, refColumn, updateRule, deleteRule string) *AddConstraintStmt {
	if stmt.typ != constraintNone {
		panic(ErrConstraintType)
	}

	stmt.typ = ConstraintFK
	stmt.name = name
	stmt.cols = []string{col}
	stmt.fk = &foreignKey{
		Name:         name,
		Column:       col,
		RefTableName: refTable,
		RefColName:   refColumn,
		UpdateRule:   updateRule,
		DeleteRule:   deleteRule,
	}

	return stmt
}

// AI 自增约束
func (stmt *AddConstraintStmt) AI(name, col string) *AddConstraintStmt {
	if stmt.typ != constraintNone {
		panic(ErrConstraintType)
	}

	stmt.typ = ConstraintAI
	stmt.name = name
	stmt.cols = []string{col}

	return stmt
}

// SQL 生成 SQL 语句
func (stmt *AddConstraintStmt) SQL() (string, []interface{}, error) {
	if stmt.table == "" {
		return "", nil, ErrTableIsEmpty
	}

	if len(stmt.cols) == 0 {
		return "", nil, ErrColumnsIsEmpty
	}

	builder := New("ALTER TABLE ").
		WriteString(stmt.table).
		WriteString(" ADD CONSTRAINT ").
		WriteString(stmt.name)

	switch stmt.typ {
	case ConstraintAI:
		// TODO
	case ConstraintCheck:
		builder.WriteString(" CHECK ").WriteString(stmt.cols[0])
	case ConstraintFK:
		// TODO
	case ConstraintPK:
		builder.WriteString(" PRIMARY KEY ")
		for _, col := range stmt.cols {
			builder.WriteString(col)
		}
	case ConstraintUnique:
		builder.WriteString(" UNIQUE ")
		for _, col := range stmt.cols {
			builder.WriteString(col)
		}
	default:
		return "", nil, ErrUnknownConstraint
	}

	return builder.String(), nil, nil
}

// Exec 执行 SQL 语句
func (stmt *AddConstraintStmt) Exec() (sql.Result, error) {
	return stmt.ExecContext(context.Background())
}

// ExecContext 执行 SQL 语句
func (stmt *AddConstraintStmt) ExecContext(ctx context.Context) (sql.Result, error) {
	return execContext(ctx, stmt.engine, stmt)
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
