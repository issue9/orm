// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package fetch

import (
	"database/sql"
	"fmt"
	"reflect"
	"unicode"

	"github.com/issue9/conv"
	t "github.com/issue9/orm/tags"
)

// 将v转换成map[string]reflect.Value形式，其中键名为对象的字段名，
// 键值为字段的值。支持匿名字段，不会转换不可导出(小写字母开头)的
// 字段，也不会转换 struct tag 以-开头的字段。
func parseObj(v reflect.Value, ret *map[string]reflect.Value) error {
	for v.Kind() == reflect.Ptr { // 嵌套类型可能为指针
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("parseObj:v参数的类型只能是reflect.Struct或是struct的指针,当前为[%v]", v.Kind())
	}

	vt := v.Type()
	num := vt.NumField()
	for i := 0; i < num; i++ {
		field := vt.Field(i)

		if field.Anonymous { // 匿名对象
			parseObj(v.Field(i), ret)
			continue
		}

		tags := field.Tag.Get("orm")
		if len(tags) > 0 { // 存在struct tag
			if tags[0] == '-' { // 该字段被标记为忽略
				continue
			}

			if name, found := t.Get(tags, "name"); found {
				if _, found := (*ret)[name[0]]; found {
					return fmt.Errorf("parseObj:已存在相同名字的字段[%v]", field.Name)
				}
				(*ret)[name[0]] = v.Field(i)
				continue
			}
		}

		// 未指定 struct tag，则尝试直接使用字段名。
		if unicode.IsUpper(rune(field.Name[0])) {
			if _, found := (*ret)[field.Name]; found {
				return fmt.Errorf("parseObj:已存在相同名字的字段[%v]", field.Name)
			}
			(*ret)[field.Name] = v.Field(i)
		}
	} // end for

	return nil
}

// 将rows中的一条记录写入到val中，必须保证val的类型为reflect.Struct。
// 仅供Obj()调用。
func fetchOnceObj(val reflect.Value, rows *sql.Rows) (int, error) {
	mapped, err := Map(true, rows)
	if err != nil {
		return 0, err
	}

	if len(mapped) == 0 { // 没有导出的数据
		return 0, nil
	}

	objItem := make(map[string]reflect.Value, len(mapped[0]))
	if err = parseObj(val, &objItem); err != nil {
		return 0, err
	}

	for index, item := range objItem {
		v, found := mapped[0][index]
		if !found {
			continue
		}
		if err = conv.Value(v, item); err != nil {
			return 0, err
		}
	}

	return 1, nil
}

// 将rows中的记录按obj的长度数量导出到obj中。
// val的类型必须是reflect.Slice或是reflect.Array.
// 可能只有部分数据被成功导入，而后发生error，
// 此时只能通过第一个返回参数来判断有多少数据是成功导入的。
func fetchObjToFixedSlice(val reflect.Value, rows *sql.Rows) (int, error) {
	itemType := val.Type().Elem()
	for itemType.Kind() == reflect.Ptr {
		itemType = itemType.Elem()
	}
	// 判断数组元素的类型是否为struct
	if itemType.Kind() != reflect.Struct {
		return 0, fmt.Errorf("fetchObjToFixedSlice:元素类型只能为reflect.Struct或是struct指针，当前为[%v]", itemType.Kind())
	}

	// 先导出数据到map中
	mapped, err := Map(false, rows)
	if err != nil {
		return 0, err
	}

	l := len(mapped)
	if l > val.Len() {
		l = val.Len()
	}

	for i := 0; i < l; i++ {
		objItem := make(map[string]reflect.Value, len(mapped[i]))
		if err = parseObj(val.Index(i), &objItem); err != nil {
			return 0, err
		}
		for index, item := range objItem {
			v, found := mapped[i][index]
			if !found {
				continue
			}
			if err = conv.Value(v, item); err != nil {
				return i, err // 已经有i条数据被正确导出
			}
		} // end for objItem
	}

	return l, nil
}

// 将rows中的所有记录导出到val中，val必须为slice的指针。
// 若val的长度不够，会根据rows中的长度调整。
// 可能只有部分数据被成功导入，而后发生error，
// 此时只能通过第一个返回参数来判断有多少数据是成功导入的。
func fetchObjToSlice(val reflect.Value, rows *sql.Rows) (int, error) {
	elem := val.Elem()

	itemType := elem.Type().Elem()
	for itemType.Kind() == reflect.Ptr {
		itemType = itemType.Elem()
	}
	// 判断数组元素的类型是否为struct
	if itemType.Kind() != reflect.Struct {
		return 0, fmt.Errorf("元素类型只能为reflect.Struct或是struct指针，当前为[%v]", itemType.Kind())
	}

	// 先导出数据到map中
	mapped, err := Map(false, rows)
	if err != nil {
		return 0, err
	}

	// 使elem表示的数组长度最起码和mapped一样。
	size := len(mapped) - elem.Len()
	if size > 0 {
		for i := 0; i < size; i++ {
			elem = reflect.Append(elem, reflect.New(itemType))
		}
		val.Elem().Set(elem)
	}

	for i := 0; i < len(mapped); i++ {
		objItem := make(map[string]reflect.Value, len(mapped[i]))
		if err = parseObj(elem.Index(i), &objItem); err != nil {
			return 0, err
		}

		for index, item := range objItem {
			e, found := mapped[i][index]
			if !found {
				continue
			}
			if err = conv.Value(e, item); err != nil {
				return i, err
			}
		} // end for objItem
	}

	return len(mapped), nil
}

// 将rows中的数据导出到obj中。obj只有在类型为slice指针时，
// 才有可能随着rows的长度变化，否则其长度是固定的，若查询结果为空值，
// 则不会对obj的内容做任何更改。
//
// obj参数可以为以下四种类型：
//
// struct指针：
// 将rows中的第一条记录转换成obj对象。
//
// struct array指针或是struct slice:
// 将rows中的len(obj)条记录导出到obj对象中；若rows中的数量不足，
// 则obj尾部的元素保存原来的值。
//
// struct slice指针：
// 将rows中的所有记录依次写入obj中。若rows中的记录比len(obj)要长，
// 则会增长obj的长度以适应rows的所有记录。
//
// struct可以在struct tag中用name指定字段名称，
// 或是以减号(-)开头表示忽略该字段的导出：
//  type user struct {
//      ID    int `orm:"name(id)"`  // 对应rows中的id字段，而不是ID。
//      age   int `orm:"name(Age)"` // 小写不会被导出。
//      Count int `orm:"-"`         // 不会匹配与该字段对应的列。
//  }
//
// 第一个参数用于表示有多少数据被正确导入到obj中
func Obj(obj interface{}, rows *sql.Rows) (int, error) {
	val := reflect.ValueOf(obj)

	switch val.Kind() {
	case reflect.Ptr:
		elem := val.Elem()
		switch elem.Kind() {
		case reflect.Slice: // slice指针，可以增长
			return fetchObjToSlice(val, rows)
		case reflect.Array: // 数组指针，只能按其大小导出
			return fetchObjToFixedSlice(elem, rows)
		case reflect.Struct: // 结构指针，只能导出一个
			return fetchOnceObj(elem, rows)
		default:
			return 0, fmt.Errorf("不允许的数据类型：[%v]", val.Kind())
		}
	case reflect.Slice: // slice只能按其大小导出。
		return fetchObjToFixedSlice(val, rows)
	default:
		return 0, fmt.Errorf("不允许的数据类型：[%v]", val.Kind())
	}
}
