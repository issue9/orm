// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/issue9/orm/builder"
	"github.com/issue9/orm/core"
)

// 检测rval中与cols对应的字段都是有效的，且为非零值。
// 若cols的长度为0，返回false。
func checkCols(cols []*core.Column, rval reflect.Value) bool {
	if len(cols) == 0 {
		return false
	}

	for _, col := range cols {
		field := rval.FieldByName(col.GoName)
		if reflect.Zero(col.GoType).Interface() == field.Interface() {
			return false
		}

		if !field.IsValid() {
			return false
		}
	}
	return true
}

// 供engine.go和tx.go调用的一系列函数。

// 根据model中的主键或是唯一索引产生where语句，
// 若两者都不存在，则返回错误信息。
// rval为struct的reflect.Value
func where(sql *builder.SQL, m *core.Model, rval reflect.Value) error {
	if checkCols(m.PK, rval) {
		for _, col := range m.PK {
			sql.Where(col.Name + "=" + core.AsSQLValue(rval.FieldByName(col.GoName).Interface()))
		}
		return nil
	}

	// 若不存在pk，也不存在唯一约束
	for _, cols := range m.UniqueIndexes {
		if !checkCols(cols, rval) {
			continue
		}

		for _, col := range cols {
			field := rval.FieldByName(col.GoName)
			sql.Where(col.Name + "=" + core.AsSQLValue(field.Interface()))
		}
		return nil
	} // end range m.UniqueIndexes

	return errors.New("where:无法产生where部分语句")
}

// 创建或是更新一个数据表。
// v为一个结构体或是结构体指针。
func createOne(db core.DB, v interface{}) error {
	rval := reflect.ValueOf(v)

	m, err := core.NewModel(v)
	if err != nil {
		return err
	}

	if rval.Kind() == reflect.Ptr {
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

// 插入一个对象到数据库
// 以v中的主键或是唯一索引作为where条件语句。
// 自增字段，即使指定了值，也不会被添加
func insertOne(sql *builder.SQL, v interface{}) error {
	rval := reflect.ValueOf(v)

	m, err := core.NewModel(v)
	if err != nil {
		return err
	}

	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return errors.New("insertOne:无效的v.Kind()")
	}

	sql.Reset().Table(m.Name)

	for name, col := range m.Cols {
		if col.IsAI() { // AI过滤
			continue
		}

		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return fmt.Errorf("insertOne:未找到该名称[%v]的值", col.GoName)
		}
		sql.Set(name, field.Interface())
	}

	_, err = sql.Insert(nil)
	return err
}

// 更新一个对象
// 以v中的主键或是唯一索引作为where条件语句，其它值为更新值
func updateOne(sql *builder.SQL, v interface{}) error {
	rval := reflect.ValueOf(v)

	m, err := core.NewModel(v)
	if err != nil {
		return err
	}

	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return errors.New("updateOne:无效的v.Kind()")
	}

	sql.Reset().Table(m.Name)

	if err := where(sql, m, rval); err != nil {
		return err
	}

	for name, col := range m.Cols {
		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return fmt.Errorf("updateOne:未找到该名称[%v]的值", col.GoName)
		}

		// 忽略零值
		if reflect.Zero(col.GoType).Interface() == field.Interface() {
			continue
		}

		sql.Set(name, field.Interface())
	}

	_, err = sql.Update(nil)
	return err
}

// 删除v表示的单个对象的内容
// 以v中的主键或是唯一索引作为where条件语句
func deleteOne(sql *builder.SQL, v interface{}) error {
	rval := reflect.ValueOf(v)

	m, err := core.NewModel(v)
	if err != nil {
		return err
	}

	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	if rval.Kind() != reflect.Struct {
		return errors.New("deleteOne:无效的v.Kind()")
	}

	sql.Reset().Table(m.Name)

	if err := where(sql, m, rval); err != nil {
		return err
	}

	_, err = sql.Delete(nil)
	return err
}

// 创建一个或多个数据表
func createMult(db core.DB, objs ...interface{}) error {
	for _, obj := range objs {
		if err := createOne(db, obj); err != nil {
			return err
		}
	}

	return nil
}

// 插入一个或多个数据
// v可以是对象或是对象数组
func insertMult(sql *builder.SQL, v interface{}) error {
	rval := reflect.ValueOf(v)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	switch rval.Kind() {
	case reflect.Struct:
		return insertOne(sql, v)
	case reflect.Slice, reflect.Array:
		elemType := rval.Type().Elem() // 数组元素的类型

		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}

		if elemType.Kind() != reflect.Struct {
			return errors.New("insertMult:数组元素类型不正确")
		}

		for i := 0; i < rval.Len(); i++ {
			if err := insertOne(sql, rval.Index(i).Interface()); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("insertMult:v的类型[%v]无效", rval.Kind())
	}

	return nil
}

// 更新一个或多个类型。
// 更新依据为每个对象的主键或是唯一索引列。
// 若不存在此两个类型的字段，则返回错误信息。
func updateMult(sql *builder.SQL, v interface{}) error {
	rval := reflect.ValueOf(v)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	switch rval.Kind() {
	case reflect.Struct:
		return updateOne(sql, v)
	case reflect.Array, reflect.Slice:
		elemType := rval.Type().Elem() // 数组元素的类型

		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}

		if elemType.Kind() != reflect.Struct {
			return errors.New("updateMult:数组元素类型不正确")
		}

		for i := 0; i < rval.Len(); i++ {
			if err := updateOne(sql, rval.Index(i).Interface()); err != nil {
				return err
			}
		}
	default:
		return errors.New("updateMult:v的类型无效")
	}

	return nil
}

// 删除指定的数据对象。
func deleteMult(sql *builder.SQL, v interface{}) error {
	rval := reflect.ValueOf(v)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	switch rval.Kind() {
	case reflect.Struct:
		return deleteOne(sql, v)
	case reflect.Array, reflect.Slice:
		elemType := rval.Type().Elem() // 数组元素的类型

		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}

		if elemType.Kind() != reflect.Struct {
			return errors.New("deleteMult:数组元素类型不正确,只能是指针或是struct的指针")
		}

		for i := 0; i < rval.Len(); i++ {
			if err := deleteOne(sql, rval.Index(i).Interface()); err != nil {
				return err
			}
		}
	default:
		return errors.New("deleteMult:v的类型无效")
	}

	return nil
}
