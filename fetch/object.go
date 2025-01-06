// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package fetch

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/issue9/conv"

	"github.com/issue9/orm/v6/core"
	t "github.com/issue9/orm/v6/internal/tags"
)

// AfterFetcher 在数据从数据库拉取之后执行的操作
type AfterFetcher interface {
	AfterFetch() error
}

var errUnsupportedKind = errors.New("不支持的 Kind 类型")

// ErrUnsupportedKind 表示当前功能对数据的 Kind 值有特殊需求
func ErrUnsupportedKind() error { return errUnsupportedKind }

// 输出无法转换时的字段信息
func convertError(field string, message error) error {
	return fmt.Errorf("字段 %s 转换出错：%w", field, message)
}

// Object 将 rows 中的数据导出到 obj 中
//
// strict 严格模式，如果为 true，不会将 null 转换成当前类型的空值。
// 在多表查询的情况下，如果有空记录，则 strict 模式下，可能会出错（nil 无法转换)。
//
// obj 只有在类型为 slice 指针时，才有可能随着 rows 的长度变化，
// 否则其长度是固定的，若查询结果为空值，则不会对 obj 的内容做任何更改。
// 可以为以下几种类型：
//
// struct 指针：
// 将 rows 中的第一条记录转换成 obj 对象。
//
// struct array 指针或是 struct slice:
// 将 rows 中的 len(obj) 条记录导出到 obj 对象中；若 rows 中的数量不足，
// 则 obj 尾部的元素保存原来的值。
//
// struct slice 指针：
// 将 rows 中的所有记录依次写入 obj 中。若 rows 中的记录比 len(obj) 要长，
// 则会增长 obj 的长度以适应 rows 的所有记录。
//
// struct 可以在 struct tag 中用 name 指定字段名称，
// 或是以减号(-)开头表示忽略该字段的导出：
//
//	type user struct {
//	    ID    int `orm:"name(id)"`  // 对应 rows 中的 id 字段，而不是 ID。
//	    age   int `orm:"name(Age)"` // 小写不会被导出。
//	    Count int `orm:"-"`         // 不会匹配与该字段对应的列。
//	}
//
// 第一个返回参数用于表示有多少数据被正确导入到 obj 中
func Object(strict bool, rows *sql.Rows, obj any) (int, error) {
	val := reflect.ValueOf(obj)

	switch val.Kind() {
	case reflect.Ptr:
		elem := val.Elem()
		switch elem.Kind() {
		case reflect.Slice: // slice 指针，可以增长
			return fetchObjToSlice(strict, val, rows)
		case reflect.Array: // 数组指针，只能按其大小导出
			return fetchObjToFixedSlice(strict, elem, rows)
		case reflect.Struct: // 结构指针，只能导出一个
			return fetchOnceObj(strict, elem, rows)
		default:
			return 0, ErrUnsupportedKind()
		}
	case reflect.Slice: // slice 只能按其大小导出。
		return fetchObjToFixedSlice(strict, val, rows)
	default:
		return 0, ErrUnsupportedKind()
	}
}

// 将 v 转换成 map[string]reflect.Value 形式，其中键名为对象的字段名，
// 键值为字段的值。支持匿名字段，不会转换不可导出(非大写字母开头)的
// 字段，也不会转换 struct tag 以 - 开头的字段。
func parseObject(v reflect.Value, ret *map[string]reflect.Value) error {
	v = getRealValue(v)
	if v.Kind() != reflect.Struct {
		return ErrUnsupportedKind()
	}

	vt := v.Type()
	num := vt.NumField()
	for i := 0; i < num; i++ {
		field := vt.Field(i)
		vf := v.Field(i)

		if field.Anonymous {
			if err := parseObject(vf, ret); err != nil {
				return err
			}
			continue
		}

		name := getName(field)
		if name == "" {
			continue
		}

		vf2 := getRealValue(vf)

		if core.GetPrimitiveType(vf.Type()) == core.Auto {
			items := make(map[string]reflect.Value, vf2.NumField())
			if err := parseObject(vf2, &items); err != nil {
				return err
			}

			for subName, val := range items {
				(*ret)[name+"."+subName] = val
			}
		} else if _, found := (*ret)[name]; found {
			return fmt.Errorf("已存在相同名字的字段 %s", name)
		} else {
			(*ret)[name] = vf2
		}
	} // end for

	return nil
}

func getRealValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		} else {
			v = v.Elem()
		}
	}
	return v
}

func getName(field reflect.StructField) string {
	tags := field.Tag.Get(Tag)
	if len(tags) > 0 { // 存在 struct tag
		if tags[0] == '-' { // 该字段被标记为忽略
			return ""
		}

		if name, found := t.Get(tags, "name"); found {
			return name[0]
		}
	}

	if field.IsExported() { // 未指定 struct tag，则尝试直接使用字段名
		return field.Name
	}

	return ""
}

func getColumns(v reflect.Value, cols []string) ([]any, error) {
	ret := make([]any, 0, len(cols))

	items := make(map[string]reflect.Value, len(cols))
	if err := parseObject(v, &items); err != nil {
		return nil, err
	}

	for _, col := range cols {
		if item, found := items[col]; found {
			ret = append(ret, item.Addr().Interface())
		} else { // 从数据库导出了该列，但是该列名不存在于模型中
			var val any
			ret = append(ret, &val)
		}
	}

	return ret, nil
}

// 将 rows 中的一条记录写入到 val 中，必须保证 val 的类型为 reflect.Struct。
// 仅供 Object() 调用。
func fetchOnceObj(strict bool, val reflect.Value, rows *sql.Rows) (int, error) {
	if !strict {
		return fetchOnceObjNoStrict(val, rows)
	}

	cols, err := rows.Columns()
	if err != nil {
		return 0, err
	}

	buff, err := getColumns(val, cols)
	if err != nil {
		return 0, err
	}

	if !rows.Next() {
		return 0, rows.Err()
	}
	if err := rows.Scan(buff...); err != nil {
		return 0, err
	}

	if err = afterFetch(val); err != nil {
		return 0, err
	}
	return 1, nil
}

func fetchOnceObjNoStrict(val reflect.Value, rows *sql.Rows) (int, error) {
	mapped, err := Map(true, rows)
	if err != nil {
		return 0, err
	}

	if len(mapped) == 0 { // 没有导出的数据
		return 0, nil
	}

	objItem := make(map[string]reflect.Value, len(mapped[0]))
	if err = parseObject(val, &objItem); err != nil {
		return 0, err
	}

	for index, item := range objItem {
		v, found := mapped[0][index]
		if !found {
			continue
		}
		if err = conv.Value(v, item); err != nil {
			return 0, convertError(index, err)
		}
	}

	if err = afterFetch(val); err != nil {
		return 0, err
	}

	return 1, nil
}

// 将 rows 中的记录按 obj 的长度数量导出到 obj 中。
// val 的类型必须是 reflect.Slice 或是 reflect.Array.
// 可能只有部分数据被成功导入，而后发生 error，
// 此时只能通过第一个返回参数来判断有多少数据是成功导入的。
func fetchObjToFixedSlice(strict bool, val reflect.Value, rows *sql.Rows) (int, error) {
	if !strict {
		return fetchObjToFixedSliceNoStrict(val, rows)
	}

	itemType := val.Type().Elem()
	for itemType.Kind() == reflect.Ptr {
		itemType = itemType.Elem()
	}
	if itemType.Kind() != reflect.Struct {
		return 0, ErrUnsupportedKind()
	}

	cols, err := rows.Columns()
	if err != nil {
		return 0, err
	}

	l := val.Len()
	for i := 0; (i < l) && rows.Next(); i++ {
		buff, err := getColumns(val.Index(i), cols)
		if err != nil {
			return 0, err
		}
		if err := rows.Scan(buff...); err != nil {
			return 0, err
		}

		if err = afterFetch(val.Index(i)); err != nil {
			return 0, err
		}
	}

	return l, nil
}

func fetchObjToFixedSliceNoStrict(val reflect.Value, rows *sql.Rows) (int, error) {
	itemType := val.Type().Elem()
	for itemType.Kind() == reflect.Ptr {
		itemType = itemType.Elem()
	}
	if itemType.Kind() != reflect.Struct {
		return 0, ErrUnsupportedKind()
	}

	// 先导出数据到 map 中
	mapped, err := Map(false, rows)
	if err != nil {
		return 0, err
	}

	l := min(len(mapped), val.Len())
	for i := 0; i < l; i++ {
		objItem := make(map[string]reflect.Value, len(mapped[i]))
		if err = parseObject(val.Index(i), &objItem); err != nil {
			return 0, err
		}
		for index, item := range objItem {
			v, found := mapped[i][index]
			if !found {
				continue
			}
			if err = conv.Value(v, item); err != nil {
				return i, convertError(index, err) // 已经有 i 条数据被正确导出
			}
		} // end for objItem

		if err = afterFetch(val.Index(i)); err != nil {
			return 0, err
		}
	}

	return l, nil
}

// 将 rows 中的所有记录导出到 val 中，val 必须为 slice 的指针。
// 若 val 的长度不够，会根据 rows 中的长度调整。
//
// 可能只有部分数据被成功导入，而后发生 error，
// 此时只能通过第一个返回参数来判断有多少数据是成功导入的。
func fetchObjToSlice(strict bool, val reflect.Value, rows *sql.Rows) (int, error) {
	if !strict {
		return fetchObjToSliceNoStrict(val, rows)
	}

	elem := val.Elem()

	itemType := elem.Type().Elem()
	for itemType.Kind() == reflect.Ptr {
		itemType = itemType.Elem()
	}
	if itemType.Kind() != reflect.Struct {
		return 0, ErrUnsupportedKind()
	}

	cols, err := rows.Columns()
	if err != nil {
		return 0, err
	}

	l := elem.Len()
	count := 0
	for i := 0; rows.Next(); i++ {
		count++
		if i >= l {
			elem = reflect.Append(elem, reflect.New(itemType))
			val.Elem().Set(elem)
		}

		buff, err := getColumns(elem.Index(i), cols)
		if err != nil {
			return 0, err
		}

		if err := rows.Scan(buff...); err != nil {
			return 0, err
		}

		if err = afterFetch(elem.Index(i)); err != nil {
			return 0, err
		}
	}

	return count, nil
}

func fetchObjToSliceNoStrict(val reflect.Value, rows *sql.Rows) (int, error) {
	elem := val.Elem()

	itemType := elem.Type().Elem()
	for itemType.Kind() == reflect.Ptr {
		itemType = itemType.Elem()
	}
	if itemType.Kind() != reflect.Struct {
		return 0, ErrUnsupportedKind()
	}

	// 先导出数据到 map 中
	mapped, err := Map(false, rows)
	if err != nil {
		return 0, err
	}

	// 使 elem 表示的数组长度最起码和 mapped 一样。
	size := len(mapped) - elem.Len()
	if size > 0 {
		for i := 0; i < size; i++ {
			elem = reflect.Append(elem, reflect.New(itemType))
		}
		val.Elem().Set(elem)
	}

	for i := 0; i < len(mapped); i++ {
		objItem := make(map[string]reflect.Value, len(mapped[i]))
		if err = parseObject(elem.Index(i), &objItem); err != nil {
			return 0, err
		}

		for index, item := range objItem {
			e, found := mapped[i][index]
			if !found {
				continue
			}
			if err = conv.Value(e, item); err != nil {
				return i, convertError(index, err)
			}
		} // end for objItem

		if err = afterFetch(elem.Index(i)); err != nil {
			return 0, err
		}
	}

	return len(mapped), nil
}

func afterFetch(v reflect.Value) error {
	if f, ok := v.Interface().(AfterFetcher); ok {
		return f.AfterFetch()
	}
	return nil
}
