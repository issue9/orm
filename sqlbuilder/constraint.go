// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import "github.com/issue9/orm/v2/core"

// AddConstraintStmtHooker AddConstraintStmt.DDLSQL 的钩子函数
type AddConstraintStmtHooker interface {
	AddConstraintStmtHook(*AddConstraintStmt) ([]string, error)
}

// AddConstraintStmt 添加约束
type AddConstraintStmt struct {
	*ddlStmt

	TableName string
	Name      string
	Type      core.Constraint

	// 约束的值，根据 Type 的不同，略有不同：
	// check 下表示的 check 表达式，仅有一个元素；
	// fk 下最多可以有 5 个值，第 1 个元素为关联的列，2、3 元素引用的表和列，
	//  4，5 元素为 UPDATE 和 DELETE 的规则定义；
	// 其它模式下为该约束关联的列名称。
	Data []string
}

// AddConstraint 声明添加约束的语句
func AddConstraint(e core.Engine) *AddConstraintStmt {
	stmt := &AddConstraintStmt{}
	stmt.ddlStmt = newDDLStmt(e, stmt)
	return stmt
}

// Reset 重置内容
func (stmt *AddConstraintStmt) Reset() *AddConstraintStmt {
	stmt.baseStmt.Reset()
	stmt.TableName = ""
	stmt.Name = ""
	stmt.Type = core.ConstraintNone
	stmt.Data = stmt.Data[:0]
	return stmt
}

// Table 指定表名
func (stmt *AddConstraintStmt) Table(t string) *AddConstraintStmt {
	stmt.TableName = t
	return stmt
}

// Unique 指定唯一约束
func (stmt *AddConstraintStmt) Unique(name string, col ...string) *AddConstraintStmt {
	if stmt.Type != core.ConstraintNone {
		stmt.err = ErrConstraintType
	}

	stmt.Type = core.ConstraintUnique
	stmt.Name = name
	stmt.Data = col

	return stmt
}

// PK 指定主键约束
func (stmt *AddConstraintStmt) PK(col ...string) *AddConstraintStmt {
	if stmt.Type != core.ConstraintNone {
		stmt.err = ErrConstraintType
	}

	stmt.Type = core.ConstraintPK
	stmt.Data = col

	return stmt
}

// Check Check 约束
func (stmt *AddConstraintStmt) Check(name, expr string) *AddConstraintStmt {
	if stmt.Type != core.ConstraintNone {
		stmt.err = ErrConstraintType
	}

	stmt.Type = core.ConstraintCheck
	stmt.Name = name
	stmt.Data = []string{expr}

	return stmt
}

// FK 外键约束
func (stmt *AddConstraintStmt) FK(name, col, refTable, refColumn, updateRule, deleteRule string) *AddConstraintStmt {
	if stmt.Type != core.ConstraintNone {
		stmt.err = ErrConstraintType
	}

	stmt.Type = core.ConstraintFK
	stmt.Name = name
	stmt.Data = []string{col, refTable, refColumn, updateRule, deleteRule}

	return stmt
}

// DDLSQL 生成 SQL 语句
func (stmt *AddConstraintStmt) DDLSQL() ([]string, error) {
	if stmt.err != nil {
		return nil, stmt.Err()
	}

	if stmt.TableName == "" {
		return nil, ErrTableIsEmpty
	}

	if len(stmt.Data) == 0 {
		return nil, ErrColumnsIsEmpty
	}

	if stmt.Type == core.ConstraintPK {
		stmt.Name = core.PKName(stmt.TableName)
	}

	if stmt.Name == "" {
		return nil, ErrConstraintIsEmpty
	}

	if hook, ok := stmt.Dialect().(AddConstraintStmtHooker); ok {
		return hook.AddConstraintStmtHook(stmt)
	}

	builder := core.NewBuilder("ALTER TABLE ").
		QuoteKey(stmt.TableName).
		WriteString(" ADD CONSTRAINT ").
		QuoteKey(stmt.Name)

	switch stmt.Type {
	case core.ConstraintCheck:
		builder.WriteString(" CHECK(").WriteString(stmt.Data[0]).WriteBytes(')')
	case core.ConstraintFK:
		builder.WriteString(" FOREIGN KEY(").
			QuoteKey(stmt.Data[0]).
			WriteString(") REFERENCES ").
			QuoteKey(stmt.Data[1]).
			WriteBytes('(').
			QuoteKey(stmt.Data[2]).
			WriteBytes(')')

		if stmt.Data[3] != "" {
			builder.WriteString(" ON UPDATE ").WriteString(stmt.Data[3])
		}

		if stmt.Data[4] != "" {
			builder.WriteString(" ON DELETE ").WriteString(stmt.Data[4])
		}
	case core.ConstraintPK:
		builder.WriteString(" PRIMARY KEY(")
		for _, col := range stmt.Data {
			builder.
				QuoteKey(col).
				WriteBytes(',')
		}
		builder.TruncateLast(1).WriteBytes(')')
	case core.ConstraintUnique:
		builder.WriteString(" UNIQUE(")
		for _, col := range stmt.Data {
			builder.
				QuoteKey(col).
				WriteBytes(',')
		}
		builder.TruncateLast(1).WriteBytes(')')
	default:
		return nil, ErrUnknownConstraint
	}

	query, err := builder.String()
	if err != nil {
		return nil, err
	}
	return []string{query}, nil
}

// DropConstraintStmtHooker DropConstraintStmt.DDLSQL 的钩子函数
type DropConstraintStmtHooker interface {
	DropConstraintStmtHook(*DropConstraintStmt) ([]string, error)
}

// DropConstraintStmt 删除约束
type DropConstraintStmt struct {
	*ddlStmt

	TableName string
	Name      string
}

// DropConstraint 声明一条删除表约束的语句
func DropConstraint(e core.Engine) *DropConstraintStmt {
	stmt := &DropConstraintStmt{}
	stmt.ddlStmt = newDDLStmt(e, stmt)
	return stmt
}

// Table 指定表名。
//
// 重复指定，会覆盖之前的。
func (stmt *DropConstraintStmt) Table(table string) *DropConstraintStmt {
	stmt.TableName = table
	return stmt
}

// Constraint 指定需要删除的约束名
//
// NOTE: 如果需要删除主键，请使用 PKName 产生主键名称
func (stmt *DropConstraintStmt) Constraint(name string) *DropConstraintStmt {
	stmt.Name = name
	return stmt
}

// DDLSQL 获取 SQL 语句以及对应的参数
func (stmt *DropConstraintStmt) DDLSQL() ([]string, error) {
	if stmt.err != nil {
		return nil, stmt.Err()
	}

	if stmt.TableName == "" {
		return nil, ErrTableIsEmpty
	}

	if stmt.Name == "" {
		return nil, ErrValueIsEmpty
	}

	if hook, ok := stmt.Dialect().(DropConstraintStmtHooker); ok {
		return hook.DropConstraintStmtHook(stmt)
	}

	builder := core.NewBuilder("ALTER TABLE ").
		QuoteKey(stmt.TableName).
		WriteString(" DROP CONSTRAINT ").
		QuoteKey(stmt.Name)

	query, err := builder.String()
	if err != nil {
		return nil, err
	}
	return []string{query}, nil
}

// Reset 重置
func (stmt *DropConstraintStmt) Reset() *DropConstraintStmt {
	stmt.baseStmt.Reset()
	stmt.TableName = ""
	stmt.Name = ""
	return stmt
}
