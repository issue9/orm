// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package orm

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/issue9/orm/v6/core"
	"github.com/issue9/orm/v6/sqlbuilder"
)

type whereWhere = sqlbuilder.WhereStmtOf[*WhereStmt]

type WhereStmt struct {
	*whereWhere
	engine Engine
}

func (db *DB) Where(cond string, args ...any) *WhereStmt {
	w := &WhereStmt{engine: db}
	w.whereWhere = sqlbuilder.NewWhereStmtOf(w)
	return w.Where(cond, args...)
}

func (tx *Tx) Where(cond string, args ...any) *WhereStmt {
	w := &WhereStmt{engine: tx}
	w.whereWhere = sqlbuilder.NewWhereStmtOf(w)
	return w.Where(cond, args...)
}

func (p *dbPrefix) Where(cond string, args ...any) *WhereStmt {
	w := &WhereStmt{engine: p}
	w.whereWhere = sqlbuilder.NewWhereStmtOf(w)
	return w.Where(cond, args...)
}

func (p *txPrefix) Where(cond string, args ...any) *WhereStmt {
	w := &WhereStmt{engine: p}
	w.whereWhere = sqlbuilder.NewWhereStmtOf(w)
	return w.Where(cond, args...)
}

func (stmt *WhereStmt) TablePrefix() string { return stmt.engine.TablePrefix() }

// Delete 从 v 表中删除符合条件的内容
func (stmt *WhereStmt) Delete(v TableNamer) (sql.Result, error) {
	m, err := stmt.engine.newModel(v)
	if err != nil {
		return nil, err
	}

	if m.Type == core.View {
		return nil, fmt.Errorf("模型 %s 的类型是视图，无法从其中删除数据", m.Name)
	}

	return stmt.WhereStmt().Delete(stmt.engine).Table(m.Name).Exec()
}

// Update 将 v 中内容更新到符合条件的行中
//
// 不会更新零值，除非通过 cols 指定了该列。
// 表名来自 v，列名为 v 的所有列或是 cols 指定的列。
func (stmt *WhereStmt) Update(v TableNamer, cols ...string) (sql.Result, error) {
	upd := stmt.WhereStmt().Update(stmt.engine)

	if _, _, err := getUpdateColumns(stmt.engine, v, upd, cols...); err != nil {
		return nil, err
	}

	return upd.Exec()
}

// Select 获取所有符合条件的数据
//
// v 可能是某个对象的指针，或是一组相同对象指针数组。表名来自 v，列名为 v 的所有列。
func (stmt *WhereStmt) Select(strict bool, v any) (int, error) {
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		t = t.Elem()
	}

	tn, ok := reflect.New(t).Interface().(TableNamer)
	if !ok {
		return 0, fmt.Errorf("v 不是 TableNamer 类型")
	}
	m, err := stmt.engine.newModel(tn)
	if err != nil {
		return 0, err
	}

	return stmt.WhereStmt().Select(stmt.engine).
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

	return stmt.WhereStmt().Select(stmt.engine).
		Count("count(*) as cnt").
		From(m.Name).
		QueryInt("cnt")
}
