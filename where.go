// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"database/sql"
	"fmt"

	"github.com/issue9/orm/v2/core"
	"github.com/issue9/orm/v2/sqlbuilder"
)

// WhereStmt 用于生成 where 语句
type WhereStmt struct {
	where  *sqlbuilder.WhereStmt
	engine Engine
}

// Where 生成 Where 语句
func (db *DB) Where(cond string, args ...interface{}) *WhereStmt {
	return &WhereStmt{
		where:  sqlbuilder.Where(),
		engine: db,
	}
}

// Where 生成 Where 语句
func (tx *Tx) Where(cond string, args ...interface{}) *WhereStmt {
	return &WhereStmt{
		where:  sqlbuilder.Where(),
		engine: tx,
	}
}

// And 添加一条 and 语句
func (stmt *WhereStmt) And(cond string, args ...interface{}) *WhereStmt {
	stmt.where.And(cond, args...)
	return stmt
}

// Or 添加一条 OR 语句
func (stmt *WhereStmt) Or(cond string, args ...interface{}) *WhereStmt {
	stmt.where.Or(cond, args)
	return stmt
}

// AndIsNull 指定 WHERE ... AND col IS NULL
func (stmt *WhereStmt) AndIsNull(col string) *WhereStmt {
	stmt.where.AndIsNull(col)
	return stmt
}

// OrIsNull 指定 WHERE ... OR col IS NULL
func (stmt *WhereStmt) OrIsNull(col string) *WhereStmt {
	stmt.where.OrIsNull(col)
	return stmt
}

// AndIsNotNull 指定 WHERE ... AND col IS NOT NULL
func (stmt *WhereStmt) AndIsNotNull(col string) *WhereStmt {
	stmt.where.AndIsNotNull(col)
	return stmt
}

// OrIsNotNull 指定 WHERE ... OR col IS NOT NULL
func (stmt *WhereStmt) OrIsNotNull(col string) *WhereStmt {
	stmt.where.OrIsNotNull(col)
	return stmt
}

// AndBetween 指定 WHERE ... AND col BETWEEN v1 AND v2
func (stmt *WhereStmt) AndBetween(col string, v1, v2 interface{}) *WhereStmt {
	stmt.where.AndBetween(col, v1, v2)
	return stmt
}

// OrBetween 指定 WHERE ... OR col BETWEEN v1 AND v2
func (stmt *WhereStmt) OrBetween(col string, v1, v2 interface{}) *WhereStmt {
	stmt.where.OrBetween(col, v1, v2)
	return stmt
}

// AndNotBetween 指定 WHERE ... AND col NOT BETWEEN v1 AND v2
func (stmt *WhereStmt) AndNotBetween(col string, v1, v2 interface{}) *WhereStmt {
	stmt.where.AndNotBetween(col, v1, v2)
	return stmt
}

// OrNotBetween 指定 WHERE ... OR col BETWEEN v1 AND v2
func (stmt *WhereStmt) OrNotBetween(col string, v1, v2 interface{}) *WhereStmt {
	stmt.where.OrNotBetween(col, v1, v2)
	return stmt
}

// AndLike 指定 WHERE ... AND col LIKE content
func (stmt *WhereStmt) AndLike(col string, content interface{}) *WhereStmt {
	stmt.where.AndLike(col, content)
	return stmt
}

// OrLike 指定 WHERE ... OR col LIKE content
func (stmt *WhereStmt) OrLike(col string, content interface{}) *WhereStmt {
	stmt.where.OrLike(col, content)
	return stmt
}

// AndNotLike 指定 WHERE ... AND col NOT LIKE content
func (stmt *WhereStmt) AndNotLike(col string, content interface{}) *WhereStmt {
	stmt.where.AndNotLike(col, content)
	return stmt
}

// OrNotLike 指定 WHERE ... OR col NOT LIKE content
func (stmt *WhereStmt) OrNotLike(col string, content interface{}) *WhereStmt {
	stmt.where.OrNotLike(col, content)
	return stmt
}

// AndIn 指定 WHERE ... AND col IN(v...)
func (stmt *WhereStmt) AndIn(col string, v ...interface{}) *WhereStmt {
	stmt.where.AndIn(col, v...)
	return stmt
}

// OrIn 指定 WHERE ... OR col IN(v...)
func (stmt *WhereStmt) OrIn(col string, v ...interface{}) *WhereStmt {
	stmt.where.OrIn(col, v...)
	return stmt
}

// AndNotIn 指定 WHERE ... AND col NOT IN(v...)
func (stmt *WhereStmt) AndNotIn(col string, v ...interface{}) *WhereStmt {
	stmt.where.AndNotIn(col, v...)
	return stmt
}

// OrNotIn 指定 WHERE ... OR col IN(v...)
func (stmt *WhereStmt) OrNotIn(col string, v ...interface{}) *WhereStmt {
	stmt.where.OrNotIn(col, v...)
	return stmt
}

// AndGroup 开始一个子条件语句
func (stmt *WhereStmt) AndGroup() *sqlbuilder.WhereStmt {
	return stmt.where.AndGroup()
}

// OrGroup 开始一个子条件语句
func (stmt *WhereStmt) OrGroup() *sqlbuilder.WhereStmt {
	return stmt.where.OrGroup()
}

// 从 v 中读取表名，该删除该表中符合条件的所有内容
func (stmt *WhereStmt) Delete(v interface{}) (sql.Result, error) {
	m, err := stmt.engine.NewModel(v)
	if err != nil {
		return nil, err
	}

	if m.Type != core.View {
		return nil, fmt.Errorf("模型 %s 的类型是视图，无法从其中删除数据", m.Name)
	}

	return stmt.where.Delete(stmt.engine).Table(m.Name).Exec()
}

// 将 v 中的所有值更新到表中，包括零值
func (stmt *WhereStmt) Update(v interface{}) (sql.Result, error) {
	m, rval, err := getModel(stmt.engine, v)
	if err != nil {
		return nil, err
	}

	if obj, ok := v.(BeforeUpdater); ok {
		if err = obj.BeforeUpdate(); err != nil {
			return nil, err
		}
	}

	upd := stmt.where.Update(stmt.engine).Table(m.Name)

	var occValue interface{}
	for _, col := range m.Columns {
		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return nil, fmt.Errorf("未找到该名称 %s 的值", col.GoName)
		}

		if m.OCC == col { // 乐观锁
			occValue = field.Interface()
		} else {
			// 非零值或是明确指定需要更新的列，才会更新
			upd.Set(col.Name, field.Interface())
		}
	}

	if m.OCC != nil {
		upd.OCC(m.OCC.Name, occValue)
	}

	return upd.Exec()
}

// Select 获取所有符合条件的数据
//
// 参数可以参考 fetch.Object() 中的说明
func (stmt *WhereStmt) Select(strict bool, v interface{}) (int, error) {
	m, err := stmt.engine.NewModel(v)
	if err != nil {
		return 0, err
	}

	if m.Type != core.View {
		return 0, fmt.Errorf("模型 %s 的类型是视图，无法从其中删除数据", m.Name)
	}

	return stmt.where.Select(stmt.engine).QueryObject(strict, v)
}
