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

type db interface {
	Dialect() Dialect
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
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

// 供engine.go和tx.go调用的一系列函数。

// 根据model中的主键或是唯一索引为sql产生where语句，
// 若两者都不存在，则返回错误信息。
// rval为struct的reflect.Value
func where(sql *bytes.Buffer, m *Model, rval reflect.Value) error {
	if checkCols(m.PK, rval) {
		sql.WriteString(" WHERE ")
		for _, col := range m.PK {
			sql.WriteString(col.Name)
			sql.WriteByte('=')
			asString(sql, rval.FieldByName(col.GoName).Interface())
		}
		return nil
	}

	// 若不存在pk，也不存在唯一约束
	for _, cols := range m.UniqueIndexes {
		if !checkCols(cols, rval) {
			continue
		}

		sql.WriteString(" WHERE ")
		for _, col := range cols {
			field := rval.FieldByName(col.GoName)
			sql.WriteString(col.Name)
			sql.WriteByte('=')
			asString(sql, field.Interface())
		}
		return nil
	} // end range m.UniqueIndexes

	return errors.New("where:无法产生where部分语句")
}

// 创建或是更新一个数据表。
// v为一个结构体或是结构体指针。
func createOne(db db, v interface{}) error {
	m, err := NewModel(v)
	if err != nil {
		return err
	}

	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return errors.New("createOne:无效的v.Kind()")
	}

	sql, err := db.Dialect().CreateTableSQL(m)
	if err != nil {
		return err
	}

	_, err = db.Exec(sql, nil)
	return err
}

// 根据v的pk或中唯一索引列查找一行数据，并赋值给v
func findOne(db db, v interface{}) error {
	m, err := NewModel(v)
	if err != nil {
		return err
	}

	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return errors.New("findOne:无效的v.Kind()")
	}

	sql := new(bytes.Buffer)
	sql.WriteString("SELECT * FROM ")
	db.Dialect().Quote(sql, m.Name)

	if err := where(sql, m, rval); err != nil {
		return err
	}

	rows, err := db.Query(sql.String())
	if err != nil {
		return err
	}

	return fetch.Obj(v, rows)
}

// 插入一个对象到数据库
// 以v中的主键或是唯一索引作为where条件语句。
// 自增字段，即使指定了值，也不会被添加
func insertOne(db db, v interface{}) error {
	m, err := NewModel(v)
	if err != nil {
		return err
	}

	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return errors.New("insertOne:无效的v.Kind()")
	}

	keys := make([]string, 0, len(m.Cols))
	vals := make([]interface{}, 0, len(m.Cols))
	for name, col := range m.Cols {
		if col.IsAI() { // AI过滤
			continue
		}

		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return fmt.Errorf("insertOne:未找到该名称[%v]的值", col.GoName)
		}
		keys = append(keys, name)
		vals = append(vals, field.Interface())
	}

	if len(keys) == 0 {
		return errors.New("insertOne:未指定任何插入的列数据")
	}

	sql := new(bytes.Buffer)
	sql.WriteString("INSERT INTO ")
	db.Dialect().Quote(sql, m.Name)

	sql.WriteByte('(')
	for _, col := range keys {
		db.Dialect().Quote(sql, col)
		sql.WriteByte(',')
	}
	sql.Truncate(sql.Len() - 1)
	sql.WriteString(")VALUES(")
	for _, val := range vals {
		asString(sql, val)
		sql.WriteByte(',')
	}
	sql.Truncate(sql.Len() - 1)
	sql.WriteByte(')')

	_, err = db.Exec(sql.String())
	return err
}

// 更新一个对象
// 以v中的主键或是唯一索引作为where条件语句，其它值为更新值
func updateOne(db db, v interface{}) error {
	m, err := NewModel(v)
	if err != nil {
		return err
	}

	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return errors.New("updateOne:无效的v.Kind()")
	}

	sql := new(bytes.Buffer)
	sql.WriteString("UPDATE ")
	db.Dialect().Quote(sql, m.Name)
	sql.WriteString(" SET ")

	for name, col := range m.Cols {
		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return fmt.Errorf("updateOne:未找到该名称[%v]的值", col.GoName)
		}

		// 忽略零值，TODO:还需要对比默认值
		if reflect.Zero(col.GoType).Interface() == field.Interface() {
			continue
		}

		db.Dialect().Quote(sql, name)
		sql.WriteByte('=')
		asString(sql, field.Interface())
	}

	if err := where(sql, m, rval); err != nil {
		return err
	}

	_, err = db.Exec(sql.String())
	return err
}

// 删除v表示的单个对象的内容
// 以v中的主键或是唯一索引作为where条件语句
func deleteOne(db db, v interface{}) error {
	m, err := NewModel(v)
	if err != nil {
		return err
	}

	rval := reflect.ValueOf(v)
	for rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return errors.New("deleteOne:无效的v.Kind()")
	}

	sql := new(bytes.Buffer)
	sql.WriteString("DELETE FROM ")
	db.Dialect().Quote(sql, m.Name)

	if err := where(sql, m, rval); err != nil {
		return err
	}

	_, err = db.Exec(sql.String())
	return err
}

// 创建一个或多个数据表
func createMult(db db, objs ...interface{}) error {
	for _, obj := range objs {
		if err := createOne(db, obj); err != nil {
			return err
		}
	}

	return nil
}

// 插入一个或多个数据
// v可以是对象或是对象数组
func insertMult(db db, v interface{}) error {
	rval := reflect.ValueOf(v)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	switch rval.Kind() {
	case reflect.Struct:
		return insertOne(db, v)
	case reflect.Slice, reflect.Array:
		elemType := rval.Type().Elem() // 数组元素的类型

		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}

		if elemType.Kind() != reflect.Struct {
			return errors.New("insertMult:数组元素类型不正确")
		}

		for i := 0; i < rval.Len(); i++ {
			if err := insertOne(db, rval.Index(i).Interface()); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("insertMult:v的类型[%v]无效", rval.Kind())
	}

	return nil
}

// 查找多个数据
func findMult(db db, v interface{}) error {
	rval := reflect.ValueOf(v)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	switch rval.Kind() {
	case reflect.Struct:
		return findOne(db, v)
	case reflect.Array, reflect.Slice:
		elemType := rval.Type().Elem() // 数组元素的类型

		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}

		if elemType.Kind() != reflect.Struct {
			return errors.New("findMult:数组元素类型不正确")
		}

		for i := 0; i < rval.Len(); i++ {
			if err := findOne(db, rval.Index(i).Interface()); err != nil {
				return err
			}
		}
	default:
		return errors.New("findMult:v的类型无效")
	}

	return nil
}

// 更新一个或多个类型。
// 更新依据为每个对象的主键或是唯一索引列。
// 若不存在此两个类型的字段，则返回错误信息。
func updateMult(db db, v interface{}) error {
	rval := reflect.ValueOf(v)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	switch rval.Kind() {
	case reflect.Struct:
		return updateOne(db, v)
	case reflect.Array, reflect.Slice:
		elemType := rval.Type().Elem() // 数组元素的类型

		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}

		if elemType.Kind() != reflect.Struct {
			return errors.New("updateMult:数组元素类型不正确")
		}

		for i := 0; i < rval.Len(); i++ {
			if err := updateOne(db, rval.Index(i).Interface()); err != nil {
				return err
			}
		}
	default:
		return errors.New("updateMult:v的类型无效")
	}

	return nil
}

// 删除指定的数据对象。
func deleteMult(db db, v interface{}) error {
	rval := reflect.ValueOf(v)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	switch rval.Kind() {
	case reflect.Struct:
		return deleteOne(db, v)
	case reflect.Array, reflect.Slice:
		elemType := rval.Type().Elem() // 数组元素的类型

		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}

		if elemType.Kind() != reflect.Struct {
			return errors.New("deleteMult:数组元素类型不正确,只能是指针或是struct的指针")
		}

		for i := 0; i < rval.Len(); i++ {
			if err := deleteOne(db, rval.Index(i).Interface()); err != nil {
				return err
			}
		}
	default:
		return errors.New("deleteMult:v的类型无效")
	}

	return nil
}
