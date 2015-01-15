// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package core

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var expr = regexp.MustCompile(`@\w+`)

// 提取sql语句中的参数。将其中的`@xx`替换成`?`，并将这类名称按顺序提取出来。
func ExtractArgs(sql string) (string, []string) {
	if strings.IndexByte(sql, '@') == -1 {
		return sql, nil
	}

	args := []string{}
	sql = expr.ReplaceAllStringFunc(sql, func(src string) string {
		args = append(args, src[1:])
		return "?"
	})
	return sql, args
}

// 将args中的数据项转换成[]interface{}，其顺序为argNames中的顺序。
func ConvArgs(argNames []string, args map[string]interface{}) ([]interface{}, error) {
	if len(argNames) != len(args) {
		return nil, fmt.Errorf("PrepareArgs:参数长度不一样:len(argName)=%v; len(args)=%v", len(argNames), len(args))
	}

	ret := make([]interface{}, len(argNames))
	for index, name := range argNames {
		ret[index] = args[name]
	}
	return ret, nil
}

// 将一个内置的数据类型转换成SQL表示的值。字符串为加上单引号
func AsSQLValue(src interface{}) string {
	switch v := src.(type) {
	case string:
		if v[0] == '@' {
			return v
		}
		return "'" + v + "'"
	case []byte:
		if v[0] == '@' {
			return string(v)
		}
		return "'" + string(v) + "'"
	case []rune:
		if v[0] == '@' {
			return string(v)
		}
		return "'" + string(v) + "'"
	}
	rv := reflect.ValueOf(src)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 32)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	}
	return fmt.Sprintf("%v", src)
}
