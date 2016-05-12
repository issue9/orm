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
	"github.com/issue9/orm/sqlbuilder"
)

var ErrInvalidKind = errors.New("不支持的reflect.Kind()，只能是结构体或是结构体指针")

// 根据model中的主键或是唯一索引为sql产生where语句，
// 若两者都不存在，则返回错误信息。rval为struct的reflect.Value
func where(e forward.Engine, sql *bytes.Buffer, m *forward.Model, rval reflect.Value) ([]interface{}, error) {
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
func whereAny(e forward.Engine, sql *bytes.Buffer, m *forward.Model, rval reflect.Value) ([]interface{}, error) {
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
func buildCreateSQL(sql *bytes.Buffer, e forward.Engine, v interface{}) error {
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

// 查找多个数据
// 根据v的pk或中唯一索引列查找一行数据，并赋值给v
// 若objs为空，则不发生任何操作。
// 第一个返回参数用于表示实际有多少数据被导入到objs中。
func buildSelectSQL(sql *bytes.Buffer, e forward.Engine, v interface{}) ([]interface{}, error) {
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

	sql.Reset()
	sql.WriteString("SELECT * FROM ")
	e.Dialect().Quote(sql, e.Prefix()+m.Name)

	return where(e, sql, m, rval)
}

// 更新一个或多个类型。
// 更新依据为每个对象的主键或是唯一索引列。
// 若不存在此两个类型的字段，则返回错误信息。
// 若objs为空，则不发生任何操作。
// zero 是否提交值为零的内容。
func buildUpdateSQL(sql *bytes.Buffer, e forward.Engine, v interface{}, zero bool) ([]interface{}, error) {
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

	vals := make([]interface{}, 0, 10)
	sql.WriteString("UPDATE ")
	e.Dialect().Quote(sql, e.Prefix()+m.Name)
	sql.WriteString(" SET ")

	for name, col := range m.Cols {
		field := rval.FieldByName(col.GoName)
		if !field.IsValid() {
			return nil, fmt.Errorf("orm.buildUpdateSQL:未找到该名称[%v]的值", col.GoName)
		}

		if !zero && col.Zero == field.Interface() {
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
func buildDeleteSQL(sql *bytes.Buffer, e forward.Engine, v interface{}) ([]interface{}, error) {
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

	sql.WriteString("DELETE FROM ")
	e.Dialect().Quote(sql, e.Prefix()+m.Name)

	return where(e, sql, m, rval)

}

// 删除objs中指定的表名。
// 系统会默认给表名加上表名前缀。
// 若v为空，则不发生任何操作。
func buildDropSQL(sql *bytes.Buffer, e forward.Engine, v interface{}) error {
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
func buildTruncateSQL(sql *bytes.Buffer, e forward.Engine, v interface{}) error {
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

	sql := bytes.NewBufferString("SELECT COUNT(*) AS count FROM ")
	e.Dialect().Quote(sql, e.Prefix()+m.Name)
	vals, err := whereAny(e, sql, m, rval)
	if err != nil {
		return 0, err
	}

	rows, err := e.Query(false, sql.String(), vals...)
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

func create(e forward.Engine, v interface{}) error {
	sql := new(bytes.Buffer)
	if err := buildCreateSQL(sql, e, v); err != nil {
		return err
	}
	if _, err := e.Exec(false, sql.String()); err != nil {
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
		sql.Reset()
		sql.WriteString("CREATE INDEX ")
		e.Dialect().Quote(sql, name)
		sql.WriteString(" ON ")
		e.Dialect().Quote(sql, e.Prefix()+m.Name)
		sql.WriteByte('(')
		for _, col := range cols {
			e.Dialect().Quote(sql, col.Name)
			sql.WriteByte(',')
		}
		sql.Truncate(sql.Len() - 1)
		sql.WriteByte(')')
		if _, err := e.Exec(false, sql.String()); err != nil {
			return err
		}
	}
	return nil
}

func drop(e forward.Engine, v interface{}) error {
	sql := new(bytes.Buffer)
	if err := buildDropSQL(sql, e, v); err != nil {
		return err
	}

	_, err := e.Exec(false, sql.String())
	return err
}

func truncate(e forward.Engine, v interface{}) error {
	sql := new(bytes.Buffer)
	if err := buildTruncateSQL(sql, e, v); err != nil {
		return err
	}

	_, err := e.Exec(false, sql.String())
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

	sql := sqlbuilder.New(e).
		Insert("{#" + m.Name + "}").
		Keys(keys...).
		Values(vals...)

	if sql.HasError() {
		return nil, sql.Errors()
	}
	return sql.Exec(true)
}

func find(e forward.Engine, v interface{}) error {
	sql := new(bytes.Buffer)
	vals, err := buildSelectSQL(sql, e, v)
	if err != nil {
		return err
	}

	rows, err := e.Query(false, sql.String(), vals...)
	if err != nil {
		return err
	}

	_, err = fetch.Obj(v, rows)
	rows.Close()
	return err
}

// 更新v到数据库，zero表示是否将零值也更新到数据库。
func update(e forward.Engine, v interface{}, zero bool) (sql.Result, error) {
	sql := new(bytes.Buffer)

	vals, err := buildUpdateSQL(sql, e, v, zero)
	if err != nil {
		return nil, err
	}

	return e.Exec(false, sql.String(), vals...)
}

func del(e forward.Engine, v interface{}) (sql.Result, error) {
	sql := new(bytes.Buffer)
	vals, err := buildDeleteSQL(sql, e, v)
	if err != nil {
		return nil, err
	}

	return e.Exec(false, sql.String(), vals...)
}

// rval 为结构体指针组成的数据
func buildInsertManySQL(e forward.Engine, rval reflect.Value) (*sqlbuilder.SQLBuilder, error) {
	sql := sqlbuilder.New(e)
	vals := make([]interface{}, 0, 10)
	keys := []string{}
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
				cols = append(cols, "{"+name+"}") // 记录列的顺序
				keys = append(keys, name)         // 记录列的顺序
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
