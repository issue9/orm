// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/issue9/orm/model"
	"github.com/issue9/orm/sqlbuilder"
)

func getModel(v interface{}) (*model.Model, reflect.Value, error) {
	m, err := model.New(v)
	if err != nil {
		return nil, reflect.Value{}, err
	}

	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	return m, rval, nil
}

// 根据 model 中的主键或是唯一索引为 sql 产生 where 语句，
// 若两者都不存在，则返回错误信息。rval 为 struct 的 reflect.Value
func where(sql sqlbuilder.WhereStmter, m *model.Model, rval reflect.Value) error {
	vals := make([]interface{}, 0, 3)
	keys := make([]string, 0, 3)

	// 获取构成 where 的键名和键值
	getKV := func(cols []*model.Column) bool {
		for _, col := range cols {
			field := rval.FieldByName(col.GoName)

			if !field.IsValid() || col.Zero == field.Interface() {
				vals = vals[:0]
				keys = keys[:0]
				return false
			}

			keys = append(keys, col.Name)
			vals = append(vals, field.Interface())
		}
		return len(keys) > 0 // 如果 keys 中有数据，表示已经采集成功，否则表示 cols 的长度为 0
	}

	if !getKV(m.PK) { // 没有主键，则尝试唯一约束
		for _, cols := range m.UniqueIndexes {
			if getKV(cols) {
				break
			}
		}
	}

	if len(keys) == 0 {
		return fmt.Errorf("没有主键或唯一约束，无法为 %s 产生 where 部分语句", m.Name)
	}

	for index, key := range keys {
		sql.WhereStmt().And("{"+key+"}=?", vals[index])
	}

	return nil
}

// 根据 rval 中任意非零值产生 where 语句
func whereAny(sql sqlbuilder.WhereStmter, m *model.Model, rval reflect.Value) error {
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
		return fmt.Errorf("没有非零值字段，无法为 %s 产生 where 部分语句", m.Name)
	}

	for index, key := range keys {
		sql.WhereStmt().And("{"+key+"}=?", vals[index])
	}

	return nil
}

// 统计符合 v 条件的记录数量。
func count(e Engine, v interface{}) (int64, error) {
	m, rval, err := getModel(v)
	if err != nil {
		return 0, err
	}

	sql := e.SQL().Select().Count("COUNT(*) AS count").From("{#" + m.Name + "}")
	if err = whereAny(sql, m, rval); err != nil {
		return 0, err
	}

	return sql.QueryInt("count")
}

// 创建表。
//
// 部分数据库可能并没有提供在 CREATE TABLE 中直接指定 index 约束的功能。
// 所以此处把创建表和创建索引分成两步操作。
func create(e Engine, v interface{}) error {
	m, _, err := getModel(v)
	if err != nil {
		return err
	}

	sqls, err := e.Dialect().CreateTableSQL(m)
	if err != nil {
		return err
	}

	for _, sql := range sqls {
		if _, err := e.Exec(sql); err != nil {
			return err
		}
	}

	return nil
}

// 删除一张表。
func drop(e Engine, v interface{}) error {
	m, err := model.New(v)
	if err != nil {
		return err
	}

	_, err = e.SQL().DropTable().Table("{#" + m.Name + "}").Exec()
	return err
}

// 清空表，并重置 AI 计数。
func truncate(e Engine, v interface{}) error {
	m, err := model.New(v)
	if err != nil {
		return err
	}

	sql := e.SQL().Truncate().Table("#" + m.Name)
	if m.AI != nil {
		sql.AI("{" + m.AI.Name + "}")
	}

	_, err = sql.Exec()
	return err
}

func insert(e Engine, v interface{}) (sql.Result, error) {
	m, rval, err := getModel(v)
	if err != nil {
		return nil, err
	}

	sql := e.SQL().Insert().Table("{#" + m.Name + "}")
	for name, col := range m.Cols {
		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return nil, fmt.Errorf("未找到该名称 %s 的值", col.GoName)
		}

		// 在为零值的情况下，若该列是 AI 或是有默认值，则过滤掉。无论该零值是否为手动设置的。
		if col.Zero == field.Interface() &&
			(col.IsAI() || col.HasDefault) {
			continue
		}

		sql.KeyValue("{"+name+"}", field.Interface())
	}

	return sql.Exec()
}

// 查找数据。
//
// 根据 v 的 pk 或中唯一索引列查找一行数据，并赋值给 v。
// 若 v 为空，则不发生任何操作，v 可以是数组。
func find(e Engine, v interface{}) error {
	m, rval, err := getModel(v)
	if err != nil {
		return err
	}

	sql := e.SQL().Select().
		Select("*").
		From("{#" + m.Name + "}")
	if err = where(sql, m, rval); err != nil {
		return err
	}

	_, err = sql.QueryObj(v)
	return err
}

// for update 只能作用于事务
func forUpdate(tx *Tx, v interface{}) error {
	m, rval, err := getModel(v)
	if err != nil {
		return err
	}

	sql := tx.SQL().Select().
		Select("*").
		From("{#" + m.Name + "}").
		ForUpdate()
	if err = where(sql, m, rval); err != nil {
		return err
	}

	_, err = sql.QueryObj(v)
	return err
}

// 更新 v 到数据库，默认情况下不更新零值。
// cols 表示必须要更新的列，即使是零值。
//
// 更新依据为每个对象的主键或是唯一索引列。
// 若不存在此两个类型的字段，则返回错误信息。
func update(e Engine, v interface{}, cols ...string) (sql.Result, error) {
	m, rval, err := getModel(v)
	if err != nil {
		return nil, err
	}

	sql := e.SQL().Update().Table("{#" + m.Name + "}")
	var occValue interface{}
	for name, col := range m.Cols {
		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return nil, fmt.Errorf("未找到该名称 %s 的值", col.GoName)
		}

		// 零值，但是不属于指定需要更新的列
		if !inStrSlice(name, cols) && col.Zero == field.Interface() {
			continue
		}

		if m.OCC == col { // 乐观锁
			occValue = field.Interface()
			continue
		} else {
			sql.Set("{"+name+"}", field.Interface())
		}
	}

	if m.OCC != nil {
		sql.OCC("{"+m.OCC.Name+"}", occValue)
	}

	if err := where(sql, m, rval); err != nil {
		return nil, err
	}

	return sql.Exec()
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
	m, rval, err := getModel(v)
	if err != nil {
		return nil, err
	}

	sql := e.SQL().Delete().Table("{#" + m.Name + "}")
	if err = where(sql, m, rval); err != nil {
		return nil, err
	}

	return sql.Exec()
}

// rval 为结构体指针组成的数据
func buildInsertManySQL(e *Tx, rval reflect.Value) (*sqlbuilder.InsertStmt, error) {
	sql := e.SQL().Insert()
	keys := []string{}         // 保存列的顺序，方便后续元素获取值
	var firstType reflect.Type // 记录数组中第一个元素的类型，保证后面的都相同

	for i := 0; i < rval.Len(); i++ {
		irval := rval.Index(i)

		m, irval, err := getModel(irval.Interface())
		if err != nil {
			return nil, err
		}

		if i == 0 { // 第一个元素，需要从中获取列信息。
			firstType = irval.Type()
			sql.Table("{#" + m.Name + "}")

			for name, col := range m.Cols {
				field := irval.FieldByName(col.GoName)
				if !field.IsValid() {
					return nil, fmt.Errorf("未找到该名称 %s 的值", col.GoName)
				}

				// 在为零值的情况下，若该列是 AI 或是有默认值，则过滤掉。无论该零值是否为手动设置的。
				if col.Zero == field.Interface() &&
					(col.IsAI() || col.HasDefault) {
					continue
				}

				sql.KeyValue("{"+name+"}", field.Interface())
				keys = append(keys, name)
			}
		} else { // 之后的元素，只需要获取其对应的值就行
			if firstType != irval.Type() { // 与第一个元素的类型不同。
				return nil, errors.New("参数 v 中包含了不同类型的元素")
			}

			vals := make([]interface{}, 0, len(keys))
			for _, name := range keys {
				col, found := m.Cols[name]
				if !found {
					return nil, fmt.Errorf("不存在的列名 %s", name)
				}

				field := irval.FieldByName(col.GoName)
				if !field.IsValid() {
					return nil, fmt.Errorf("未找到该名称 %s 的值", col.GoName)
				}

				// 在为零值的情况下，若该列是 AI 或是有默认值，则过滤掉。无论该零值是否为手动设置的。
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
