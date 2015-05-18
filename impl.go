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
)

// DB与Tx的共有接口，方便以下方法调用。
type engine interface {
	Dialect() Dialect
	Query(replace bool, query string, args ...interface{}) (*sql.Rows, error)
	Exec(replace bool, query string, args ...interface{}) (sql.Result, error)
	Prepare(replace bool, query string) (*sql.Stmt, error)
}

// 将src转换成sql的值，并写入到w中。
func WriteString(w *bytes.Buffer, src interface{}) error {
	switch v := src.(type) {
	case string:
		w.WriteString(v)
	case []byte:
		w.Write(v)
	case []rune:
		w.WriteString(string(v))
	default:
		rv := reflect.ValueOf(src)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			w.WriteString(strconv.FormatInt(rv.Int(), 10))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			w.WriteString(strconv.FormatUint(rv.Uint(), 10))
		case reflect.Float64:
			w.WriteString(strconv.FormatFloat(rv.Float(), 'g', -1, 64))
		case reflect.Float32:
			w.WriteString(strconv.FormatFloat(rv.Float(), 'g', -1, 32))
		case reflect.Bool:
			w.WriteString(strconv.FormatBool(rv.Bool()))
		default:
			_, err := fmt.Fprint(w, src)
			return err
		}
	}
	return nil
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
// 若两者都不存在，则返回错误信息。
// rval为struct的reflect.Value
func where(sql *bytes.Buffer, m *Model, rval reflect.Value) error {
	if checkCols(m.PK, rval) {
		sql.WriteString(" WHERE ")
		for _, col := range m.PK {
			sql.WriteString(col.Name)
			sql.WriteByte('=')
			WriteString(sql, rval.FieldByName(col.GoName).Interface())
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
			WriteString(sql, field.Interface())
		}
		return nil
	} // end range m.UniqueIndexes

	return errors.New("where:无法产生where部分语句")
}

// 创建或是更新一个数据表。
// v为一个结构体或是结构体指针。
func createOne(e engine, v interface{}) error {
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

	sql, err := e.Dialect().CreateTableSQL(m)
	if err != nil {
		return err
	}

	_, err = e.Exec(false, sql, nil)
	return err
}

// 根据v的pk或中唯一索引列查找一行数据，并赋值给v
func findOne(e engine, v interface{}) error {
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
	e.Dialect().Quote(sql, m.Name)

	if err := where(sql, m, rval); err != nil {
		return err
	}

	rows, err := e.Query(false, sql.String())
	if err != nil {
		return err
	}

	return fetch.Obj(v, rows)
}

// 插入一个对象到数据库
// 以v中的主键或是唯一索引作为where条件语句。
// 自增字段，即使指定了值，也不会被添加
func insertOne(e engine, v interface{}) error {
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
	e.Dialect().Quote(sql, m.Name)

	sql.WriteByte('(')
	for _, col := range keys {
		e.Dialect().Quote(sql, col)
		sql.WriteByte(',')
	}
	sql.Truncate(sql.Len() - 1)
	sql.WriteString(")VALUES(")
	for _, val := range vals {
		WriteString(sql, val)
		sql.WriteByte(',')
	}
	sql.Truncate(sql.Len() - 1)
	sql.WriteByte(')')

	_, err = e.Exec(false, sql.String())
	return err
}

// 更新一个对象
// 以v中的主键或是唯一索引作为where条件语句，其它值为更新值
func updateOne(e engine, v interface{}) error {
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
	e.Dialect().Quote(sql, m.Name)
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

		e.Dialect().Quote(sql, name)
		sql.WriteByte('=')
		WriteString(sql, field.Interface())
	}

	if err := where(sql, m, rval); err != nil {
		return err
	}

	_, err = e.Exec(false, sql.String())
	return err
}

// 删除v表示的单个对象的内容
// 以v中的主键或是唯一索引作为where条件语句
func deleteOne(e engine, v interface{}) error {
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
	e.Dialect().Quote(sql, m.Name)

	if err := where(sql, m, rval); err != nil {
		return err
	}

	_, err = e.Exec(false, sql.String())
	return err
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

	m, err := NewModel(v)
	if err != nil {
		return "", err
	}
	return m.Name, nil
}

func dropOne(e engine, v interface{}) error {
	tbl, err := getTableName(e, v)
	if err != nil {
		return err
	}

	sql := bytes.NewBufferString("DROP ")
	WriteString(sql, tbl)
	_, err = e.Exec(false, sql.String())
	return err
}

func truncateOne(e engine, v interface{}) error {
	tbl, err := getTableName(e, v)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err := e.Dialect().Quote(buf, tbl); err != nil {
		return err
	}

	sql := e.Dialect().TruncateTableSQL(buf.String())
	_, err = e.Exec(false, sql)
	return err
}

// 创建一个或多个数据表
func createMult(e engine, objs ...interface{}) error {
	for _, obj := range objs {
		if err := createOne(e, obj); err != nil {
			return err
		}
	}

	return nil
}

// 插入一个或多个数据
// v可以是对象或是对象数组
func insertMult(e engine, objs ...interface{}) error {
	for _, obj := range objs {
		if err := insertOne(e, obj); err != nil {
			return err
		}
	}
	return nil
}

// 查找多个数据
func findMult(e engine, objs ...interface{}) error {
	for _, obj := range objs {
		if err := findOne(e, obj); err != nil {
			return err
		}
	}
	return nil
}

// 更新一个或多个类型。
// 更新依据为每个对象的主键或是唯一索引列。
// 若不存在此两个类型的字段，则返回错误信息。
func updateMult(e engine, objs ...interface{}) error {
	for _, obj := range objs {
		if err := updateOne(e, obj); err != nil {
			return err
		}
	}
	return nil
}

// 删除指定的数据对象。
func deleteMult(e engine, objs ...interface{}) error {
	for _, obj := range objs {
		if err := deleteOne(e, obj); err != nil {
			return err
		}
	}
	return nil
}

func dropMult(e engine, objs ...interface{}) error {
	for _, obj := range objs {
		if err := dropOne(e, obj); err != nil {
			return err
		}
	}
	return nil
}

func truncateMult(e engine, objs ...interface{}) error {
	for _, obj := range objs {
		if err := truncateOne(e, obj); err != nil {
			return err
		}
	}
	return nil
}
