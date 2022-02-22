// SPDX-License-Identifier: MIT

package orm

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/issue9/orm/v4/core"
	"github.com/issue9/orm/v4/sqlbuilder"
)

// WhereStmt 用于生成 where 语句
type WhereStmt struct {
	where  *sqlbuilder.WhereStmt
	engine Engine
}

// Where 生成 Where 语句
func (db *DB) Where(cond string, args ...any) *WhereStmt {
	return &WhereStmt{
		where:  sqlbuilder.Where().And(cond, args...),
		engine: db,
	}
}

// Where 生成 Where 语句
func (tx *Tx) Where(cond string, args ...any) *WhereStmt {
	return &WhereStmt{
		where:  sqlbuilder.Where().And(cond, args...),
		engine: tx,
	}
}

// And 添加一条 and 语句
func (stmt *WhereStmt) And(cond string, args ...any) *WhereStmt {
	stmt.where.And(cond, args...)
	return stmt
}

// Or 添加一条 OR 语句
func (stmt *WhereStmt) Or(cond string, args ...any) *WhereStmt {
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
func (stmt *WhereStmt) AndBetween(col string, v1, v2 any) *WhereStmt {
	stmt.where.AndBetween(col, v1, v2)
	return stmt
}

// OrBetween 指定 WHERE ... OR col BETWEEN v1 AND v2
func (stmt *WhereStmt) OrBetween(col string, v1, v2 any) *WhereStmt {
	stmt.where.OrBetween(col, v1, v2)
	return stmt
}

// AndNotBetween 指定 WHERE ... AND col NOT BETWEEN v1 AND v2
func (stmt *WhereStmt) AndNotBetween(col string, v1, v2 any) *WhereStmt {
	stmt.where.AndNotBetween(col, v1, v2)
	return stmt
}

// OrNotBetween 指定 WHERE ... OR col BETWEEN v1 AND v2
func (stmt *WhereStmt) OrNotBetween(col string, v1, v2 any) *WhereStmt {
	stmt.where.OrNotBetween(col, v1, v2)
	return stmt
}

// AndLike 指定 WHERE ... AND col LIKE content
func (stmt *WhereStmt) AndLike(col string, content any) *WhereStmt {
	stmt.where.AndLike(col, content)
	return stmt
}

// OrLike 指定 WHERE ... OR col LIKE content
func (stmt *WhereStmt) OrLike(col string, content any) *WhereStmt {
	stmt.where.OrLike(col, content)
	return stmt
}

// AndNotLike 指定 WHERE ... AND col NOT LIKE content
func (stmt *WhereStmt) AndNotLike(col string, content any) *WhereStmt {
	stmt.where.AndNotLike(col, content)
	return stmt
}

// OrNotLike 指定 WHERE ... OR col NOT LIKE content
func (stmt *WhereStmt) OrNotLike(col string, content any) *WhereStmt {
	stmt.where.OrNotLike(col, content)
	return stmt
}

// AndIn 指定 WHERE ... AND col IN(v...)
func (stmt *WhereStmt) AndIn(col string, v ...any) *WhereStmt {
	stmt.where.AndIn(col, v...)
	return stmt
}

// OrIn 指定 WHERE ... OR col IN(v...)
func (stmt *WhereStmt) OrIn(col string, v ...any) *WhereStmt {
	stmt.where.OrIn(col, v...)
	return stmt
}

// AndNotIn 指定 WHERE ... AND col NOT IN(v...)
func (stmt *WhereStmt) AndNotIn(col string, v ...any) *WhereStmt {
	stmt.where.AndNotIn(col, v...)
	return stmt
}

// OrNotIn 指定 WHERE ... OR col IN(v...)
func (stmt *WhereStmt) OrNotIn(col string, v ...any) *WhereStmt {
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

// Delete 从 v 表中删除符合条件的内容
func (stmt *WhereStmt) Delete(v TableNamer) (sql.Result, error) {
	m, err := stmt.engine.NewModel(v)
	if err != nil {
		return nil, err
	}

	if m.Type == core.View {
		return nil, fmt.Errorf("模型 %s 的类型是视图，无法从其中删除数据", m.Name)
	}

	return stmt.where.Delete(stmt.engine).Table(m.Name).Exec()
}

// Update 将 v 中内容更新到符合条件的行中
//
// 不会更新零值，除非通过 cols 指定了该列。
// 表名来自 v，列名为 v 的所有列或是 cols 指定的列。
func (stmt *WhereStmt) Update(v TableNamer, cols ...string) (sql.Result, error) {
	upd := stmt.where.Update(stmt.engine)

	if _, _, err := getUpdateColumns(stmt.engine, v, upd, cols...); err != nil {
		return nil, err
	}

	return upd.Exec()
}

// Select 获取所有符合条件的数据
//
// v 可能是某个对象的指针，或是一组相同对象指针数组。
// 表名来自 v，列名为 v 的所有列。
func (stmt *WhereStmt) Select(strict bool, v any) (int, error) {
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		t = t.Elem()
	}

	tn, ok := reflect.New(t).Interface().(TableNamer)
	if !ok {
		return 0, fmt.Errorf("v 不是 TableNamer 类型")
	}
	m, err := stmt.engine.NewModel(tn)
	if err != nil {
		return 0, err
	}

	return stmt.where.Select(stmt.engine).
		Column("*").
		From(m.Name).
		QueryObject(strict, v)
}

// Count 返回符合条件数量
//
// 表名来自 v。
func (stmt *WhereStmt) Count(v TableNamer) (int64, error) {
	m, _, err := getModel(stmt.engine, v)
	if err != nil {
		return 0, err
	}

	return stmt.where.Select(stmt.engine).
		Count("count(*) as cnt").
		From(m.Name).
		QueryInt("cnt")
}
