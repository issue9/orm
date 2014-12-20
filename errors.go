// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package orm

import (
	"strconv"
)

type Errors []error

func (err Errors) Error() string {
	ret := "发生以下错误:"
	for index, msg := range err {
		ret += (strconv.Itoa(index) + ":" + msg.Error())
	}

	return ret
}
