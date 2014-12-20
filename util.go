// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/issue9/orm/core"
)

// 供engine.go和tx.go调用的一系列函数。

// 插入一个对象到数据库
// v.Kind()必须是reflect.Struct
func insertOne(sql *SQL, v interface{}) error {
	rval := reflect.ValueOf(v)

	m, err := core.NewModel(v)
	if err != nil {
		return err
	}

	sql.Reset().Table(m.Name)

	for name, col := range m.Cols {
		sql.Add(name, rval.FieldByName(col.GoType.Name()).Interface())
	}

	_, err = sql.Insert()
	return err
}

// 更新一个对象
func updateOne(sql *SQL, v interface{}) error {
	rval := reflect.ValueOf(v)

	m, err := core.NewModel(v)
	if err != nil {
		return err
	}

	sql.Reset().Table(m.Name)

	switch {
	case len(m.PK) != 0:
		for _, col := range m.PK {
			sql.Where(col.Name+"=?", rval.FieldByName(col.GoType.Name()).Interface())
		}
	case len(m.UniqueIndexes) != 0:
		for _, cols := range m.UniqueIndexes {
			for _, col := range cols {
				sql.Where(col.Name+"=?", rval.FieldByName(col.GoType.Name()).Interface())
			}
			break // 只取一个UniqueIndex就可以了
		}
	default:
		return errors.New("无法产生where部分语句")
	}

	for name, col := range m.Cols {
		sql.Add(name, rval.FieldByName(col.GoType.Name()).Interface())
	}

	_, err = sql.Update()
	return err
}

// 删除单个对象的内容
func deleteOne(sql *SQL, v interface{}) error {
	rval := reflect.ValueOf(v)

	m, err := core.NewModel(v)
	if err != nil {
		return err
	}

	sql.Reset().Table(m.Name)

	switch {
	case len(m.PK) != 0:
		for _, col := range m.PK {
			sql.Where(col.Name+"=?", rval.FieldByName(col.GoType.Name()).Interface())
		}
	case len(m.UniqueIndexes) != 0:
		for _, cols := range m.UniqueIndexes {
			for _, col := range cols {
				sql.Where(col.Name+"=?", rval.FieldByName(col.GoType.Name()).Interface())
			}
			break // 只取一个UniqueIndex就可以了
		}
	default:
		return errors.New("无法产生where部分语句")
	}

	_, err = sql.Delete()
	return err
}

// 插入一个或多个数据
// v可以是对象或是对象数组
func insertMult(sql *SQL, v interface{}) error {
	rval := reflect.ValueOf(v)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	switch rval.Kind() {
	case reflect.Struct:
		return insertOne(sql, v)
	case reflect.Slice, reflect.Array:
		if rval.Elem().Kind() != reflect.Struct {
			return errors.New("数组元素类型不正确")
		}

		for i := 0; i < rval.Len(); i++ {
			if err := insertOne(sql, rval.Index(i).Interface()); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("v的类型[%v]无效", rval.Kind())
	}

	return nil
}

// 更新一个或多个类型。
// 更新依据为每个对象的主键或是唯一索引列。
// 若不存在此两个类型的字段，则返回错误信息。
func updateMult(sql *SQL, v interface{}) error {
	rval := reflect.ValueOf(v)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	switch rval.Kind() {
	case reflect.Struct:
		return updateOne(sql, v)
	case reflect.Array, reflect.Slice:
		if rval.Elem().Kind() != reflect.Struct {
			return errors.New("数组元素类型不正确")
		}

		for i := 0; i < rval.Len(); i++ {
			if err := updateOne(sql, rval.Index(i).Interface()); err != nil {
				return err
			}
		}
	default:
		return errors.New("v的类型无效")
	}

	return nil
}

// 删除指定的数据对象。
func deleteMult(sql *SQL, v interface{}) error {
	rval := reflect.ValueOf(v)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	switch rval.Kind() {
	case reflect.Struct:
		return deleteOne(sql, v)
	case reflect.Array, reflect.Slice:
		if rval.Elem().Kind() != reflect.Struct {
			return errors.New("数组元素类型不正确")
		}

		for i := 0; i < rval.Len(); i++ {
			if err := deleteOne(sql, rval.Index(i).Interface()); err != nil {
				return err
			}
		}
	default:
		return errors.New("v的类型无效")
	}

	return nil
}
