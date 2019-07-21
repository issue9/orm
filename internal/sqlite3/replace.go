// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sqlite3

import "strings"

var quoteReplacer = strings.NewReplacer("`", "")

func filterCreateTableSQL(sql string) []string {
	sql = quoteReplacer.Replace(sql)
	var deep, start int
	var lines []string

	for index, c := range sql {
		switch c {
		case ',':
			if deep == 1 && index > start {
				lines = append(lines, strings.TrimSpace(sql[start:index]))
				start = index + 1 // 不包含 ( 本身
			}
		case '(':
			deep++
			if deep == 1 {
				start = index + 1 // 不包含 ( 本身
			}
		case ')':
			deep--
			if deep == 0 { // 不需要 create table xx() 之后的内容
				if start != index {
					lines = append(lines, strings.TrimSpace(sql[start:index]))
				}
				break
			}
		} // end switch
	} // end for

	return lines
}
