// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/issue9/orm/fetch"
	"github.com/issue9/orm/forward"
)

// DB与Tx的共有接口，方便以下方法调用。
type engine interface {
	Dialect() forward.Dialect
	Query(replace bool, query string, args ...interface{}) (*sql.Rows, error)
	Exec(replace bool, query string, args ...interface{}) (sql.Result, error)
	Prepare(replace bool, query string) (*sql.Stmt, error)
	Prefix() string
}

// 根据model中的主键或是唯一索引为sql产生where语句，
// 若两者都不存在，则返回错误信息。rval为struct的reflect.Value
func where(e engine, sql *bytes.Buffer, m *forward.Model, rval reflect.Value) ([]interface{}, error) {
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
		return nil, fmt.Errorf("orm.where:无法为[%v]产生where部分语句", m.Name)
	}

	sql.WriteString(" WHERE ")
	for _, key := range keys {
		e.Dialect().Quote(sql, key)
		sql.WriteString("=? AND ")
	}
	sql.Truncate(sql.Len() - 5) // 去掉最后5个字符" AND "

	return vals, nil
}

// 根据rval中任意非零值产生where语句
func whereAny(e engine, sql *bytes.Buffer, m *forward.Model, rval reflect.Value) ([]interface{}, error) {
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
		return nil, fmt.Errorf("orm.whereAny:无法为[%v]产生where部分语句", m.Name)
	}

	sql.WriteString(" WHERE ")
	for _, key := range keys {
		e.Dialect().Quote(sql, key)
		sql.WriteString("=? AND ")
	}
	sql.Truncate(sql.Len() - 5) // 去掉最后5个字符" AND "

	return vals, nil
}

// 创建一个或多个数据表
// 若objs为空，则不发生任何操作。
func buildCreateSQL(sql *bytes.Buffer, e engine, v interface{}) error {
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
		return fmt.Errorf("orm.buildCreateSQL:类型必须为结构体或是结构体指针")
	}

	sql.WriteString("CREATE TABLE IF NOT EXISTS ")
	d.Quote(sql, e.Prefix()+m.Name)
	sql.WriteByte('(')
	d.AIColSQL(sql, m)
	d.NoAIColSQL(sql, m)
	d.ConstraintsSQL(sql, m)
	sql.Truncate(sql.Len() - 1)
	sql.WriteByte(')')

	_, err = e.Exec(false, sql.String())
	return err
}

// 将v生成一个insert的sql语句。
func buildInsertSQL(sql *bytes.Buffer, e engine, v interface{}) ([]interface{}, error) {
	vals := make([]interface{}, 0, 10)
	m, err := forward.NewModel(v)
	if err != nil {
		return nil, err
	}

	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return nil, errors.New("orm.buildInsertSQL:类型必须为结构体或是结构体指针")
	}

	sql.WriteString("INSERT INTO ")
	e.Dialect().Quote(sql, e.Prefix()+m.Name)
	sql.WriteByte('(')
	for name, col := range m.Cols {
		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return nil, fmt.Errorf("orm.buildInsertSQL:未找到该名称[%v]的值", col.GoName)
		}

		// 在为零值的情况下，若该列是AI或是有默认值，则过滤掉。无论该零值是否为手动设置的。
		if col.Zero == field.Interface() &&
			(col.IsAI() || col.HasDefault) {
			continue
		}

		e.Dialect().Quote(sql, name)
		sql.WriteByte(',')
		vals = append(vals, field.Interface())
	}

	if len(vals) == 0 {
		return nil, errors.New("orm.buildInsertSQL:未指定任何插入的列数据")
	}

	sql.Truncate(sql.Len() - 1)
	sql.WriteString(")VALUES(")
	for range vals {
		sql.WriteString("?,")
	}
	sql.Truncate(sql.Len() - 1)
	sql.WriteByte(')')

	return vals, nil
}

// 查找多个数据
// 根据v的pk或中唯一索引列查找一行数据，并赋值给v
// 若objs为空，则不发生任何操作。
// 第一个返回参数用于表示实际有多少数据被导入到objs中。
func buildSelectSQL(sql *bytes.Buffer, e engine, v interface{}) ([]interface{}, error) {
	m, err := forward.NewModel(v)
	if err != nil {
		return nil, err
	}

	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return nil, errors.New("orm.buildSelectSQL:类型必须为结构体或是结构体指针")
	}

	sql.Reset()
	sql.WriteString("SELECT * FROM ")
	e.Dialect().Quote(sql, e.Prefix()+m.Name)

	return where(e, sql, m, rval)
}

// 更新一个或多个类型。
// 更新依据为每个对象的主键或是唯一索引列。
// 若不存在此两个类型的字段，则返回错误信息。
// 若objs为空，则不发生任何操作。
func buildUpdateSQL(sql *bytes.Buffer, e engine, v interface{}) ([]interface{}, error) {

	m, err := forward.NewModel(v)
	if err != nil {
		return nil, err
	}

	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return nil, errors.New("orm.buildUpdateSQL:类型必须为结构体或是结构体指针")
	}

	vals := make([]interface{}, 0, 10)
	sql.WriteString("UPDATE ")
	e.Dialect().Quote(sql, e.Prefix()+m.Name)
	sql.WriteString(" SET ")

	for name, col := range m.Cols {
		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return nil, fmt.Errorf("orm.buildUpdateSQL:未找到该名称[%v]的值", col.GoName)
		}

		// 忽略零值，TODO:还需要对比默认值
		if col.Zero == field.Interface() {
			continue
		}

		e.Dialect().Quote(sql, name)
		sql.WriteString("=?,")
		vals = append(vals, field.Interface())
	}
	sql.Truncate(sql.Len() - 1)

	whereVals, err := where(e, sql, m, rval)
	if err != nil {
		return nil, err
	}
	return append(vals, whereVals...), nil
}

// 将v生成delete的sql语句
func buildDeleteSQL(sql *bytes.Buffer, e engine, v interface{}) ([]interface{}, error) {
	m, err := forward.NewModel(v)
	if err != nil {
		return nil, err
	}

	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return nil, fmt.Errorf("orm.buildDeleteSQL:类型必须为结构体或是结构体指针")
	}

	sql.WriteString("DELETE FROM ")
	e.Dialect().Quote(sql, e.Prefix()+m.Name)

	return where(e, sql, m, rval)

}

// 删除objs中指定的表名。
// 系统会默认给表名加上表名前缀。
// 若objs为空，则不发生任何操作。
func buildDropSQL(sql *bytes.Buffer, e engine, v interface{}) error {
	m, err := forward.NewModel(v)
	if err != nil {
		return err
	}

	sql.WriteString("DROP TABLE IF EXISTS ")
	e.Dialect().Quote(sql, e.Prefix()+m.Name)
	_, err = e.Exec(false, sql.String())
	return err
}

// 清空表，并重置AI计数。
// 系统会默认给表名加上表名前缀。
func buildTruncateSQL(sql *bytes.Buffer, e engine, v interface{}) error {
	m, err := forward.NewModel(v)
	if err != nil {
		return err
	}

	aiName := ""
	if m.AI != nil {
		aiName = m.AI.Name
	}
	e.Dialect().TruncateTableSQL(sql, e.Prefix()+m.Name, aiName)

	if _, err = e.Exec(false, sql.String()); err != nil {
		return err
	}
	return nil
}

// 统计符合v条件的记录数量。
func count(e engine, v interface{}) (int, error) {
	m, err := forward.NewModel(v)
	if err != nil {
		return 0, err
	}

	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return 0, errors.New("orm.count:参数v类型必须为结构体或是结构体指针")
	}

	sql := bytes.NewBufferString("SELECT COUNT(*) AS count FROM ")
	e.Dialect().Quote(sql, e.Prefix()+m.Name)
	vals, err := whereAny(e, sql, m, rval)
	if err != nil {
		return 0, err
	}

	rows, err := e.Query(false, sql.String(), vals...)
	if err != nil {
		rows.Close() // 错误时关闭rows
		return 0, err
	}
	data, err := fetch.ColumnString(true, "count", rows)
	rows.Close() // 及时关闭rows
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(data[0])
}

func buildInsertManySQL(sql *bytes.Buffer, e engine, rval reflect.Value) ([]interface{}, error) {
	sql.WriteString("INSERT INTO ")
	vals := make([]interface{}, 0, 10)
	keys := []string{}
	var firstType reflect.Type

	for i := 0; i < rval.Len(); i++ {
		irval := rval.Index(i)
		for irval.Kind() == reflect.Ptr {
			irval = irval.Elem()
		}

		if irval.Kind() != reflect.Struct {
			return nil, fmt.Errorf("orm.buildInsertManySQL:第[%v]个元素的类型必须为结构体或是结构体指针，当前实际为:[%v]", i, irval.Kind())
		}

		m, err := forward.NewModel(irval.Interface())
		if err != nil {
			return nil, err
		}

		if i == 0 { // 第一个元素，需要从中获取列信息。
			vs := new(bytes.Buffer)

			firstType = irval.Type()
			e.Dialect().Quote(sql, e.Prefix()+m.Name) // 指定表名
			sql.WriteByte('(')
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

				e.Dialect().Quote(sql, name)
				sql.WriteByte(',')

				vs.WriteString("?,")
				vals = append(vals, field.Interface())
				keys = append(keys, name) // 记录列的顺序
			}
			sql.Truncate(sql.Len() - 1)
			vs.Truncate(vs.Len() - 1)
			sql.WriteString(")VALUES(")
			sql.WriteString(vs.String())
			sql.WriteByte(')')
		} else { // 之后的元素，只需要获取其对应的值就行
			if firstType != irval.Type() { // 与第一个元素的类型不同。
				return nil, errors.New("orm.buildInsertManySQL:参数v中包含了不同类型的元素")
			}

			sql.WriteString(",(")
			for _, name := range keys {
				col := m.Cols[name]
				field := irval.FieldByName(col.GoName)
				if !field.IsValid() {
					return nil, fmt.Errorf("orm.buildInsertManySQL:未找到该名称[%v]的值", col.GoName)
				}

				// 在为零值的情况下，若该列是AI或是有默认值，则过滤掉。无论该零值是否为手动设置的。
				if col.Zero == field.Interface() &&
					(col.IsAI() || col.HasDefault) {
					continue
				}

				sql.WriteString("?,")
				vals = append(vals, field.Interface())
			}
			sql.Truncate(sql.Len() - 1)
			sql.WriteByte(')')
		}
	} // end for array

	return vals, nil
}
