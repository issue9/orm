// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/issue9/orm/fetch"
	"github.com/issue9/orm/forward"
)

var ErrInvalidKind = errors.New("不支持的reflect.Kind()，只能是结构体或是结构体指针")

// 根据model中的主键或是唯一索引为sql产生where语句，
// 若两者都不存在，则返回错误信息。rval为struct的reflect.Value
func where(e forward.Engine, sql *forward.SQL, m *forward.Model, rval reflect.Value) error {
	vals := make([]interface{}, 0, 3)
	keys := make([]string, 0, 3)

	// 获取构成where的键名和键值
	getKV := func(cols []*forward.Column) bool {
		for _, col := range cols {
			field := rval.FieldByName(col.GoName)

			if !field.IsValid() ||
				col.Zero == field.Interface() {
				vals = vals[:0]
				keys = keys[:0]
				return false
			}

			keys = append(keys, col.Name)
			vals = append(vals, field.Interface())
		}
		return true
	}

	if !getKV(m.PK) { // 没有主键，则尝试唯一约束
		for _, cols := range m.UniqueIndexes {
			if getKV(cols) {
				break
			}
		}
	}

	if len(keys) == 0 {
		return fmt.Errorf("orm.where:无法为[%v]产生where部分语句", m.Name)
	}

	for index, key := range keys {
		sql.And("{"+key+"}=?", vals[index])
	}

	return nil
}

// 根据rval中任意非零值产生where语句
func whereAny(e forward.Engine, sql *forward.SQL, m *forward.Model, rval reflect.Value) error {
	vals := make([]interface{}, 0, 3)
	keys := make([]string, 0, 3)

	for _, col := range m.Cols {
		field := rval.FieldByName(col.GoName)

		if !field.IsValid() || col.Zero == field.Interface() {
			continue
		}

		keys = append(keys, col.Name)
		vals = append(vals, field.Interface())
	}

	if len(keys) == 0 {
		return fmt.Errorf("orm.whereAny:无法为[%v]产生where部分语句", m.Name)
	}

	for index, key := range keys {
		sql.And("{"+key+"}=?", vals[index])
	}

	return nil
}

// 创建一个或多个数据表
// 若objs为空，则不发生任何操作。
func buildCreateSQL(sql *forward.SQL, e forward.Engine, v interface{}) error {
	d := e.Dialect()
	m, err := forward.NewModel(v)
	if err != nil {
		return err
	}

	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return ErrInvalidKind
	}

	sql.WriteString("CREATE TABLE IF NOT EXISTS ").
		WriteString("{#").
		WriteString(m.Name).
		WriteString("}(")
	d.AIColSQL(sql, m)
	d.NoAIColSQL(sql, m)
	d.ConstraintsSQL(sql, m)
	sql.TruncateLast(1).WriteByte(')')

	return nil
}

// 统计符合 v 条件的记录数量。
func count(e forward.Engine, v interface{}) (int, error) {
	m, err := forward.NewModel(v)
	if err != nil {
		return 0, err
	}

	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return 0, ErrInvalidKind
	}

	sql := forward.NewSQL(e).Select("COUNT(*)AS count").From("{#" + m.Name + "}")
	err = whereAny(e, sql, m, rval)
	if err != nil {
		return 0, err
	}

	rows, err := sql.Query(true)
	if err != nil {
		return 0, err
	}
	data, err := fetch.ColumnString(true, "count", rows)
	rows.Close() // 及时关闭rows
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(data[0])
}

// 创建表。
func create(e forward.Engine, v interface{}) error {
	sql := forward.NewSQL(e)
	if err := buildCreateSQL(sql, e, v); err != nil {
		return err
	}
	if _, err := sql.Exec(true); err != nil {
		return err
	}

	// CREATE INDEX
	m, err := forward.NewModel(v)
	if err != nil {
		return err
	}
	if len(m.KeyIndexes) == 0 {
		return nil
	}
	for name, cols := range m.KeyIndexes {
		sql.Reset().
			WriteString("CREATE INDEX ").
			WriteByte('{').WriteString(name).WriteByte('}').
			WriteString(" ON ").
			WriteString("{#").WriteString(m.Name).WriteString("}(")
		for _, col := range cols {
			sql.WriteByte('{').WriteString(col.Name).WriteString("},")
		}
		sql.TruncateLast(1)
		sql.WriteByte(')')
		if _, err := sql.Exec(true); err != nil {
			return err
		}
	}
	return nil
}

// 删除一张表。
func drop(e forward.Engine, v interface{}) error {
	m, err := forward.NewModel(v)
	if err != nil {
		return err
	}

	_, err = forward.NewSQL(e).
		WriteString("DROP TABLE IF EXISTS ").
		WriteString("{#").
		WriteString(m.Name).
		WriteByte('}').
		Exec(true)
	return err
}

// 清空表，并重置AI计数。
// 系统会默认给表名加上表名前缀。
func truncate(e forward.Engine, v interface{}) error {
	m, err := forward.NewModel(v)
	if err != nil {
		return err
	}

	aiName := ""
	if m.AI != nil {
		aiName = m.AI.Name
	}

	sql := forward.NewSQL(e)
	e.Dialect().TruncateTableSQL(sql, "#"+m.Name, aiName)

	_, err = sql.Exec(true)
	return err
}

func insert(e forward.Engine, v interface{}) (sql.Result, error) {
	m, err := forward.NewModel(v)
	if err != nil {
		return nil, err
	}

	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return nil, ErrInvalidKind
	}

	keys := make([]string, 0, len(m.Cols))
	vals := make([]interface{}, 0, len(m.Cols))
	for name, col := range m.Cols {
		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return nil, fmt.Errorf("orm.insert:未找到该名称[%v]的值", col.GoName)
		}

		// 在为零值的情况下，若该列是AI或是有默认值，则过滤掉。无论该零值是否为手动设置的。
		if col.Zero == field.Interface() &&
			(col.IsAI() || col.HasDefault) {
			continue
		}

		keys = append(keys, "{"+name+"}")
		vals = append(vals, field.Interface())
	}

	if len(vals) == 0 {
		return nil, errors.New("orm.insert:未指定任何插入的列数据")
	}

	sql := forward.NewSQL(e).
		Insert("{#" + m.Name + "}").
		Keys(keys...).
		Values(vals...)

	if sql.HasError() {
		return nil, sql.Errors()
	}
	return sql.Exec(true)
}

// 查找多个数据。
//
// 根据 v 的 pk 或中唯一索引列查找一行数据，并赋值给 v。
// 若 v 为空，则不发生任何操作，v 可以是数组。
func find(e forward.Engine, v interface{}) error {
	m, err := forward.NewModel(v)
	if err != nil {
		return err
	}

	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return ErrInvalidKind
	}

	sql := forward.NewSQL(e).Select("*").From("{#" + m.Name + "}")
	if err = where(e, sql, m, rval); err != nil {
		return err
	}

	_, err = sql.QueryObj(true, v)
	return err
}

// 更新 v 到数据库，默认情况下不更新零值。
// cols 表示必须要更新的列，即使是零值。
//
// 更新依据为每个对象的主键或是唯一索引列。
// 若不存在此两个类型的字段，则返回错误信息。
func update(e forward.Engine, v interface{}, cols ...string) (sql.Result, error) {
	m, err := forward.NewModel(v)
	if err != nil {
		return nil, err
	}

	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return nil, ErrInvalidKind
	}

	sql := forward.NewSQL(e).Update("{#" + m.Name + "}")
	for name, col := range m.Cols {
		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return nil, fmt.Errorf("orm.update:未找到该名称[%v]的值", col.GoName)
		}

		// 零值，但是不属于指定需要更新的列
		if !inStrSlice(name, cols) && col.Zero == field.Interface() {
			continue
		}

		sql.Set("{"+name+"}", field.Interface())
	}

	if err := where(e, sql, m, rval); err != nil {
		return nil, err
	}

	return sql.Exec(true)
}

func inStrSlice(key string, slice []string) bool {
	for _, v := range slice {
		if v == key {
			return true
		}
	}
	return false
}

// 将v生成delete的sql语句
func del(e forward.Engine, v interface{}) (sql.Result, error) {
	m, err := forward.NewModel(v)
	if err != nil {
		return nil, err
	}

	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return nil, ErrInvalidKind
	}

	sql := forward.NewSQL(e).Delete("{#" + m.Name + "}")
	if err = where(e, sql, m, rval); err != nil {
		return nil, err
	}

	return sql.Exec(true)
}

// rval 为结构体指针组成的数据
func buildInsertManySQL(e forward.Engine, rval reflect.Value) (*forward.SQL, error) {
	sql := forward.NewSQL(e)
	vals := make([]interface{}, 0, 10)
	keys := []string{}         // 保存列的顺序，方便后续元素获取值
	var firstType reflect.Type // 记录数组中第一个元素的类型，保证后面的都相同

	for i := 0; i < rval.Len(); i++ {
		irval := rval.Index(i)
		for irval.Kind() == reflect.Ptr {
			irval = irval.Elem()
		}

		if irval.Kind() != reflect.Struct {
			return nil, ErrInvalidKind
		}

		m, err := forward.NewModel(irval.Interface())
		if err != nil {
			return nil, err
		}

		if i == 0 { // 第一个元素，需要从中获取列信息。
			firstType = irval.Type()
			sql.Insert("{#" + m.Name + "}")
			cols := []string{}

			for name, col := range m.Cols {
				field := irval.FieldByName(col.GoName)
				if !field.IsValid() {
					return nil, fmt.Errorf("orm.buildInsertManySQL:未找到该名称[%v]的值", col.GoName)
				}

				// 在为零值的情况下，若该列是AI或是有默认值，则过滤掉。无论该零值是否为手动设置的。
				if col.Zero == field.Interface() &&
					(col.IsAI() || col.HasDefault) {
					continue
				}

				vals = append(vals, field.Interface())
				cols = append(cols, "{"+name+"}")
				keys = append(keys, name)
			}
			sql.Keys(cols...).Values(vals...)
		} else { // 之后的元素，只需要获取其对应的值就行
			if firstType != irval.Type() { // 与第一个元素的类型不同。
				return nil, errors.New("orm.buildInsertManySQL:参数v中包含了不同类型的元素")
			}

			vals = vals[:0]
			for _, name := range keys {
				col, found := m.Cols[name]
				if !found {
					return nil, fmt.Errorf("orm:buildInsertManySQL:不存在的列名:[%v]", name)
				}

				field := irval.FieldByName(col.GoName)
				if !field.IsValid() {
					return nil, fmt.Errorf("orm.buildInsertManySQL:未找到该名称[%v]的值", col.GoName)
				}

				// 在为零值的情况下，若该列是AI或是有默认值，则过滤掉。无论该零值是否为手动设置的。
				if col.Zero == field.Interface() &&
					(col.IsAI() || col.HasDefault) {
					continue
				}

				vals = append(vals, field.Interface())
			}
			sql.Values(vals...)
		}
	} // end for array

	return sql, nil
}
