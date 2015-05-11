// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
)

// 装饰src转换成sql的值，并写入到w中。
func asString(w *bytes.Buffer, src interface{}) {
	switch v := src.(type) {
	case string:
		w.WriteString(v)
	case []byte:
		w.Write(v)
	case []rune:
		w.WriteString(string(v))
	}

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
	}
	fmt.Fprint(w, src)
}
