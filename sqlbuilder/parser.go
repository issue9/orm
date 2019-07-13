// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlbuilder

import (
	"strings"
	"unicode"
)

func splitWithAS(col string) (column, alias string) {
	var state byte
	for index, c := range col {
		switch {
		case unicode.IsSpace(c):
			if state == 's' {
				alias = strings.TrimSpace(col[index+1:])
				return
			}
			state = ' '
			column = col[:index]
		case c == 'a' || c == 'A':
			if state == ' ' {
				state = 'a'
			}
		case c == 's' || c == 'S':
			if state == 'a' {
				state = 's'
			}
		}
	}
	return col, ""
}