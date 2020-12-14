// SPDX-License-Identifier: MIT

package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/issue9/orm/v3/core"
	"github.com/issue9/orm/v3/sqlbuilder"
)

// ErrNeedAutoIncrementColumn 当以 LastInsertID
// 的方式插入一条没有 AI 列的对象时，会返回此错误。
var ErrNeedAutoIncrementColumn = errors.New("必须存在自增列")

func getModel(e Engine, v interface{}) (*Model, reflect.Value, error) {
	m, err := e.NewModel(v)
	if err != nil {
		return nil, reflect.Value{}, err
	}

	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	return m, rval, nil
}

// 根据 Model 中的主键或是唯一索引为 sql 产生 where 语句，
// 若两者都不存在，则返回错误信息。rval 为 struct 的 reflect.Value
func where(sb sqlbuilder.WhereStmter, m *Model, rval reflect.Value) error {
	var keys []string
	var vals []interface{}
	var constraint string

	if m.AutoIncrement != nil {
		if keys, vals = getKV(rval, m.AutoIncrement); len(keys) > 0 {
			goto RET
		}
	}

	if len(m.PrimaryKey) > 0 {
		if keys, vals = getKV(rval, m.PrimaryKey...); len(keys) > 0 {
			goto RET
		}
	}

	for name, cols := range m.Uniques {
		k, v := getKV(rval, cols...)
		if len(k) == 0 {
			continue
		}

		if len(keys) > 0 {
			return fmt.Errorf("多个唯一约束 %s、%s 满足查询条件", constraint, name)
		}

		keys, vals = k, v
		constraint = name
	}

RET:
	if len(keys) == 0 || len(vals) == 0 {
		return fmt.Errorf("可作为唯一条件的自增、主键和唯一约束都为空值，无法为 %s 生成查询条件", m.Name)
	}

	for index, key := range keys {
		sb.WhereStmt().And("{"+key+"}=?", vals[index])
	}

	return nil
}

func getKV(rval reflect.Value, cols ...*core.Column) (keys []string, vals []interface{}) {
	for _, col := range cols {
		field := rval.FieldByName(col.GoName)

		if field.IsZero() {
			vals = vals[:0]
			keys = keys[:0]
			return nil, nil
		}

		keys = append(keys, col.Name)
		vals = append(vals, field.Interface())
	}
	return keys, vals
}

// 创建表或是视图。
func create(e Engine, v interface{}) error {
	m, _, err := getModel(e, v)
	if err != nil {
		return err
	}

	if m.Type == core.View {
		return createView(e, m)
	}

	sb := sqlbuilder.CreateTable(e)
	sb.Table(m.Name)
	for _, col := range m.Columns {
		if col.AI {
			sb.AutoIncrement(col.Name, col.PrimitiveType)
		} else {
			sb.Columns(col.Clone())
		}
	}

	for name, index := range m.Indexes {
		cols := make([]string, 0, len(index))
		for _, col := range index {
			cols = append(cols, col.Name)
		}
		sb.Index(core.IndexDefault, name, cols...)
	}

	for name, unique := range m.Uniques {
		cols := make([]string, 0, len(unique))
		for _, col := range unique {
			cols = append(cols, col.Name)
		}
		sb.Unique(name, cols...)
	}

	for name, expr := range m.Checks {
		sb.Check(name, expr)
	}

	for name, fk := range m.ForeignKeys {
		sb.ForeignKey(name, fk.Column.Name, fk.RefTableName, fk.RefColName, fk.UpdateRule, fk.DeleteRule)
	}

	if m.AutoIncrement == nil && len(m.PrimaryKey) > 0 {
		cols := make([]string, 0, len(m.PrimaryKey))
		for _, col := range m.PrimaryKey {
			cols = append(cols, col.Name)
		}
		sb.PK(cols...)
	}

	return sb.Exec()
}

func createView(e Engine, m *core.Model) error {
	stmt := sqlbuilder.CreateView(e).Name(m.Name)

	for _, col := range m.Columns {
		stmt.Column(col.Name)
	}
	stmt.SelectQuery = m.ViewAs
	return stmt.Exec()
}

func truncate(e Engine, v interface{}) error {
	m, err := e.NewModel(v)
	if err != nil {
		return err
	}

	if m.Type == core.View {
		return fmt.Errorf("模型 %s 的类型是视图，无法从其中删除数据", m.Name)
	}

	stmt := sqlbuilder.TruncateTable(e)
	if m.AutoIncrement != nil {
		stmt.Table(m.Name, m.AutoIncrement.Name)
	} else {
		stmt.Table(m.Name, "")
	}

	return stmt.Exec()
}

// 删除一张表或视图。
func drop(e Engine, v interface{}) error {
	m, err := e.NewModel(v)
	if err != nil {
		return err
	}

	if m.Type == core.View {
		return sqlbuilder.DropView(e).Name(m.Name).Exec()
	}

	return sqlbuilder.DropTable(e).Table(m.Name).Exec()
}

func lastInsertID(e Engine, v interface{}) (int64, error) {
	m, rval, err := getModel(e, v)
	if err != nil {
		return 0, err
	}

	if m.Type == core.View {
		return 0, fmt.Errorf("模型 %s 的类型是视图，无法从其中删除数据", m.Name)
	}

	if m.AutoIncrement == nil {
		return 0, ErrNeedAutoIncrementColumn
	}

	if obj, ok := v.(BeforeInserter); ok {
		if err = obj.BeforeInsert(); err != nil {
			return 0, err
		}
	}

	stmt := sqlbuilder.Insert(e).Table(m.Name)
	for _, col := range m.Columns {
		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return 0, fmt.Errorf("未找到该名称 %s 的值", col.GoName)
		}

		// 在为零值的情况下，若该列是 AI 或是有默认值，则过滤掉。无论该零值是否为手动设置的。
		if field.IsZero() && (col.AI || col.HasDefault) {
			continue
		}

		stmt.KeyValue(col.Name, field.Interface())
	}

	return stmt.LastInsertID(m.Name, m.AutoIncrement.Name)
}

func insert(e Engine, v interface{}) (sql.Result, error) {
	m, rval, err := getModel(e, v)
	if err != nil {
		return nil, err
	}

	if m.Type == core.View {
		return nil, fmt.Errorf("模型 %s 的类型是视图，无法从其中删除数据", m.Name)
	}

	if obj, ok := v.(BeforeInserter); ok {
		if err = obj.BeforeInsert(); err != nil {
			return nil, err
		}
	}

	stmt := sqlbuilder.Insert(e).Table(m.Name)
	for _, col := range m.Columns {
		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return nil, fmt.Errorf("未找到该名称 %s 的值", col.GoName)
		}

		// 在为零值的情况下，若该列是 AI 或是有默认值，则过滤掉。无论该零值是否为手动设置的。
		if field.IsZero() && (col.AI || col.HasDefault) {
			continue
		}

		stmt.KeyValue(col.Name, field.Interface())
	}

	return stmt.Exec()
}

// 查找数据。
//
// 根据 v 的 pk 或中唯一索引列查找一行数据，并赋值给 v。
// 若 v 为空，则不发生任何操作，v 可以是数组。
func find(e Engine, v interface{}) error {
	m, rval, err := getModel(e, v)
	if err != nil {
		return err
	}

	stmt := sqlbuilder.Select(e).
		Column("*").
		From(m.Name)
	if err = where(stmt, m, rval); err != nil {
		return err
	}

	_, err = stmt.QueryObject(true, v)
	return err
}

// for update 只能作用于事务
func forUpdate(tx *Tx, v interface{}) error {
	m, rval, err := getModel(tx, v)
	if err != nil {
		return err
	}

	if m.Type == core.View {
		return fmt.Errorf("模型 %s 的类型是视图，无法从其中删除数据", m.Name)
	}

	if obj, ok := v.(BeforeUpdater); ok {
		if err = obj.BeforeUpdate(); err != nil {
			return err
		}
	}

	stmt := tx.SQLBuilder().Select().
		Column("*").
		From(m.Name).
		ForUpdate()
	if err = where(stmt, m, rval); err != nil {
		return err
	}

	_, err = stmt.QueryObject(true, v)
	return err
}

// 更新 v 到数据库，默认情况下不更新零值。
// cols 表示必须要更新的列，即使是零值。
//
// 更新依据为每个对象的主键或是唯一索引列。
// 若不存在此两个类型的字段，则返回错误信息。
func update(e Engine, v interface{}, cols ...string) (sql.Result, error) {
	stmt := sqlbuilder.Update(e)

	m, rval, err := getUpdateColumns(e, v, stmt, cols...)
	if err != nil {
		return nil, err
	}

	if err := where(stmt, m, rval); err != nil {
		return nil, err
	}

	return stmt.Exec()
}

func getUpdateColumns(e Engine, v interface{}, stmt *sqlbuilder.UpdateStmt, cols ...string) (*Model, reflect.Value, error) {
	m, rval, err := getModel(e, v)
	if err != nil {
		return nil, reflect.Value{}, err
	}
	if m.Type == core.View {
		return nil, reflect.Value{}, fmt.Errorf("模型 %s 的类型是视图，无法从其中删除数据", m.Name)
	}

	if obj, ok := v.(BeforeUpdater); ok {
		if err = obj.BeforeUpdate(); err != nil {
			return nil, reflect.Value{}, err
		}
	}

	var occValue interface{}
	for _, col := range m.Columns {
		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return nil, reflect.Value{}, fmt.Errorf("未找到该名称 %s 的值", col.GoName)
		}

		if m.OCC == col { // 乐观锁
			occValue = field.Interface()
		} else if inStrSlice(col.Name, cols) || !field.IsZero() {
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

func inStrSlice(key string, slice []string) bool {
	for _, v := range slice {
		if v == key {
			return true
		}
	}
	return false
}

// 将 v 生成 delete 的 sql 语句
func del(e Engine, v interface{}) (sql.Result, error) {
	m, rval, err := getModel(e, v)
	if err != nil {
		return nil, err
	}

	if m.Type == core.View {
		return nil, fmt.Errorf("模型 %s 的类型是视图，无法从其中删除数据", m.Name)
	}

	stmt := sqlbuilder.Delete(e).Table(m.Name)
	if err = where(stmt, m, rval); err != nil {
		return nil, err
	}

	return stmt.Exec()
}

var errInsertHasDifferentType = errors.New("参数中包含了不同类型的元素")

// rval 为结构体指针组成的数据
func buildInsertManySQL(e *Tx, rval reflect.Value) (*sqlbuilder.InsertStmt, error) {
	query := e.SQLBuilder().Insert()
	var keys []string          // 保存列的顺序，方便后续元素获取值
	var firstType reflect.Type // 记录数组中第一个元素的类型，保证后面的都相同

	for i := 0; i < rval.Len(); i++ {
		irval := rval.Index(i)

		// 判断 beforeInsert
		if obj, ok := irval.Interface().(BeforeInserter); ok {
			if err := obj.BeforeInsert(); err != nil {
				return nil, err
			}
		}

		m, irval, err := getModel(e, irval.Interface())
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
				return nil, errInsertHasDifferentType
			}

			vals := make([]interface{}, 0, len(keys))
			for _, name := range keys {
				col := m.FindColumn(name)
				if col == nil {
					return nil, fmt.Errorf("不存在的列名 %s", name)
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
	} // end for array

	return query, nil
}
