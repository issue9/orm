// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

type Errors []error

func (e Errors) Error() string {
	msg := "发生以下错误："
	for _, err := range e {
		msg += err.Error() + "\n"
	}

	return msg
}
