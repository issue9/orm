// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"slices"

	"github.com/issue9/orm/v6/core"
	"github.com/issue9/orm/v6/sqlbuilder"
)

// ErrNeedAutoIncrementColumn 缺少必要的自增列
//
// 当以 LastInsertID 的方式插入一条没有 AI 列的对象时，会返回此错误。
var ErrNeedAutoIncrementColumn = errors.New("必须存在自增列")

func (db *DB) newModel(obj TableNamer) (*core.Model, error) { return db.models.New(obj) }

func (tx *Tx) newModel(obj TableNamer) (*core.Model, error) { return tx.db.models.New(obj) }

func (e *txEngine) newModel(obj TableNamer) (*core.Model, error) { return e.tx.newModel(obj) }

func getModel(e Engine, v TableNamer) (*core.Model, reflect.Value, error) {
	m, err := e.newModel(v)
	if err != nil {
		return nil, reflect.Value{}, err
	}

	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	return m, rval, nil
}

// 根据 Model 中的主键或是唯一索引生成 where 语句，若两者都不存在，则返回错误信息。
func where(ws *sqlbuilder.WhereStmt, m *core.Model, rval reflect.Value) error {
	var keys []string
	var vals []any
	var constraint string

	if m.AutoIncrement != nil {
		if keys, vals = getKV(rval, m.AutoIncrement); len(keys) > 0 {
			goto RET
		}
	}

	if m.PrimaryKey != nil {
		if keys, vals = getKV(rval, m.PrimaryKey.Columns...); len(keys) > 0 {
			goto RET
		}
	}

	for _, u := range m.Uniques {
		k, v := getKV(rval, u.Columns...)
		if len(k) == 0 {
			continue
		}

		if len(keys) > 0 {
			// 可能每个唯一约束查询至的结果是不一样的
			return fmt.Errorf("多个唯一约束 %s、%s 满足查询条件", constraint, u.Name)
		}

		keys, vals = k, v
		constraint = u.Name
	}

RET:
	if len(keys) == 0 || len(vals) == 0 {
		return fmt.Errorf("可作为唯一条件的自增、主键和唯一约束都为空值，无法为 %s 生成查询条件", m.Name)
	}

	for index, key := range keys {
		ws.And(string(core.QuoteLeft)+key+string(core.QuoteRight)+"=?", vals[index])
	}

	return nil
}

func getKV(rval reflect.Value, cols ...*core.Column) (keys []string, vals []any) {
	for _, col := range cols {
		field := rval.FieldByName(col.GoName)

		if field.IsZero() {
			return nil, nil
		}

		keys = append(keys, col.Name)
		vals = append(vals, field.Interface())
	}
	return keys, vals
}

// 创建表或是视图
func create(ctx context.Context, e Engine, v TableNamer) error {
	m, _, err := getModel(e, v)
	if err != nil {
		return err
	}

	if m.Type == core.View {
		return createView(ctx, e, m)
	}

	sb := e.SQLBuilder().CreateTable().Table(m.Name)
	for _, col := range m.Columns {
		if col.AI {
			sb.AutoIncrement(col.Name, col.PrimitiveType)
		} else {
			sb.Columns(col.Clone())
		}
	}

	for _, index := range m.Indexes {
		cols := make([]string, 0, len(index.Columns))
		for _, col := range index.Columns {
			cols = append(cols, col.Name)
		}
		sb.Index(core.IndexDefault, constraintName(m.Name, index.Name), cols...)
	}

	for _, unique := range m.Uniques {
		cols := make([]string, 0, len(unique.Columns))
		for _, col := range unique.Columns {
			cols = append(cols, col.Name)
		}
		sb.Unique(constraintName(m.Name, unique.Name), cols...)
	}

	for name, expr := range m.Checks {
		sb.Check(constraintName(m.Name, name), expr)
	}

	for _, fk := range m.ForeignKeys {
		name := constraintName(m.Name, fk.Name)
		sb.ForeignKey(name, fk.Column.Name, fk.RefTableName, fk.RefColName, fk.UpdateRule, fk.DeleteRule)
	}

	if m.AutoIncrement == nil && m.PrimaryKey != nil {
		cols := make([]string, 0, len(m.PrimaryKey.Columns))
		for _, col := range m.PrimaryKey.Columns {
			cols = append(cols, col.Name)
		}
		sb.PK(constraintName(m.Name, m.PrimaryKey.Name), cols...)
	}

	return sb.ExecContext(ctx)
}

func createView(ctx context.Context, e Engine, m *core.Model) error {
	stmt := e.SQLBuilder().CreateView().Name(m.Name)

	for _, col := range m.Columns {
		stmt.Column(col.Name)
	}
	return stmt.FromQuery(m.ViewAs).ExecContext(ctx)
}

func truncate(ctx context.Context, e Engine, v TableNamer) error {
	m, err := e.newModel(v)
	if err != nil {
		return err
	}

	if m.Type == core.View {
		return fmt.Errorf("模型 %s 的类型是视图，无法从其中删除数据", m.Name)
	}

	stmt := e.SQLBuilder().TruncateTable()
	if m.AutoIncrement != nil {
		stmt.Table(m.Name, constraintName(m.Name, m.AutoIncrement.Name))
	} else {
		stmt.Table(m.Name, "")
	}

	return stmt.ExecContext(ctx)
}

// 删除表或视图
func drop(ctx context.Context, e Engine, v TableNamer) error {
	m, err := e.newModel(v)
	if err != nil {
		return err
	}

	if m.Type == core.View {
		return e.SQLBuilder().DropView().Name(m.Name).ExecContext(ctx)
	}

	return e.SQLBuilder().DropTable().Table(m.Name).ExecContext(ctx)
}

func lastInsertID(ctx context.Context, e Engine, v TableNamer) (int64, error) {
	m, rval, err := getModel(e, v)
	if err != nil {
		return 0, err
	}

	if m.Type == core.View {
		return 0, fmt.Errorf("模型 %s 的类型是视图，无法添中数据", m.Name)
	}

	if m.AutoIncrement == nil {
		return 0, ErrNeedAutoIncrementColumn
	}

	if obj, ok := v.(BeforeInserter); ok {
		if err = obj.BeforeInsert(); err != nil {
			return 0, err
		}
	}

	stmt := e.SQLBuilder().Insert().Table(m.Name)
	for _, col := range m.Columns {
		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return 0, fmt.Errorf("未找到该名称 %s 的值", col.GoName)
		}

		if col.AI && !field.IsZero() {
			return 0, fmt.Errorf("自增列 %s 不允许指定值", col.Name)
		}

		// 自增或是含有默认值的零值列。
		if col.AI || (field.IsZero() && col.HasDefault) {
			continue
		}

		stmt.KeyValue(col.Name, field.Interface())
	}

	return stmt.LastInsertIDContext(ctx, m.AutoIncrement.Name)
}

func insert(ctx context.Context, e Engine, v TableNamer) (sql.Result, error) {
	m, rval, err := getModel(e, v)
	if err != nil {
		return nil, err
	}

	if m.Type == core.View {
		return nil, fmt.Errorf("模型 %s 的类型是视图，无法添中数据", m.Name)
	}

	if obj, ok := v.(BeforeInserter); ok {
		if err = obj.BeforeInsert(); err != nil {
			return nil, err
		}
	}

	stmt := e.SQLBuilder().Insert().Table(m.Name)
	for _, col := range m.Columns {
		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return nil, fmt.Errorf("未找到该名称 %s 的值", col.GoName)
		}

		if col.AI && !field.IsZero() {
			return nil, fmt.Errorf("自增列 %s 不允许指定值", col.Name)
		}

		// 自增或是含有默认值的零值列。
		if col.AI || (field.IsZero() && col.HasDefault) {
			continue
		}

		stmt.KeyValue(col.Name, field.Interface())
	}

	return stmt.ExecContext(ctx)
}

// 查找数据
//
// 根据 v 的 pk 或中唯一索引列查找一行数据，并赋值给 v。
// 若 v 为空，则不发生任何操作，v 可以是数组。
func find(ctx context.Context, e Engine, v TableNamer) (bool, error) {
	m, rval, err := getModel(e, v)
	if err != nil {
		return false, err
	}

	stmt := e.SQLBuilder().Select().Column("*").From(m.Name)
	if err = where(stmt.WhereStmt(), m, rval); err != nil {
		return false, err
	}

	size, err := stmt.QueryObjectContext(ctx, true, v)
	if err != nil {
		return false, err
	}
	return size > 0, nil
}

// for update 只能作用于事务
func forUpdate(ctx context.Context, tx *Tx, v TableNamer) error {
	m, rval, err := getModel(tx, v)
	if err != nil {
		return err
	}

	if m.Type == core.View {
		return fmt.Errorf("模型 %s 的类型是视图，无法更新其数据", m.Name)
	}

	if obj, ok := v.(BeforeUpdater); ok {
		if err = obj.BeforeUpdate(); err != nil {
			return err
		}
	}

	stmt := tx.SQLBuilder().Select().Column("*").From(m.Name).ForUpdate()
	if err = where(stmt.WhereStmt(), m, rval); err != nil {
		return err
	}

	_, err = stmt.QueryObjectContext(ctx, true, v)
	return err
}

// 更新 v 到数据库，默认情况下不更新零值。
// cols 表示必须要更新的列，即使是零值。
//
// 更新依据为每个对象的主键或是唯一索引列。
// 若不存在此两个类型的字段，则返回错误信息。
func update(ctx context.Context, e Engine, v TableNamer, cols ...string) (sql.Result, error) {
	stmt := e.SQLBuilder().Update()

	m, rval, err := getUpdateColumns(e, v, stmt, cols...)
	if err != nil {
		return nil, err
	}

	if err := where(stmt.WhereStmt(), m, rval); err != nil {
		return nil, err
	}

	return stmt.ExecContext(ctx)
}

func save(ctx context.Context, e Engine, v TableNamer, cols ...string) (int64, bool, error) {
	if found, err := e.SelectContext(ctx, v); err != nil || !found {
		id, err := lastInsertID(ctx, e, v)
		return id, true, err
	}

	_, err := update(ctx, e, v, cols...)
	return 0, false, err
}

func getUpdateColumns(e Engine, v TableNamer, stmt *sqlbuilder.UpdateStmt, cols ...string) (*core.Model, reflect.Value, error) {
	m, rval, err := getModel(e, v)
	if err != nil {
		return nil, reflect.Value{}, err
	}
	if m.Type == core.View {
		return nil, reflect.Value{}, fmt.Errorf("模型 %s 的类型是视图，无法更新其数据", m.Name)
	}

	if obj, ok := v.(BeforeUpdater); ok {
		if err = obj.BeforeUpdate(); err != nil {
			return nil, reflect.Value{}, err
		}
	}

	var occValue any
	for _, col := range m.Columns {
		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return nil, reflect.Value{}, fmt.Errorf("未找到该名称 %s 的值", col.GoName)
		}

		if m.OCC == col { // 乐观锁
			occValue = field.Interface()
		} else if slices.Index(cols, col.Name) >= 0 || !field.IsZero() {
			// 非零值或是明确指定需要更新的列，才会更新
			stmt.Set(col.Name, field.Interface())
		}
	}

	if m.OCC != nil {
		stmt.OCC(m.OCC.Name, occValue)
	}

	stmt.Table(m.Name)

	return m, rval, nil
}

// 将 v 生成 delete 的 sql 语句
func del(ctx context.Context, e Engine, v TableNamer) (sql.Result, error) {
	m, rval, err := getModel(e, v)
	if err != nil {
		return nil, err
	}

	if m.Type == core.View {
		return nil, fmt.Errorf("模型 %s 的类型是视图，无法从其中删除数据", m.Name)
	}

	stmt := e.SQLBuilder().Delete().Table(m.Name)
	if err = where(stmt.WhereStmt(), m, rval); err != nil {
		return nil, err
	}

	return stmt.ExecContext(ctx)
}

var errInsertManyHasDifferentType = errors.New("InsertMany 必须是相同的数据类型")

// rval 为结构体指针组成的数据
func buildInsertManySQL(e Engine, v ...TableNamer) (*sqlbuilder.InsertStmt, error) {
	query := e.SQLBuilder().Insert()
	if len(v) == 0 {
		return query, nil
	}

	var keys []string          // 保存列的顺序，方便后续元素获取值
	var firstType reflect.Type // 记录数组中第一个元素的类型，保证后面的都相同

	for i := 0; i < len(v); i++ {
		if obj, ok := v[i].(BeforeInserter); ok {
			if err := obj.BeforeInsert(); err != nil {
				return nil, err
			}
		}

		m, irval, err := getModel(e, v[i])
		if err != nil {
			return nil, err
		}

		if i == 0 { // 第一个元素，需要从中获取列信息。
			firstType = irval.Type()
			query.Table(m.Name)

			for _, col := range m.Columns {
				field := irval.FieldByName(col.GoName)
				if !field.IsValid() {
					return nil, fmt.Errorf("未找到该名称 %s 的值", col.GoName)
				}

				// 在为零值的情况下，若该列是 AI 或是有默认值，则过滤掉。无论该零值是否为手动设置的。
				if field.IsZero() && (col.AI || col.HasDefault) {
					continue
				}

				query.KeyValue(col.Name, field.Interface())
				keys = append(keys, col.Name)
			}
		} else { // 之后的元素，只需要获取其对应的值就行
			if firstType != irval.Type() { // 与第一个元素的类型不同。
				return nil, errInsertManyHasDifferentType
			}

			vals := make([]any, 0, len(keys))
			for _, name := range keys {
				col := m.FindColumn(name)
				if col == nil {
					return nil, core.ErrColumnNotFound(name)
				}

				field := irval.FieldByName(col.GoName)
				if !field.IsValid() {
					return nil, fmt.Errorf("未找到该名称 %s 的值", col.GoName)
				}

				// 在为零值的情况下，若该列是 AI 或是有默认值，则过滤掉。无论该零值是否为手动设置的。
				if field.IsZero() && (col.AI || col.HasDefault) {
					continue
				}

				vals = append(vals, field.Interface())
			}
			query.Values(vals...)
		}
	}

	return query, nil
}

func constraintName(table, name string) string { return table + "_" + name }
