// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

// Constraint 表示约束类型
type Constraint int8

// 约束类型
const (
	constraintNone   Constraint = iota
	ConstraintUnique            // 唯一约束
	ConstraintFK                // 外键约束
	ConstraintCheck             // Check 约束
	ConstraintPK                // 主键约束
	ConstraintAI                // 自增
)

const (
	defaultAINameSuffix = "_ai"
	defaultPKNameSuffix = "_pk"
)

// PKName 生成主键约束的名称
//
// 各个数据库对主键约束的规定并不统一，mysql 会忽略约束名，
// 为了统一，主键约束的名称统一由此函数生成，用户不能别外指定。
func PKName(table string) string {
	return table + defaultPKNameSuffix
}

// AIName 生成 AI 约束名称
//
// 自增约束的实现，各个数据库并不相同，诸如 mysql 直接加在列信息上，
// 而 postgres 会创建 sequence，需要指定 sequence 名称。
func AIName(table string) string {
	return table + defaultAINameSuffix
}

// AddConstraintStmtHooker AddConstraintStmt.DDLSQL 的钩子函数
type AddConstraintStmtHooker interface {
	AddConstraintStmtHook(*AddConstraintStmt) ([]string, error)
}

// AddConstraintStmt 添加约束
type AddConstraintStmt struct {
	*ddlStmt

	TableName string
	Name      string
	Type      Constraint

	// 约束的值，根据 Type 的不同，略有不同：
	// check 下表示的 check 表达式，仅有一个元素；
	// fk 下最多可以有 5 个值，第 1 个元素为关联的列，2、3 元素引用的表和列，
	//  4，5 元素为 UPDATE 和 DELETE 的规则定义；
	// 其它模式下为该约束关联的列名称。
	Data []string
}

// AddConstraint 声明添加约束的语句
func AddConstraint(e Engine, d Dialect) *AddConstraintStmt {
	stmt := &AddConstraintStmt{}
	stmt.ddlStmt = newDDLStmt(e, d, stmt)
	return stmt
}

// Reset 重置内容
func (stmt *AddConstraintStmt) Reset() *AddConstraintStmt {
	stmt.TableName = ""
	stmt.Name = ""
	stmt.Type = constraintNone
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
	if stmt.Type != constraintNone {
		panic(ErrConstraintType)
	}

	stmt.Type = ConstraintUnique
	stmt.Name = name
	stmt.Data = col

	return stmt
}

// PK 指定主键约束
func (stmt *AddConstraintStmt) PK(col ...string) *AddConstraintStmt {
	if stmt.Type != constraintNone {
		panic(ErrConstraintType)
	}

	stmt.Type = ConstraintPK
	stmt.Data = col

	return stmt
}

// Check Check 约束
func (stmt *AddConstraintStmt) Check(name, expr string) *AddConstraintStmt {
	if stmt.Type != constraintNone {
		panic(ErrConstraintType)
	}

	stmt.Type = ConstraintCheck
	stmt.Name = name
	stmt.Data = []string{expr}

	return stmt
}

// FK 外键约束
func (stmt *AddConstraintStmt) FK(name, col, refTable, refColumn, updateRule, deleteRule string) *AddConstraintStmt {
	if stmt.Type != constraintNone {
		panic(ErrConstraintType)
	}

	stmt.Type = ConstraintFK
	stmt.Name = name
	stmt.Data = []string{col, refTable, refColumn, updateRule, deleteRule}

	return stmt
}

// DDLSQL 生成 SQL 语句
func (stmt *AddConstraintStmt) DDLSQL() ([]string, error) {
	if stmt.TableName == "" {
		return nil, ErrTableIsEmpty
	}

	if len(stmt.Data) == 0 {
		return nil, ErrColumnsIsEmpty
	}

	if stmt.Type == ConstraintPK {
		stmt.Name = PKName(stmt.TableName)
	}

	if stmt.Name == "" {
		return nil, ErrConstraintIsEmpty
	}

	if hook, ok := stmt.dialect.(AddConstraintStmtHooker); ok {
		return hook.AddConstraintStmtHook(stmt)
	}

	builder := New("ALTER TABLE ").
		Quote(stmt.TableName, stmt.l, stmt.r).
		WriteString(" ADD CONSTRAINT ").
		Quote(stmt.Name, stmt.l, stmt.r)

	switch stmt.Type {
	case ConstraintCheck:
		builder.WriteString(" CHECK(").WriteString(stmt.Data[0]).WriteBytes(')')
	case ConstraintFK:
		builder.WriteString(" FOREIGN KEY(").
			Quote(stmt.Data[0], stmt.l, stmt.r).
			WriteString(") REFERENCES ").
			Quote(stmt.Data[1], stmt.l, stmt.r).
			WriteBytes('(').
			Quote(stmt.Data[2], stmt.l, stmt.r).
			WriteBytes(')')

		if stmt.Data[3] != "" {
			builder.WriteString(" ON UPDATE ").WriteString(stmt.Data[3])
		}

		if stmt.Data[4] != "" {
			builder.WriteString(" ON DELETE ").WriteString(stmt.Data[4])
		}
	case ConstraintPK:
		builder.WriteString(" PRIMARY KEY(")
		for _, col := range stmt.Data {
			builder.
				Quote(col, stmt.l, stmt.r).
				WriteBytes(',')
		}
		builder.TruncateLast(1).WriteBytes(')')
	case ConstraintUnique:
		builder.WriteString(" UNIQUE(")
		for _, col := range stmt.Data {
			builder.
				Quote(col, stmt.l, stmt.r).
				WriteBytes(',')
		}
		builder.TruncateLast(1).WriteBytes(')')
	default:
		return nil, ErrUnknownConstraint
	}

	return []string{builder.String()}, nil
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
func DropConstraint(e Engine, d Dialect) *DropConstraintStmt {
	stmt := &DropConstraintStmt{}
	stmt.ddlStmt = newDDLStmt(e, d, stmt)
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
	if stmt.TableName == "" {
		return nil, ErrTableIsEmpty
	}

	if stmt.Name == "" {
		return nil, ErrValueIsEmpty
	}

	if hook, ok := stmt.dialect.(DropConstraintStmtHooker); ok {
		return hook.DropConstraintStmtHook(stmt)
	}

	buf := New("ALTER TABLE ").
		Quote(stmt.TableName, stmt.l, stmt.r).
		WriteString(" DROP CONSTRAINT ").
		Quote(stmt.Name, stmt.l, stmt.r)
	return []string{buf.String()}, nil
}

// Reset 重置
func (stmt *DropConstraintStmt) Reset() *DropConstraintStmt {
	stmt.TableName = ""
	stmt.Name = ""
	return stmt
}
