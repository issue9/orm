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

	"github.com/issue9/orm/fetch"
)

// DB与Tx的共有接口，方便以下方法调用。
type engine interface {
	Dialect() Dialect
	Query(replace bool, query string, args ...interface{}) (*sql.Rows, error)
	Exec(replace bool, query string, args ...interface{}) (sql.Result, error)
	Prepare(replace bool, query string) (*sql.Stmt, error)
	Prefix() string
}

// 检测rval中与cols对应的字段都是有效的，且为非零值。
// 若cols的长度为0，返回false。
func checkCols(cols []*Column, rval reflect.Value) bool {
	if len(cols) == 0 {
		return false
	}

	for _, col := range cols {
		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return false
		}

		if reflect.Zero(col.GoType).Interface() == field.Interface() {
			return false
		}
	}
	return true
}

// 根据model中的主键或是唯一索引为sql产生where语句，
// 若两者都不存在，则返回错误信息。rval为struct的reflect.Value
func where(e engine, sql *bytes.Buffer, m *Model, rval reflect.Value) ([]interface{}, error) {
	ret := []interface{}{}

	if checkCols(m.PK, rval) {
		sql.WriteString(" WHERE ")
		for _, col := range m.PK {
			e.Dialect().Quote(sql, col.Name)
			sql.WriteString("=?")
			ret = append(ret, rval.FieldByName(col.GoName).Interface())
		}
		return ret, nil
	}

	// 若不存在pk，也不存在唯一约束
	for _, cols := range m.UniqueIndexes {
		if !checkCols(cols, rval) {
			continue
		}

		sql.WriteString(" WHERE ")
		for _, col := range cols {
			e.Dialect().Quote(sql, col.Name)
			sql.WriteString("=?")
			ret = append(ret, rval.FieldByName(col.GoName).Interface())
			sql.WriteString(" AND ")
		}
		sql.Truncate(sql.Len() - 5) // 去掉最后的" AND "五个字符
		return ret, nil
	} // end range m.UniqueIndexes

	return nil, errors.New("where:无法产生where部分语句")
}

// 获取v对象的表名，v可以是一个结构体，也可以是一个字符串。
func getTableName(e engine, v interface{}) (string, error) {
	switch tbl := v.(type) {
	case string:
		return tbl, nil
	case []rune:
		return string(tbl), nil
	case []byte:
		return string(tbl), nil
	}

	m, err := newModel(v)
	if err != nil {
		return "", err
	}
	return m.Name, nil
}

// 创建一个或多个数据表
// 若objs为空，则不发生任何操作。
func create(e engine, objs ...interface{}) error {
	sql := new(bytes.Buffer)
	d := e.Dialect()

	for i, v := range objs {
		m, err := newModel(v)
		if err != nil {
			return err
		}

		rval := reflect.ValueOf(v)
		for rval.Kind() == reflect.Ptr {
			rval = rval.Elem()
		}

		if rval.Kind() != reflect.Struct {
			return fmt.Errorf("createMult:objs[%v]类型必须为结构体或是结构体指针", i)
		}

		sql.Reset()
		sql.WriteString("CREATE TABLE IF NOT EXISTS ")
		d.Quote(sql, e.Prefix()+m.Name)
		sql.WriteByte('(')
		d.AIColSQL(sql, m)
		d.NoAIColSQL(sql, m)
		d.ConstraintsSQL(sql, m)
		sql.Truncate(sql.Len() - 1)
		sql.WriteByte(')')

		if _, err = e.Exec(false, sql.String()); err != nil {
			return err
		}
	}

	return nil
}

// 插入一个或多个数据
// v可以是对象或是对象数组
// 若objs为空，则不发生任何操作。
func insert(e engine, objs ...interface{}) error {
	sql := new(bytes.Buffer)
	vals := make([]interface{}, 0, 10)

	for i, v := range objs {
		m, err := newModel(v)
		if err != nil {
			return err
		}

		rval := reflect.ValueOf(v)
		for rval.Kind() == reflect.Ptr {
			rval = rval.Elem()
		}

		if rval.Kind() != reflect.Struct {
			return fmt.Errorf("insert:objs[%v]类型必须为结构体或是结构体指针", i)
		}

		vals = vals[:0]
		sql.Reset()
		sql.WriteString("INSERT INTO ")
		e.Dialect().Quote(sql, e.Prefix()+m.Name)
		sql.WriteByte('(')
		for name, col := range m.Cols {
			field := rval.FieldByName(col.GoName)
			if !field.IsValid() {
				return fmt.Errorf("insert:未找到该名称[%v]的值", col.GoName)
			}

			// 在为零值的情况下，若该列是AI或是有默认值，则过滤掉。无论该零值是否为手动设置的。
			if reflect.Zero(col.GoType).Interface() == field.Interface() &&
				(col.IsAI() || col.HasDefault) {
				continue
			}

			e.Dialect().Quote(sql, name)
			sql.WriteByte(',')
			vals = append(vals, field.Interface())
		}

		if len(vals) == 0 {
			return errors.New("insert:未指定任何插入的列数据")
		}

		sql.Truncate(sql.Len() - 1)
		sql.WriteString(")VALUES(")
		for range vals {
			sql.WriteString("?,")
		}
		sql.Truncate(sql.Len() - 1)
		sql.WriteByte(')')

		if _, err = e.Exec(false, sql.String(), vals...); err != nil {
			return err
		}
	}
	return nil
}

// 查找多个数据
// 根据v的pk或中唯一索引列查找一行数据，并赋值给v
// 若objs为空，则不发生任何操作。
func find(e engine, objs ...interface{}) error {
	sql := new(bytes.Buffer)

	for i, v := range objs {
		m, err := newModel(v)
		if err != nil {
			return err
		}

		rval := reflect.ValueOf(v)
		for rval.Kind() == reflect.Ptr {
			rval = rval.Elem()
		}

		if rval.Kind() != reflect.Struct {
			return fmt.Errorf("find:objs[%v]类型必须为结构体或是结构体指针", i)
		}

		sql.Reset()
		sql.WriteString("SELECT * FROM ")
		e.Dialect().Quote(sql, e.Prefix()+m.Name)

		vals, err := where(e, sql, m, rval)
		if err != nil {
			return err
		}

		rows, err := e.Query(false, sql.String(), vals...)
		if err != nil {
			return err
		}
		defer rows.Close()

		if err := fetch.Obj(v, rows); err != nil {
			return err
		}
	}
	return nil
}

// 更新一个或多个类型。
// 更新依据为每个对象的主键或是唯一索引列。
// 若不存在此两个类型的字段，则返回错误信息。
// 若objs为空，则不发生任何操作。
func update(e engine, objs ...interface{}) error {
	sql := new(bytes.Buffer)
	vals := make([]interface{}, 0, 10)

	for i, v := range objs {
		m, err := newModel(v)
		if err != nil {
			return err
		}

		rval := reflect.ValueOf(v)
		for rval.Kind() == reflect.Ptr {
			rval = rval.Elem()
		}

		if rval.Kind() != reflect.Struct {
			return fmt.Errorf("update:objs[%v]类型必须为结构体或是结构体指针", i)
		}

		sql.Reset()
		vals = vals[:0]

		sql.WriteString("UPDATE ")
		e.Dialect().Quote(sql, e.Prefix()+m.Name)
		sql.WriteString(" SET ")

		for name, col := range m.Cols {
			field := rval.FieldByName(col.GoName)
			if !field.IsValid() {
				return fmt.Errorf("update:未找到该名称[%v]的值", col.GoName)
			}

			// 忽略零值，TODO:还需要对比默认值
			if reflect.Zero(col.GoType).Interface() == field.Interface() {
				continue
			}

			e.Dialect().Quote(sql, name)
			sql.WriteString("=?,")
			vals = append(vals, field.Interface())
		}
		sql.Truncate(sql.Len() - 1)

		whereVals, err := where(e, sql, m, rval)
		if err != nil {
			return err
		}
		vals = append(vals, whereVals...)

		if _, err = e.Exec(false, sql.String(), vals...); err != nil {
			return err
		}
	}
	return nil
}

// 删除objs每个元素表示的数据。
// 以objs中每个元素的主键或是唯一索引作为where条件语句。
// 若objs为空，则不发生任何操作。
func del(e engine, objs ...interface{}) error {
	sql := new(bytes.Buffer)
	for i, v := range objs {
		m, err := newModel(v)
		if err != nil {
			return err
		}

		rval := reflect.ValueOf(v)
		for rval.Kind() == reflect.Ptr {
			rval = rval.Elem()
		}

		if rval.Kind() != reflect.Struct {
			return fmt.Errorf("del:objs[%v]类型必须为结构体或是结构体指针", i)
		}

		sql.Reset()
		sql.WriteString("DELETE FROM ")
		e.Dialect().Quote(sql, e.Prefix()+m.Name)

		vals, err := where(e, sql, m, rval)
		if err != nil {
			return err
		}

		if _, err = e.Exec(false, sql.String(), vals...); err != nil {
			return err
		}
	}
	return nil
}

// 删除objs中指定的表名。
// objs可以是字符串表名，或是一个表示model的实例。
// 系统会默认给表名加上表名前缀。
// 若objs为空，则不发生任何操作。
func drop(e engine, objs ...interface{}) error {
	sql := new(bytes.Buffer)
	for _, v := range objs {
		tbl, err := getTableName(e, v)
		if err != nil {
			return err
		}

		sql.Reset()
		sql := bytes.NewBufferString("DROP TABLE IF EXISTS ")
		e.Dialect().Quote(sql, e.Prefix()+tbl)
		if _, err = e.Exec(false, sql.String()); err != nil {
			return err
		}
	}

	return nil
}

// 清空表，并重置AI计数。
// objs可以是字符串表名，或是一个表示model的实例。
// 系统会默认给表名加上表名前缀。
// 若objs为空，则不发生任何操作。
func truncate(e engine, objs ...interface{}) error {
	for _, v := range objs {
		tbl, err := getTableName(e, v)
		if err != nil {
			return err
		}

		sql := e.Dialect().TruncateTableSQL(e.Prefix() + tbl)
		if _, err = e.Exec(false, sql); err != nil {
			return err
		}
	}

	return nil
}

// 插入多条同一model表示的不同数据。
// v 可以是数组，数组指针，或是struct
// NOTE:在go中不能将[]int展开成v...interface{}，
// 所以此处不用...interface{}形式的参数反而会更方便调用者。
func insertMany(e engine, v interface{}) error {
	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	switch rval.Kind() {
	case reflect.Struct: // 单个元素
		return insert(e, v)
	case reflect.Array, reflect.Slice:
	default:
		return errors.New("参数v的类型只能是struct或是数组")
	}

	l := rval.Len()
	sql := bytes.NewBufferString("INSERT INTO ")
	vals := make([]interface{}, 0, 10)
	keys := []string{}
	var firstType reflect.Type

	for i := 0; i < l; i++ {
		irval := rval.Index(i)
		for irval.Kind() == reflect.Ptr {
			irval = irval.Elem()
		}
		if irval.Kind() != reflect.Struct {
			return fmt.Errorf("insert:objs[%v]类型必须为结构体或是结构体指针，当前实际为:[%v]", i, irval.Kind())
		}

		m, err := newModel(irval.Interface())
		if err != nil {
			return err
		}

		if i == 0 { // 第一个元素，需要从只获取列信息。
			vs := new(bytes.Buffer)
			firstType = irval.Type()
			e.Dialect().Quote(sql, e.Prefix()+m.Name) // 指定表名
			sql.WriteByte('(')
			for name, col := range m.Cols {
				field := irval.FieldByName(col.GoName)
				if !field.IsValid() {
					return fmt.Errorf("insert:未找到该名称[%v]的值", col.GoName)
				}

				// 在为零值的情况下，若该列是AI或是有默认值，则过滤掉。无论该零值是否为手动设置的。
				if reflect.Zero(col.GoType).Interface() == field.Interface() &&
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
				return errors.New("参数v中包含了不同类型的元素")
			}

			sql.WriteString(",(")
			for _, name := range keys {
				col := m.Cols[name]
				field := irval.FieldByName(col.GoName)
				if !field.IsValid() {
					return fmt.Errorf("insert:未找到该名称[%v]的值", col.GoName)
				}

				// 在为零值的情况下，若该列是AI或是有默认值，则过滤掉。无论该零值是否为手动设置的。
				if reflect.Zero(col.GoType).Interface() == field.Interface() &&
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

	_, err := e.Exec(false, sql.String(), vals...)
	return err
}
